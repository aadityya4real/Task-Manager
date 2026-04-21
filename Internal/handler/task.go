package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/aadityya4real/Task-manager/internal/storage"
	"github.com/aadityya4real/Task-manager/internal/types"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func TaskHandler(store *storage.Store, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("user_id").(int)
		key := fmt.Sprintf("tasks:%d", userID)

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

			id, err := store.InsertTask(t, userID)
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

			userID := r.Context().Value("user_id").(int)

			// 🔥 ADD THIS (pagination)
			limitStr := r.URL.Query().Get("limit")
			offsetStr := r.URL.Query().Get("offset")

			limit := 10
			offset := 0

			if l, err := strconv.Atoi(limitStr); err == nil {
				limit = l
			}
			if o, err := strconv.Atoi(offsetStr); err == nil {
				offset = o
			}

			// 🔥 PASS limit + offset
			tasks, err := store.GetTasks(userID, limit, offset)
			if err != nil {
				http.Error(w, "Failed to fetch", http.StatusInternalServerError)
				return
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

			var t types.Task
			err = json.NewDecoder(r.Body).Decode(&t)
			if err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			err = store.UpdateTask(id, userID, t)
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

			err = store.DeleteTask(id, userID)
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
