/*
 *   Copyright (c) 2024 Edward Stock

 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.

 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.

 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"database/sql"
	"fmt"
)

func loadDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "arcWiki.db")
	return db, err
}
func dbSetup() {

	dbName := "arcWiki.db"
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Settings (installed BOOLEAN UNIQUE NOT NULL DEFAULT FALSE); `)
	if err != nil {
		fmt.Println("Error creating Settings table:", err)
		return
	} else {
		query := `SELECT installed FROM Settings`
		row := db.QueryRow(query)

		var installed bool
		err = row.Scan(&installed)
		var status string
		if installed {
			status = "TRUE"
		} else {
			status = "FALSE"
		}
		fmt.Println("The current value of installed is:", status)
		if err != nil {
			fmt.Println(err)
		}

		if !installed {
			fmt.Println("now creating tables")

			_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL UNIQUE,
		body TEXT,
		user_id INTEGER,  
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);  `)
			if err != nil {
				fmt.Println("Error creating Categories table:", err)
				return
			}

			_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Pages (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL UNIQUE,
	body TEXT,
	user_id INTEGER, 
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME
	);`)
			if err != nil {
				fmt.Println("Error creating Categories table:", err)
				return
			}

			_, err = db.Exec(`CREATE TABLE IF NOT EXISTS Subcategories (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL UNIQUE,
	description TEXT,
	parent_id INTEGER REFERENCES Categories(id) ON DELETE CASCADE,
	user_id INTEGER,  
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`)
			if err != nil {
				fmt.Println("Error creating Categories table:", err)
				return
			}

			_, err = db.Exec(`CREATE TABLE IF NOT EXISTS CategoryPages (
	page_id INTEGER REFERENCES Pages(id) ON DELETE CASCADE,
	category_id INTEGER REFERENCES Categories(id) ON DELETE CASCADE
	);`)
			if err != nil {
				fmt.Println("Error creating Categories table:", err)
				return
			}

			_, err = db.Exec(`CREATE TABLE IF NOT EXISTS SubCategoryPages (
			subcategory_id INTEGER REFERENCES Subcategories(id) ON DELETE CASCADE,
			category_id INTEGER REFERENCES Categories(id) ON DELETE CASCADE
		);`)
			if err != nil {
				fmt.Println("Error creating Categories table:", err)
				return
			}
			_, err = db.Exec(`CREATE TABLE IF NOT EXISTS  Users (
		id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
		password BLOB NOT NULL, 
		email CLOB NOT NULL,  
		username BLOB DEFAULT '' NOT NULL   
		)`)
			if err != nil {
				fmt.Println("Error creating Categories table:", err)
				return
			}
			_, err = db.Exec(`INSERT INTO Settings (installed) VALUES (TRUE)`)
			if err != nil {
				fmt.Println("Error updating installed value:", err)
				return
			} else {
				fmt.Println("Installed value updated successfully")
			}
			_, err = db.Exec(`INSERT INTO Pages (title, body, user_id, created_at, updated_at) VALUES (?, ?, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, "Main_page", "## Welcome to ArcWiki\n let the games begin")

			if err != nil {
				fmt.Println("Error inserting into pages value:", err)
				return
			} else {
				fmt.Println("Installed value updated successfully")
			}
			//fmt.Println("Tables created successfully!")
			// }
			defer db.Close()
			//fmt.Println("Database opened successfully!")
		} else {
			fmt.Println("it's already installed please delete the database if you wish for a fresh install")
		}
	}
}
