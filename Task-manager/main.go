package main

import (
	"database/sql"
	"net/http"

	_ "modernc.org/sqlite"

	"github.com/aadityya4real/Task-manager/internal/handler"
	"github.com/aadityya4real/Task-manager/internal/middleware"
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
	// Users table
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

	// Tasks table
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

	// 🔹 Initialize Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// 🔹 Create Store
	store := storage.New(db)

	http.HandleFunc("/signup", handler.SignupHandler(store))
	http.HandleFunc("/login", handler.LoginHandler(store))

	http.HandleFunc("/task", middleware.AuthMiddleware(
		handler.TaskHandler(store, rdb),
	))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Task Manager API Running 🚀"))
	})
	// 🔹 Start server
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
