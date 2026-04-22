package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("mysecretkey")

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// 🔹 Get header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		// 🔹 Extract token
		parts := strings.Split(authHeader, "Bearer ")
		if len(parts) != 2 {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]

		// 🔹 Parse token
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// 🔥 Extract claims safely
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// ✅ SAFE extraction
		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		userID := int(userIDFloat)

		// 🔹 username extraction
		username, ok := claims["username"].(string)
		if !ok {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// ✅ store in context
		ctx := context.WithValue(r.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "username", username)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
