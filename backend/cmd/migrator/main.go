package main

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

type migration struct {
	name string
	sql string
}

func main() {
	const DB_PATH = "/db/app.db"

	db, err := sql.Open("sqlite", DB_PATH)
	if err != nil {
		log.Fatal("Failed to open sqlite database: ", err)
	}
	defer db.Close()

	migrations := []migration {
		{
			name: "users",
			sql: `
			CREATE TABLE IF NOT EXISTS users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				discord_id TEXT UNIQUE NOT NULL
			);`,
		},
		{
			name: "refresh_tokens",
			sql: `
			CREATE TABLE IF NOT EXISTS refresh_tokens (
				id TEXT PRIMARY KEY,
				user_id INTEGER NOT NULL,
				expires_at DATETIME NOT NULL,
				FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
			);`,
		},
	}

	for _, m := range migrations {
		_, err := db.Exec(m.sql)
		if err != nil {
			log.Fatalf("failed to migrate table %s: %v", m.name, err)
		}
	}
}