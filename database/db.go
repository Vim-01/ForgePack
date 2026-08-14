package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func InitDB(path string, ownerID int64) *DB {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}

	createTables(conn)

	db := &DB{conn: conn}

	if ownerID != 0 {
		db.AllowUser(ownerID)
	}

	return db
}

func createTables(conn *sql.DB) {
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY,
		is_allowed BOOLEAN DEFAULT 0
	);`
	if _, err := conn.Exec(usersTable); err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}

	statesTable := `
	CREATE TABLE IF NOT EXISTS user_states (
		user_id INTEGER PRIMARY KEY,
		state TEXT DEFAULT 'IDLE',
		current_pack_name TEXT DEFAULT '',
		current_pack_title TEXT DEFAULT '',
		current_pack_type TEXT DEFAULT 'static'
	);`
	if _, err := conn.Exec(statesTable); err != nil {
		log.Fatalf("Failed to create user_states table: %v", err)
	}
}

func (db *DB) AllowUser(userID int64) error {
	_, err := db.conn.Exec(`INSERT INTO users (id, is_allowed) VALUES (?, 1) ON CONFLICT(id) DO UPDATE SET is_allowed=1`, userID)
	return err
}

func (db *DB) RevokeUser(userID int64) error {
	_, err := db.conn.Exec(`UPDATE users SET is_allowed=0 WHERE id=?`, userID)
	return err
}

func (db *DB) IsAllowed(userID int64) bool {
	var isAllowed bool
	err := db.conn.QueryRow(`SELECT is_allowed FROM users WHERE id = ?`, userID).Scan(&isAllowed)
	if err != nil {
		return false
	}
	return isAllowed
}

type UserState struct {
	State            string
	CurrentPackName  string
	CurrentPackTitle string
	CurrentPackType  string // "static", "video", "animated"
}

func (db *DB) GetState(userID int64) UserState {
	var s UserState
	err := db.conn.QueryRow(`SELECT state, current_pack_name, current_pack_title, current_pack_type FROM user_states WHERE user_id = ?`, userID).
		Scan(&s.State, &s.CurrentPackName, &s.CurrentPackTitle, &s.CurrentPackType)
	if err != nil {
		return UserState{State: "IDLE"}
	}
	return s
}

func (db *DB) SetState(userID int64, state, packName, packTitle, packType string) error {
	_, err := db.conn.Exec(`INSERT INTO user_states (user_id, state, current_pack_name, current_pack_title, current_pack_type) 
	VALUES (?, ?, ?, ?, ?) ON CONFLICT(user_id) DO UPDATE SET 
	state=?, current_pack_name=?, current_pack_title=?, current_pack_type=?`,
		userID, state, packName, packTitle, packType,
		state, packName, packTitle, packType)
	return err
}
