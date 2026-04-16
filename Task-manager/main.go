package main

import (
	"database/sql"
	"net/http"

	_ "modernc.org/sqlite"

	"github.com/aadityya4real/Task-manager/internal/handler"
	"github.com/aadityya4real/Task-manager/internal/storage"
	"github.com/redis/go-redis/v9"
)

func main() {

	// 🔹 Connect SQLite DB
	db, err := sql.Open("sqlite", "tasks.db")
	if err != nil {
		panic(err)
	}

	// 🔹 Create table if not exists
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT,
		done BOOLEAN
	);
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT,
		password TEXT
	);
	`)
	if err != nil {
		panic(err)
	}

	// 🔹 Initialize Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 🔹 Create Store
	store := storage.New(db)

	// 🔹 Routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Task Manager API Running 🚀"))
	})

	//Task routes
	http.HandleFunc("/task", handler.TaskHandler(store, rdb))

	// Auth routes
	http.HandleFunc("/signup", handler.AuthHandler(store))
	http.HandleFunc("/login", handler.AuthHandler(store))

	// 🔹 Start server
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
