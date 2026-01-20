package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

//
// --------------------
// Globals
// --------------------
//
var (
	ctx = context.Background()
	rdb *redis.Client
)

//
// --------------------
// Helpers
// --------------------
//
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func containsIgnoreCase(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}

//
// --------------------
// Data Structures
// --------------------
//
type RankedUser struct {
	Username string `json:"username"`
	Rating   int    `json:"rating"`
	Rank     int    `json:"rank"`
}

//
// --------------------
// Seed users into Redis
// --------------------
//
func seedUsers() {
	first := []string{"rahul", "sai", "arjun", "kiran", "rohit"}
	last := []string{"kumar", "reddy", "sharma", "verma", "patel"}

	pipe := rdb.Pipeline()

	for i := 1; i <= 10000; i++ {
		username := fmt.Sprintf(
			"%s_%s_%d",
			first[rand.Intn(len(first))],
			last[rand.Intn(len(last))],
			i,
		)

		rating := rand.Intn(4800) + 100

		pipe.ZAdd(ctx, "leaderboard", redis.Z{
			Score:  float64(rating),
			Member: username,
		})

		if i%1000 == 0 {
			_, _ = pipe.Exec(ctx)
		}
	}
	_, _ = pipe.Exec(ctx)

	// Explicit top users
	rdb.ZAdd(ctx, "leaderboard", redis.Z{Score: 5000, Member: "Legendary_Player_1"})
	rdb.ZAdd(ctx, "leaderboard", redis.Z{Score: 4998, Member: "Master_Gamer_2"})

	for i := 1; i <= 8; i++ {
		rdb.ZAdd(ctx, "leaderboard", redis.Z{
			Score:  float64(4990 + rand.Intn(8)),
			Member: fmt.Sprintf("Pro_Player_%d", i),
		})
	}

	fmt.Println("Seeded 10,000 users")
}

//
// --------------------
// Leaderboard Logic
// --------------------
//
func getLeaderboard(page, limit int, query string) ([]RankedUser, int64) {
	// SEARCH
	if query != "" {
		var results []RankedUser
		cursor := uint64(0)
		pattern := "*" + query + "*"

		for {
			items, next, err := rdb.ZScan(ctx, "leaderboard", cursor, pattern, 5000).Result()
			if err != nil {
				break
			}
			cursor = next

			for i := 0; i < len(items); i += 2 {
				member := items[i]
				score, _ := strconv.ParseFloat(items[i+1], 64)
				rank, _ := rdb.ZRevRank(ctx, "leaderboard", member).Result()

				results = append(results, RankedUser{
					Username: member,
					Rating:   int(score),
					Rank:     int(rank) + 1,
				})
			}
			if cursor == 0 {
				break
			}
		}

		sort.Slice(results, func(i, j int) bool {
			return results[i].Rank < results[j].Rank
		})

		total := int64(len(results))
		start := (page - 1) * limit
		end := start + limit
		if start >= len(results) {
			return []RankedUser{}, total
		}
		if end > len(results) {
			end = len(results)
		}

		return results[start:end], total
	}

	// STANDARD LEADERBOARD
	total, _ := rdb.ZCard(ctx, "leaderboard").Result()
	start := int64((page - 1) * limit)
	stop := start + int64(limit) - 1

	zs, _ := rdb.ZRevRangeWithScores(ctx, "leaderboard", start, stop).Result()

	var leaderboard []RankedUser
	prevScore := -1.0
	rank := int(start)
	actual := int(start)

	for _, z := range zs {
		actual++
		if z.Score != prevScore {
			rank = actual
			prevScore = z.Score
		}

		leaderboard = append(leaderboard, RankedUser{
			Username: z.Member.(string),
			Rating:   int(z.Score),
			Rank:     rank,
		})
	}

	return leaderboard, total
}

//
// --------------------
// Random Score Updates
// --------------------
//
func startRandomScoreUpdates() {
	go func() {
		for {
			time.Sleep(6 * time.Second)

			users, err := rdb.ZRevRange(ctx, "leaderboard", 0, 49).Result()
			if err != nil || len(users) == 0 {
				continue
			}

			for i := 0; i < rand.Intn(6)+5; i++ {
				u := users[rand.Intn(len(users))]
				score, err := rdb.ZScore(ctx, "leaderboard", u).Result()
				if err != nil {
					continue
				}

				delta := rand.Intn(150) - 50
				if score >= 4950 && rand.Float64() < 0.8 {
					delta = -rand.Intn(50)
				}

				newScore := score + float64(delta)
				if newScore < 100 {
					newScore = 100
				}
				if newScore > 5000 {
					newScore = 5000
				}

				rdb.ZAdd(ctx, "leaderboard", redis.Z{
					Score:  newScore,
					Member: u,
				})
			}
		}
	}()
}

//
// --------------------
// MAIN
// --------------------
//
func main() {
	rand.Seed(time.Now().UnixNano())

	// Redis connection
	redisURL := os.Getenv("REDIS_URL")
	if redisURL != "" {
		opt, _ := redis.ParseURL(redisURL)
		rdb = redis.NewClient(opt)
	} else {
		rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	}

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		panic("Redis not connected")
	}

	rdb.FlushDB(ctx)
	seedUsers()
	startRandomScoreUpdates()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		fmt.Fprintln(w, "Backend is running")
	})

	http.HandleFunc("/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		query := r.URL.Query().Get("q")

		if page < 1 {
			page = 1
		}
		if limit < 1 {
			limit = 50
		}

		users, total := getLeaderboard(page, limit, query)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"users": users,
			"total": total,
		})
	})

	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		username := r.URL.Query().Get("username")
		rating, _ := strconv.Atoi(r.URL.Query().Get("rating"))

		if rating < 100 || rating > 5000 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		rdb.ZAdd(ctx, "leaderboard", redis.Z{
			Score:  float64(rating),
			Member: username,
		})

		json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
	})

	// ✅ Render-compatible PORT
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server running on port", port)
	http.ListenAndServe(":"+port, nil)
}
