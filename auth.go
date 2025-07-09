/*
 *   Copyright (c) 2024 Edward Stock
 *
 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.
 *
 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.
 *
 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const envFile = ".env"

// session store
var store *sessions.CookieStore

// database handle for auth
var authDB *sql.DB

func init() {
	// Ensure .env with SESSION_KEY
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		raw := make([]byte, 32)
		if _, err := rand.Read(raw); err != nil {
			log.Fatalf("auth: cannot generate session key: %v", err)
		}
		b64 := base64.StdEncoding.EncodeToString(raw)
		if err := ioutil.WriteFile(envFile, []byte("SESSION_KEY="+b64+"\n"), 0600); err != nil {
			log.Fatalf("auth: cannot write %s: %v", envFile, err)
		}
	}

	// Load SESSION_KEY
	_ = godotenv.Load(envFile)
	b64key := os.Getenv("SESSION_KEY")
	if b64key == "" {
		log.Fatal("auth: SESSION_KEY not set")
	}
	key, err := base64.StdEncoding.DecodeString(b64key)
	if err != nil {
		log.Fatalf("auth: invalid SESSION_KEY: %v", err)
	}

	// Initialize Gorilla session store
	store = sessions.NewCookieStore(key)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

// InitAuthDB opens (or creates) arcWiki.db and ensures users table exists.
func InitAuthDB(path string) error {
	var err error
	authDB, err = sql.Open("sqlite3", path)
	if err != nil {
		return err
	}
	// Create or migrate the users table
	_, err = authDB.Exec(`
		 CREATE TABLE IF NOT EXISTS users (
			 id        INTEGER PRIMARY KEY AUTOINCREMENT,
			 username  TEXT    NOT NULL UNIQUE,
			 password  TEXT    NOT NULL
		 );`)
	if err != nil {
		return err
	}
	// Add is_admin column if missing
	_, err = authDB.Exec(`ALTER TABLE users ADD COLUMN is_admin INTEGER NOT NULL DEFAULT 0;`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
		return err
	}
	return nil
}

// CreateUser inserts a bcrypt-hashed user. No-op if username exists.
func CreateUser(username, plainPassword string, isAdmin bool) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = authDB.Exec(
		"INSERT INTO users(username,password,is_admin) VALUES(?,?,?)",
		username, string(hash), boolToInt(isAdmin),
	)
	if err != nil && !isUniqueConstraintError(err) {
		return err
	}
	return nil
}

// Authenticate checks credentials; returns (ok, isAdmin, error).
func Authenticate(username, plainPassword string) (bool, bool, error) {
	var storedHash string
	var isAdminInt int
	err := authDB.QueryRow(
		"SELECT password,is_admin FROM users WHERE username = ?",
		username,
	).Scan(&storedHash, &isAdminInt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, false, nil
		}
		return false, false, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(plainPassword)); err != nil {
		return false, false, nil
	}
	return true, isAdminInt == 1, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func isUniqueConstraintError(err error) bool {
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// Presents the login form
func loginFormHandler(w http.ResponseWriter, r *http.Request, title string, userAgent string) {
	size := ""
	if userAgent == Desktop {
		size = `<div class="col-11 d-none d-sm-block">`
	} else {
		size = `<div class="col-12 d-block d-sm-none">`
	}
	bodyMark := `<form action="/loginPost" method="post">
		 <div class="form-group">
		   <label for="username">Username:</label>
		   <input class="form-control" type="text" id="username" name="username">
		 </div>
		 <div class="form-group">
		   <label for="password">Password:</label>
		   <input class="form-control" type="password" id="password" name="password">
		 </div>
		 <button class="bg-dark hover:bg-gray-100 text-white font-semibold py-2 px-4 border border-gray-400 rounded shadow" type="submit">Login</button>
	 </form>`

	parsedText := addHeadingIDs(bodyMark)
	happyhtml := createHeadingList(parsedText)
	categoryLink := findAllCategoryLinks(happyhtml)
	noLinks := removeCategoryLinks(happyhtml)
	perfecthtml := parseWikiText(noLinks)
	internalLinks := convertLinksToAnchors(perfecthtml)
	safeBodyHTML := template.HTML(internalLinks)

	safeMenu, err := loadMenu()
	if err != nil {
		log.Error("error loading menu")
	}

	p := Page{
		NavTitle:     config.SiteTitle,
		ThemeColor:   template.HTML(arcWikiLogo()),
		CTitle:       removeUnderscores(title),
		Title:        "login",
		Body:         safeBodyHTML,
		Size:         template.HTML(size),
		Menu:         safeMenu,
		CategoryLink: categoryLink,
	}

	renderTemplate(w, "login", &p)
}

// Logs the user out by clearing the session
func logout(w http.ResponseWriter, r *http.Request, title string, userAgent string) {
	session, _ := store.Get(r, "cookie-name")
	session.Values["authenticated"] = false
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

// Processes login submissions
func loginHandler(w http.ResponseWriter, r *http.Request, title string, userAgent string) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	user := r.FormValue("username")
	pass := r.FormValue("password")
	ok, isAdmin, err := Authenticate(user, pass)
	if err != nil {
		log.Errorf("auth error: %v", err)
		http.Redirect(w, r, "/error", http.StatusSeeOther)
		return
	}
	if !ok {
		http.Redirect(w, r, "/error", http.StatusSeeOther)
		return
	}
	session, _ := store.Get(r, "cookie-name")
	session.Values["authenticated"] = true
	session.Values["is_admin"] = isAdmin
	session.Save(r, w)
	log.Infof("Logged in: %s (admin=%v)", user, isAdmin)
	http.Redirect(w, r, "/admin", http.StatusFound)
}
