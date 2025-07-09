package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// LoadDatabase opens (or creates) arcWiki.db
func LoadDatabase() (*sql.DB, error) {
	return sql.Open("sqlite3", "arcWiki.db")
}

// DBSetup ensures schema is in place and seeds initial data once.
func DBSetup() {
	db, err := LoadDatabase()
	if err != nil {
		log.Fatalf("Error opening DB: %v", err)
	}
	defer db.Close()

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS Settings (
            installed BOOLEAN UNIQUE NOT NULL DEFAULT FALSE
        );`,
		`CREATE TABLE IF NOT EXISTS Categories (
            id          INTEGER PRIMARY KEY AUTOINCREMENT,
            title       TEXT    NOT NULL UNIQUE,
            body        TEXT,
            user_id     INTEGER,
            created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
        );`,
		`CREATE TABLE IF NOT EXISTS Pages (
            id          INTEGER PRIMARY KEY AUTOINCREMENT,
            title       TEXT    NOT NULL UNIQUE,
            body        TEXT,
            user_id     INTEGER,
            created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at  DATETIME
        );`,
		`CREATE TABLE IF NOT EXISTS Subcategories (
            id          INTEGER PRIMARY KEY AUTOINCREMENT,
            name        TEXT    NOT NULL UNIQUE,
            description TEXT,
            parent_id   INTEGER REFERENCES Categories(id) ON DELETE CASCADE,
            user_id     INTEGER,
            created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
        );`,
		`CREATE TABLE IF NOT EXISTS CategoryPages (
            page_id     INTEGER REFERENCES Pages(id) ON DELETE CASCADE,
            category_id INTEGER REFERENCES Categories(id) ON DELETE CASCADE
        );`,
		`CREATE TABLE IF NOT EXISTS SubCategoryPages (
            subcategory_id INTEGER REFERENCES Subcategories(id) ON DELETE CASCADE,
            category_id    INTEGER REFERENCES Categories(id) ON DELETE CASCADE
        );`,
		`CREATE TABLE IF NOT EXISTS Users (
            id        INTEGER PRIMARY KEY AUTOINCREMENT,
            username  TEXT    NOT NULL UNIQUE,
            password  BLOB    NOT NULL,
            email     TEXT,
            is_admin  INTEGER NOT NULL DEFAULT 0
        );`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			log.Fatalf("Error creating table: %v", err)
		}
	}

	var installed bool
	err = db.QueryRow(`SELECT installed FROM Settings`).Scan(&installed)
	if err == sql.ErrNoRows || !installed {
		if _, err := db.Exec(`INSERT OR REPLACE INTO Settings(installed) VALUES(TRUE)`); err != nil {
			log.Fatalf("Error setting installed flag: %v", err)
		}
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO Pages(title,body,user_id,created_at,updated_at) VALUES(?,?,?,?,?)`,
			"Main_page",
			"## Welcome to ArcWiki\nlet the games begin",
			1,
			time.Now(),
			time.Now(),
		); err != nil {
			log.Fatalf("Error seeding Pages: %v", err)
		}
	} else if err != nil {
		log.Fatalf("Error checking installer flag: %v", err)
	}
}
