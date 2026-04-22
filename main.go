package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	_ "modernc.org/sqlite"

	"github.com/aadityya4real/Task-manager/internal/handler"
	"github.com/aadityya4real/Task-manager/internal/middleware"
	"github.com/aadityya4real/Task-manager/internal/storage"
	"github.com/redis/go-redis/v9"
)

// ✅ CORS FIX (important for frontend)
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// ✅ handle preflight properly
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {

	fmt.Println("🚀 SERVER STARTED")

	// 🔹 DB connection
	db, err := sql.Open("sqlite", "tasks.db")
	if err != nil {
		panic(err)
	}
	fmt.Println("✅ DB connected")

	// 🔹 Create tables
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT,
		password TEXT
	)
	`)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT,
		done BOOLEAN,
		user_id INTEGER
	)
	`)
	if err != nil {
		panic(err)
	}

	// 🔹 Redis
	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(opt)
	// rdb := redis.NewClient(&redis.Options{
	// 	Addr: "localhost:6379",
	// })
	fmt.Println("✅ Redis initialized")

	// 🔹 Store
	store := storage.New(db)

	// 🔥 ROUTER (VERY IMPORTANT)
	mux := http.NewServeMux()

	fmt.Println("📌 Registering routes...")

	mux.HandleFunc("/signup", handler.SignupHandler(store))
	mux.HandleFunc("/login", handler.LoginHandler(store))
	mux.HandleFunc("/tasks", middleware.AuthMiddleware(handler.TaskHandler(store, rdb)))

	fmt.Println("🌍 Server running on http://localhost:8080")

	// 🔥 START SERVER
	err = http.ListenAndServe(":8080", enableCORS(mux))
	if err != nil {
		panic(err)
	}
}
