package main
import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"sort"
)


// --------------------
// Globals
// --------------------
var (
	ctx = context.Background()
	rdb *redis.Client
)

// --------------------
// Helpers
// --------------------
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func containsIgnoreCase(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}

// --------------------
// Data Structures
// --------------------
type RankedUser struct {
	Username string `json:"username"`
	Rating   int    `json:"rating"`
	Rank     int    `json:"rank"`
}

// --------------------
// Seed users into Redis
// --------------------
func seedUsers() {
	first := []string{"rahul", "sai", "arjun", "kiran", "rohit"}
	last := []string{"kumar", "reddy", "sharma", "verma", "patel"}

	// 1. Bulk seed normal users (Scores 100 - 4899)
	pipe := rdb.Pipeline()
	for i := 1; i <= 10000; i++ {
		username := fmt.Sprintf(
			"%s_%s_%d",
			first[rand.Intn(len(first))],
			last[rand.Intn(len(last))],
			i,
		)
		// Most users get lower scores to make 5000 rare
		rating := rand.Intn(4800) + 100

		pipe.ZAdd(ctx, "leaderboard", redis.Z{
			Score:  float64(rating),
			Member: username,
		})

		// Execute in batches of 1000 to be faster
		if i%1000 == 0 {
			_, err := pipe.Exec(ctx)
			if err != nil {
				fmt.Println("Error seeding batch:", err)
			}
		}
	}
	_, _ = pipe.Exec(ctx)

	// 2. Explicitly add 1-2 TOP players (The only ones with ~5000)
	rdb.ZAdd(ctx, "leaderboard", redis.Z{Score: 5000, Member: "Legendary_Player_1"})
	rdb.ZAdd(ctx, "leaderboard", redis.Z{Score: 4998, Member: "Master_Gamer_2"})

	// 3. Add crowding at the top (competitors) so updates are visible
	for i := 1; i <= 8; i++ {
		rdb.ZAdd(ctx, "leaderboard", redis.Z{
			Score:  float64(4990 + rand.Intn(8)), // 4990 - 4997
			Member: fmt.Sprintf("Pro_Player_%d", i),
		})
	}

	fmt.Println("Seeded 10,000 users. Logic: Rare 5000s + Top Crowding.")
}

// --------------------
// Leaderboard from Redis
// --------------------
func getLeaderboard(page, limit int, query string) ([]RankedUser, int64) {
	// ---------------------------
	// 1. SEARCH LOGIC
	// ---------------------------
	if query != "" {
		// Use ZSCAN to find keys matching the query
		var matches []RankedUser
		cursor := uint64(0)
		pattern := "*" + query + "*"

		for {
			keys, cursorVal, err := rdb.ZScan(ctx, "leaderboard", cursor, pattern, 10000).Result()
			if err != nil {
				break
			}
			cursor = cursorVal

			// ZScan returns [member, score, member, score...]
			// keys[i] is member, keys[i+1] is score (as string)
			for i := 0; i < len(keys); i += 2 {
				member := keys[i]
				scoreStr := keys[i+1]
				score, _ := strconv.ParseFloat(scoreStr, 64)

				// Get actual rank (expensive but needed for correct rank display)
				rank, _ := rdb.ZRevRank(ctx, "leaderboard", member).Result()

				matches = append(matches, RankedUser{
					Username: member,
					Rating:   int(score),
					Rank:     int(rank) + 1, // 0-indexed to 1-indexed
				})
			}

			if cursor == 0 {
				break
			}
		}

		// Sort matches by Rank (High Score first)
		sort.Slice(matches, func(i, j int) bool {
			return matches[i].Rank < matches[j].Rank
		})

		// Pagination for search results
		total := int64(len(matches))
		start := (page - 1) * limit
		if start >= len(matches) {
			return []RankedUser{}, total
		}
		end := start + limit
		if end > len(matches) {
			end = len(matches)
		}

		return matches[start:end], total
	}

	// ---------------------------
	// 2. STANDARD LIST LOGIC
	// ---------------------------
	
	// Get total count
	total, err := rdb.ZCard(ctx, "leaderboard").Result()
	if err != nil {
		return []RankedUser{}, 0
	}

	// Calculate range
	start := int64((page - 1) * limit)
	stop := start + int64(limit) - 1

	results, err := rdb.ZRevRangeWithScores(ctx, "leaderboard", start, stop).Result()
	if err != nil {
		return []RankedUser{}, total
	}

	var leaderboard []RankedUser
	prevScore := -1.0
	rank := int(start) // Rank starts from offset
	actual := int(start)

	for _, z := range results {
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


// --------------------
// Random score simulation
// --------------------
func startRandomScoreUpdates() {
	go func() {
		for {
			// ⚡ UPDATE: Frequency 6 seconds (User Request)
			time.Sleep(6 * time.Second)

			// 1. Target TOP 50 users (Visible on main screen)
			topUsers, err := rdb.ZRevRange(ctx, "leaderboard", 0, 49).Result()
			if err != nil {
				topUsers = []string{}
			}

			// 2. Target RANDOM 50 users (So searched users also update)
			var randomUsers []string
			total, err := rdb.ZCard(ctx, "leaderboard").Result()
			if err == nil && total > 0 {
				start := int64(rand.Intn(int(total)))
				randomUsers, _ = rdb.ZRange(ctx, "leaderboard", start, start+49).Result()
			}

			// Combine unique users
			userMap := make(map[string]bool)
			var users []string

			for _, u := range topUsers {
				if !userMap[u] {
					userMap[u] = true
					users = append(users, u)
				}
			}
			for _, u := range randomUsers {
				if !userMap[u] {
					userMap[u] = true
					users = append(users, u)
				}
			}

			if len(users) == 0 {
				continue
			}

			// Update 5-10 users at a time
			updates := rand.Intn(6) + 5 
			for i := 0; i < updates; i++ {
				u := users[rand.Intn(len(users))]

				score, err := rdb.ZScore(ctx, "leaderboard", u).Result()
				if err != nil {
					continue
				}

				// Logic to keep 5000 rare:
				delta := 0
				if score >= 4950 {
					if rand.Float64() < 0.8 {
						delta = -(rand.Intn(50) + 1)
					} else {
						delta = rand.Intn(10) + 1
					}
				} else {
					delta = rand.Intn(150) - 50 // Bias towards gain
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


// --------------------
// MAIN
// --------------------
func main() {
	rand.Seed(time.Now().UnixNano())

	// 🔌 Redis connection (MUST be inside main)
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		panic("Redis not connected")
	}
	fmt.Println("Connected to Redis")

	// Flush old data for tuning
	rdb.FlushDB(ctx)
	
	seedUsers()
	startRandomScoreUpdates()

	// --------------------
	// Routes
	// --------------------

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		fmt.Fprintln(w, "Backend is running")
	})

	http.HandleFunc("/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)

		pageStr := r.URL.Query().Get("page")
		limitStr := r.URL.Query().Get("limit")
		query := r.URL.Query().Get("q")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			limit = 50
		}

		users, total := getLeaderboard(page, limit, query)

		response := map[string]interface{}{
			"users": users,
			"total": total,
		}

		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		username := r.URL.Query().Get("username")
		ratingStr := r.URL.Query().Get("rating")

		rating, err := strconv.Atoi(ratingStr)
		if err != nil || rating < 100 || rating > 5000 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		rdb.ZAdd(ctx, "leaderboard", redis.Z{
			Score:  float64(rating),
			Member: username,
		})

		json.NewEncoder(w).Encode(map[string]string{
			"status": "rating updated",
		})
	})

	fmt.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
