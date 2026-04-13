package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	_ "modernc.org/sqlite"
)

// Task struct
type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

// Handler with DB
func taskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 🔹 POST → Create Task
		if r.Method == "POST" {
			var t Task

			err := json.NewDecoder(r.Body).Decode(&t)
			if err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			if t.Title == "" {
				http.Error(w, "Title is required", http.StatusBadRequest)
				return
			}

			result, err := db.Exec(
				"INSERT INTO tasks (title, done) VALUES (?, ?)",
				t.Title, false,
			)
			if err != nil {
				http.Error(w, "Failed to insert", http.StatusInternalServerError)
				return
			}

			id, _ := result.LastInsertId()
			t.ID = int(id)
			t.Done = false

			json.NewEncoder(w).Encode(t)
			return
		}

		// 🔹 GET → Get all tasks
		if r.Method == "GET" {
			rows, err := db.Query("SELECT id, title, done FROM tasks")
			if err != nil {
				http.Error(w, "Failed to fetch", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var tasks []Task

			for rows.Next() {
				var t Task
				err := rows.Scan(&t.ID, &t.Title, &t.Done)
				if err != nil {
					http.Error(w, "Error reading data", http.StatusInternalServerError)
					return
				}
				tasks = append(tasks, t)
			}

			json.NewEncoder(w).Encode(tasks)
			return
		}

		// 🔹 PUT → Update Task
		if r.Method == "PUT" {
			idStr := r.URL.Query().Get("id")
			if idStr == "" {
				http.Error(w, "ID is required", http.StatusBadRequest)
				return
			}

			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}

			var t Task
			err = json.NewDecoder(r.Body).Decode(&t)
			if err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			_, err = db.Exec(
				"UPDATE tasks SET title = ?, done = ? WHERE id = ?",
				t.Title, t.Done, id,
			)
			if err != nil {
				http.Error(w, "Failed to update", http.StatusInternalServerError)
				return
			}

			json.NewEncoder(w).Encode(map[string]string{
				"status": "updated",
			})
			return
		}

		// 🔹 DELETE → Delete Task
		if r.Method == "DELETE" {
			idStr := r.URL.Query().Get("id")
			if idStr == "" {
				http.Error(w, "ID is required", http.StatusBadRequest)
				return
			}

			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid ID", http.StatusBadRequest)
				return
			}

			_, err = db.Exec("DELETE FROM tasks WHERE id = ?", id)
			if err != nil {
				http.Error(w, "Failed to delete", http.StatusInternalServerError)
				return
			}

			json.NewEncoder(w).Encode(map[string]string{
				"status": "deleted",
			})
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {

	// 🔹 Connect DB
	db, err := sql.Open("sqlite", "tasks.db")
	if err != nil {
		panic(err)
	}

	// 🔹 Create Table
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT,
		done BOOLEAN
	)
	`)
	if err != nil {
		panic(err)
	}

	// 🔹 Root route
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Task Manager API Running 🚀"))
	})

	// 🔹 Task route
	http.HandleFunc("/task", taskHandler(db))

	// 🔹 Start server
	http.ListenAndServe(":8080", nil)
}
