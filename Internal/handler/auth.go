package handler

import (
	"encoding/json"
	"net/http"

	"github.com/aadityya4real/Task-manager/internal/storage"
	"github.com/aadityya4real/Task-manager/internal/types"
	"github.com/aadityya4real/Task-manager/internal/utils"
)

// 🔹 SIGNUP
func SignupHandler(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var u types.User

		err := json.NewDecoder(r.Body).Decode(&u)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if u.Username == "" || u.Password == "" {
			http.Error(w, "Username and password required", http.StatusBadRequest)
			return
		}

		id, err := store.CreateUser(u)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		u.ID = int(id)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(u)
	}
}

// 🔹 LOGIN
func LoginHandler(store *storage.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var u types.User

		err := json.NewDecoder(r.Body).Decode(&u)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		dbUser, err := store.GetUser(u.Username)
		if err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}

		if dbUser.Password != u.Password {
			http.Error(w, "Invalid password", http.StatusUnauthorized)
			return
		}

		token, err := utils.GenerateToken(dbUser.Username)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"token": token,
		})
	}
}
