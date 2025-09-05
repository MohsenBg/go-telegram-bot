package models

import (
	"tel-bot/internal/db"
)

type UserChat struct {
	Step        int
	Username    string
	Information User
}

type User struct {
	ID          int64  `db:"id"`
	ChatID      int64  `db:"chat_id"`
	Username    string `db:"username"`
	Name        string `db:"name"`
	Phone       string `db:"phone"`
	Companions  int    `db:"companions"`
	Major       string `db:"major"`
	Transaction string `db:"payment_transaction"`
	StudentID   string `db:"student_id"`
	CreatedAt   string `db:"created_at"`
}

// CreateUsersTable ensures the users table exists
func CreateUsersTable() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_id INTEGER NOT NULL,
		username TEXT,
		name TEXT NOT NULL,
		phone TEXT NOT NULL,
		companions INTEGER NOT NULL,
		major TEXT NOT NULL,
		payment_transaction TEXT NOT NULL,
		student_id TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.DB.Exec(schema)
	return err
}

// Save inserts a new user record
func (u *User) Save() error {
	_, err := db.DB.NamedExec(`
		INSERT INTO users (chat_id, username, name, phone, companions, major, payment_transaction, student_id)
		VALUES (:chat_id, :username, :name, :phone, :companions, :major, :payment_transaction, :student_id)
	`, u)
	return err
}

// GetByChatID fetches all users by chat_id
func GetByChatID(chatID int64) ([]User, error) {
	var users []User
	err := db.DB.Select(&users, "SELECT * FROM users WHERE chat_id = ?", chatID)
	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetAllUsers returns all users ordered by newest first
func GetAllUsers() ([]User, error) {
	var users []User
	err := db.DB.Select(&users, "SELECT * FROM users ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	return users, nil
}
