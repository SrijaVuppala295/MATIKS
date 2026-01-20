package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Helper for case-insensitive search
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// --------------------
// Data Structures
// --------------------
type User struct {
	Username string `json:"username"`
	Rating   int    `json:"rating"`
}

type RankedUser struct {
	Username string `json:"username"`
	Rating   int    `json:"rating"`
	Rank     int    `json:"rank"`
}

// --------------------
// In-memory storage
// --------------------
var (
	users = make(map[string]User)
	mu    sync.RWMutex
)

// --------------------
// Enable CORS
// --------------------
func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// --------------------
// Seed 10,000 users
// --------------------
func seedUsers() {
	firstNames := []string{
		"rahul", "sai", "arjun", "kiran", "rohit",
		"aman", "vikram", "priya", "neha", "anita",
		"pooja", "ravi", "suresh", "manoj", "vishal",
	}
	lastNames := []string{
		"kumar", "reddy", "sharma", "verma", "patel",
		"singh", "naidu", "gupta", "mehta", "iyer",
	}

	for i := 1; i <= 10000; i++ {
		first := firstNames[rand.Intn(len(firstNames))]
		last := lastNames[rand.Intn(len(lastNames))]
		username := fmt.Sprintf("%s_%s_%d", first, last, i)
		rating := rand.Intn(4901) + 100 // 100–5000

		users[username] = User{
			Username: username,
			Rating:   rating,
		}
	}

	fmt.Println("Seeded users:", len(users))
}

// --------------------
// Build leaderboard (tie-aware)
// --------------------
func buildLeaderboard() []RankedUser {
	mu.RLock()
	defer mu.RUnlock()

	ratingGroups := make(map[int][]User)
	for _, user := range users {
		ratingGroups[user.Rating] = append(ratingGroups[user.Rating], user)
	}

	var ratings []int
	for rating := range ratingGroups {
		ratings = append(ratings, rating)
	}

	sort.Sort(sort.Reverse(sort.IntSlice(ratings)))

	var leaderboard []RankedUser
	currentRank := 1

	for _, rating := range ratings {
		group := ratingGroups[rating]

		for _, user := range group {
			leaderboard = append(leaderboard, RankedUser{
				Username: user.Username,
				Rating:   user.Rating,
				Rank:     currentRank,
			})
		}

		currentRank += len(group) // skip ranks
	}

	return leaderboard
}

// --------------------
// Simulate random score updates (1–2 users every 5–10 sec)
// --------------------
func startRandomScoreUpdates() {
	go func() {
		for {
			// wait 5–10 seconds
			time.Sleep(time.Duration(5+rand.Intn(6)) * time.Second)

			mu.Lock()

			updates := rand.Intn(2) + 1 // update 1 or 2 users
			for username, user := range users {
				// small rating change (delta)
				delta := rand.Intn(201) - 100 // -100 to +100
				user.Rating += delta

				// keep rating in bounds
				if user.Rating < 100 {
					user.Rating = 100
				}
				if user.Rating > 5000 {
					user.Rating = 5000
				}

				users[username] = user
				updates--

				if updates == 0 {
					break
				}
			}

			mu.Unlock()
		}
	}()
}




// --------------------
// Main
// --------------------
func main() {
	rand.Seed(time.Now().UnixNano())
    seedUsers()
startRandomScoreUpdates()


	// Health check
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		fmt.Fprintln(w, "Backend is running")
	})

	// Total users count
	http.HandleFunc("/users/count", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		json.NewEncoder(w).Encode(map[string]int{
			"total_users": len(users),
		})
	})

	// Leaderboard (Paginated + Search)
	http.HandleFunc("/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		// Parse query params
		q := r.URL.Query().Get("q")
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit < 1 {
			limit = 50 // default
		}

		fullLeaderboard := buildLeaderboard()
		var filtered []RankedUser

		// Global Search / Filter
		if q != "" {
			// Simple case-insensitive contains
			// For 10k users, this linear scan is acceptable for a demo.
			// Ideally, use a trie or database.
			for _, u := range fullLeaderboard {
				if containsIgnoreCase(u.Username, q) {
					filtered = append(filtered, u)
				}
			}
		} else {
			filtered = fullLeaderboard
		}

		// Pagination
		total := len(filtered)
		start := (page - 1) * limit
		if start > total {
			start = total
		}
		end := start + limit
		if end > total {
			end = total
		}

		pagedUsers := filtered[start:end]

		// Response
		response := map[string]interface{}{
			"total": total,
			"page":  page,
			"users": pagedUsers,
		}

		json.NewEncoder(w).Encode(response)
	})

	// Search user (global rank)
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		username := r.URL.Query().Get("username")

		mu.RLock()
		_, exists := users[username]
		mu.RUnlock()

		if !exists {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "user not found",
			})
			return
		}

		leaderboard := buildLeaderboard()
		for _, ranked := range leaderboard {
			if ranked.Username == username {
				json.NewEncoder(w).Encode(ranked)
				return
			}
		}
	})

	// Update user rating
	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)

		if r.Method == http.MethodOptions {
			return
		}

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		username := r.URL.Query().Get("username")
		ratingStr := r.URL.Query().Get("rating")

		newRating, err := strconv.Atoi(ratingStr)
		if err != nil || newRating < 100 || newRating > 5000 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "rating must be between 100 and 5000",
			})
			return
		}

		mu.Lock()
		user, exists := users[username]
		if !exists {
			mu.Unlock()
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "user not found",
			})
			return
		}

		user.Rating = newRating
		users[username] = user
		mu.Unlock()

		json.NewEncoder(w).Encode(map[string]string{
			"status": "rating updated",
		})
	})

	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
