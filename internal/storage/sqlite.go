package storage

import (
	"database/sql"
	"fmt"

	"github.com/aadityya4real/task-manager/internal/types"
)

type Store struct {
	DB *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{DB: db}
}

func (s *Store) InsertTask(t types.Task, userID int) (int64, error) {
	result, err := s.DB.Exec(
		"INSERT INTO tasks (title, done, user_id) VALUES (?, ?, ?)",
		t.Title, false, userID,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
func (s *Store) GetTasks(userID int, limit, offset int) ([]types.Task, error) {

	rows, err := s.DB.Query(
		"SELECT id, title, done FROM tasks WHERE user_id = ? LIMIT ? OFFSET ?",
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []types.Task

	for rows.Next() {
		var t types.Task
		rows.Scan(&t.ID, &t.Title, &t.Done)
		tasks = append(tasks, t)
	}

	return tasks, nil
}
func (s *Store) UpdateTask(id int, userID int, t types.Task) error {
	result, err := s.DB.Exec(
		"UPDATE tasks SET title = ?, done = ? WHERE id = ? AND user_id = ?",
		t.Title, t.Done, id, userID,
	)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("task not found or not owned by user")
	}

	return nil
}
func (s *Store) DeleteTask(id int, userID int) error {
	result, err := s.DB.Exec(
		"DELETE FROM tasks WHERE id = ? AND user_id = ?",
		id, userID,
	)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("task not found or not owned by user")
	}

	return nil
}
