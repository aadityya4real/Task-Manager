package storage

import (
	"github.com/aadityya4real/task-manager/internal/types"
)

// Insert User
func (s *Store) CreateUser(u types.User) (int64, error) {
	result, err := s.DB.Exec(
		"INSERT INTO users (username, password) VALUES (?, ?)",
		u.Username, u.Password,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// Get User (for login)
func (s *Store) GetUser(username string) (types.User, error) {
	var u types.User
	err := s.DB.QueryRow(
		"SELECT id, username, password FROM users WHERE username = ?",
		username,
	).Scan(&u.ID, &u.Username, &u.Password)

	return u, err
}
