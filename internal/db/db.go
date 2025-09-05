package db

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sqlx.DB

// Connect opens a global SQLite connection
func Connect(dataSourceName string) {
	var err error
	DB, err = sqlx.Connect("sqlite3", dataSourceName)
	if err != nil {
		log.Fatal("❌ cannot connect to sqlite:", err)
	}
	log.Println("✅ connected to sqlite database:", dataSourceName)
}
