package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/aadityya4real/task-manager/internal/storage"
	"github.com/aadityya4real/task-manager/internal/types"

	"github.com/redis/go-redis/v9"
)

// cacheKeyForTasks builds a pagination-aware cache key per user.
func cacheKeyForTasks(userID, limit, offset int) string {
	return fmt.Sprintf("tasks:%d:%d:%d", userID, limit, offset)
}

// invalidateUserCache deletes all cached pages for a user using SCAN.
// This is necessary because pagination creates multiple keys per user.
func invalidateUserCache(ctx context.Context, rdb *redis.Client, userID int) {
	pattern := fmt.Sprintf("tasks:%d:*", userID)
	var cursor uint64

	for {
		keys, nextCursor, err := rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			fmt.Println("⚠️ REDIS SCAN ERROR:", err)
			return
		}
		if len(keys) > 0 {
			if err := rdb.Del(ctx, keys...).Err(); err != nil {
				fmt.Println("⚠️ REDIS DEL ERROR:", err)
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	fmt.Println("🧹 CACHE CLEARED for user:", userID)
}

func TaskHandler(store *storage.Store, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// FIX 1: Use request-scoped context, not a package-level background context.
		// This respects cancellation and deadlines from the HTTP request lifecycle.
		ctx := r.Context()

		// FIX 2: Safe type assertion — panics if middleware didn't set user_id.
		userID, ok := ctx.Value("user_id").(int)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		switch r.Method {

		// ─── POST → Create Task ────────────────────────────────────────────────
		case http.MethodPost:
			var t types.Task

			if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
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

			// FIX 3: Invalidate all paginated cache keys for this user.
			invalidateUserCache(ctx, rdb, userID)

			t.ID = int(id)
			t.Done = false

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"message": "Task created successfully",
				"data":    t,
			})

			fmt.Println("POST /tasks | user:", userID)
			return

		// ─── GET → Fetch Tasks (with Redis + Pagination) ──────────────────────
		case http.MethodGet:

			// FIX 4: Parse pagination BEFORE building the cache key so the key
			// is page-specific. Previously the key was page-agnostic, so all
			// pages returned the same cached data.
			limitStr := r.URL.Query().Get("limit")
			offsetStr := r.URL.Query().Get("offset")

			limit := 10
			offset := 0

			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
			if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
				offset = o
			}

			key := cacheKeyForTasks(userID, limit, offset)

			// Try Redis cache first.
			data, err := rdb.Get(ctx, key).Result()
			if err == nil {
				fmt.Println("⚡ CACHE HIT for user:", userID, "| limit:", limit, "| offset:", offset)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(data))
				return
			}

			if err != redis.Nil {
				// Non-nil error means Redis itself is having issues; log and fall through.
				fmt.Println("⚠️ REDIS ERROR:", err)
			} else {
				fmt.Println("❌ CACHE MISS for user:", userID, "| limit:", limit, "| offset:", offset)
			}

			// Fetch from DB on cache miss.
			tasks, err := store.GetTasks(userID, limit, offset)
			if err != nil {
				http.Error(w, "Failed to fetch", http.StatusInternalServerError)
				return
			}

			response := map[string]interface{}{
				"message": "Tasks fetched successfully",
				"count":   len(tasks),
				"data":    tasks,
			}

			jsonData, err := json.Marshal(response)
			if err != nil {
				http.Error(w, "JSON error", http.StatusInternalServerError)
				return
			}

			// Populate cache for this specific page.
			if err := rdb.Set(ctx, key, jsonData, 5*time.Minute).Err(); err != nil {
				fmt.Println("⚠️ REDIS SET ERROR:", err)
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonData)

			fmt.Println("GET /tasks | user:", userID, "| limit:", limit, "| offset:", offset)
			return
		// ─── PUT → Update Task ────────────────────────────────────────────────
		case http.MethodPut:

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
			if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			// NOTE: Ensure store.UpdateTask uses WHERE id = ? AND user_id = ?
			// to prevent a user from updating another user's task.
			if err := store.UpdateTask(id, userID, t); err != nil {
				http.Error(w, "Failed to update", http.StatusInternalServerError)
				return
			}

			// FIX 5: Invalidate all paginated cache keys for this user.
			invalidateUserCache(ctx, rdb, userID)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Task updated successfully",
			})

			fmt.Println("PUT /tasks | user:", userID, "| task ID:", id)

		// ─── DELETE → Delete Task ─────────────────────────────────────────────
		case http.MethodDelete:

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

			// NOTE: Ensure store.DeleteTask uses WHERE id = ? AND user_id = ?
			// to prevent a user from deleting another user's task.
			if err := store.DeleteTask(id, userID); err != nil {
				http.Error(w, "Failed to delete", http.StatusInternalServerError)
				return
			}

			// FIX 5: Invalidate all paginated cache keys for this user.
			invalidateUserCache(ctx, rdb, userID)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Task deleted successfully",
			})

			fmt.Println("DELETE /tasks | user:", userID, "| task ID:", id)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)

			return
		}
	}
}
