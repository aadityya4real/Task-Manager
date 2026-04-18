package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"context"

	"github.com/aadityya4real/Task-manager/internal/storage"
	"github.com/aadityya4real/Task-manager/internal/types"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func TaskHandler(store *storage.Store, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value("username").(string)
		key := "tasks:" + username

		if r.Method == "POST" {
			var t types.Task

			err := json.NewDecoder(r.Body).Decode(&t)
			if err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			if t.Title == "" {
				http.Error(w, "Title is required", http.StatusBadRequest)
				return
			}

			id, err := store.InsertTask(t, username)
			if err != nil {
				http.Error(w, "Failed to insert", http.StatusInternalServerError)
				return
			}
			rdb.Del(ctx, key)
			fmt.Println("CACHE CLEARED")

			t.ID = int(id)
			t.Done = false
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(t)
			return
		}

		if r.Method == "GET" {
			// Try Redis first
			data, err := rdb.Get(ctx, key).Result()

			if err == nil {
				fmt.Println("CACHE HIT ✅")
				// ✅ CACHE HIT
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(data))
				return

			} else if err == redis.Nil {
				fmt.Println("CACHE MISS ❌")

			} else {
				fmt.Println("REDIS ERROR ⚠️:", err)
			}

			// Fetch from DB
			tasks, err := store.GetTasks(username)
			if err != nil {
				http.Error(w, "Failed to fetch", http.StatusInternalServerError)
				return
			}

			jsonData, err := json.Marshal(tasks)
			if err != nil {
				http.Error(w, "JSON error", http.StatusInternalServerError)
				return
			}
			err = rdb.Set(ctx, key, jsonData, 5*time.Minute).Err()
			if err != nil {
				fmt.Println("REDIS SET ERROR:", err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonData)
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

			var t types.Task
			err = json.NewDecoder(r.Body).Decode(&t)
			if err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			err = store.UpdateTask(id, t)
			if err != nil {
				http.Error(w, "Failed to update", http.StatusInternalServerError)
				return
			}

			rdb.Del(ctx, key)
			fmt.Println("CACHE CLEARED")
			w.Header().Set("Content-Type", "application/json")
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

			err = store.DeleteTask(id)
			if err != nil {
				http.Error(w, "Failed to delete", http.StatusInternalServerError)
				return
			}
			rdb.Del(ctx, key)
			fmt.Println("CACHE CLEARED")

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"status": "deleted",
			})
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
