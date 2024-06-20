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
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/ArcWiki/ArcWiki/menu"

	"github.com/gorilla/sessions"
)

var (
	// key must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256)
	key   = []byte("super-secret-key")
	store = sessions.NewCookieStore(key)
)

//showing how to login
// func secret(w http.ResponseWriter, r *http.Request) {
// 	session, _ := store.Get(r, "cookie-name")

// 	// Check if user is authenticated
// 	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
// 		http.Error(w, "Forbidden", http.StatusForbidden)
// 		return
// 	}

// 	// Print secret message
// 	fmt.Fprintln(w, "The cake is a lie!")
// }

func loginFormHandler(w http.ResponseWriter, r *http.Request, title string, userAgent string) {
	size := ""
	if userAgent == "desktop" {
		size = "<div class=\"col-11 d-none d-sm-block\">"
	} else {
		size = "<div class=\"col-12 d-block d-sm-none\">"
	}
	//title = "Login"
	bodyMark :=
		`<form action="/loginPost" method="post">

		<div class="form-group">
		<label for="username">Username:</label>
		<input class="form-control" type="text" id="username" name="username">
		</div>

		<div class="form-group">
		<label for="password">Password:</label>
		<input class="form-control" type="password" id="password" name="password">
		</div>
		
		<button 
		class="bg-dark hover:bg-gray-100 text-white font-semibold py-2 px-4 border border-gray-400 rounded shadow"
		type="submit">Login</button>
	</form>`
	// 	`<a href="/edit/menu"> Edit Menu </a><br />
	// <a href="/add"> Add Page </a>
	// `
	//bodyMark := markdown.ToHTML([]byte(readBody), nil, nil)
	//bodyMark := "hey hey"
	parsedText := addHeadingIDs(string(bodyMark))
	//parsedText := addHeadingIDs(parseToc(parseLink(parseWikiText(string(bodyMark)))))
	happyhtml := createHeadingList(parsedText)
	//This grabs all Category links
	categoryLink := findAllCategoryLinks(happyhtml)
	noLinks := removeCategoryLinks(happyhtml)
	perfecthtml := parseWikiText(noLinks)
	internalLinks := convertLinksToAnchors(perfecthtml)
	safeBodyHTML := template.HTML(internalLinks)
	//load menu
	safeMenu, err := menu.Load()
	if err != nil {
		fmt.Println("error loading menu")
	}

	p := Page{NavTitle: config.SiteTitle, ThemeColor: template.HTML(arcWikiLogo()), CTitle: removeUnderscores(title), Title: "login", Body: safeBodyHTML, Size: template.HTML(size), Menu: safeMenu, CategoryLink: categoryLink}

	// Assuming renderTemplate accepts a string for body content:
	renderTemplate(w, "login", &p) // Pass only the body string
}
func logout(w http.ResponseWriter, r *http.Request, title string, userAgent string) {
	session, _ := store.Get(r, "cookie-name")

	// Revoke users authentication
	session.Values["authenticated"] = false
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func loginHandler(w http.ResponseWriter, r *http.Request, title string, userAgent string) {

	if r.Method == "POST" {
		// Process login credentials
		username := r.FormValue("username")
		password := r.FormValue("password")
		usernamef, passwordf := loadAdmin()
		// Perform authentication (replace with your actual logic)
		if username == usernamef && password == passwordf {
			session, _ := store.Get(r, "cookie-name")
			session.Values["authenticated"] = true
			session.Save(r, w)
			fmt.Println("login success")
			// Authentication successful
			http.Redirect(w, r, "/admin", http.StatusFound)
		} else {
			// Authentication failed
			http.Redirect(w, r, "/error", http.StatusFound)

			//http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		}
	}
}
func loadAdmin() (string, string) {
	// Allow password to be set in docker
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")

	if username != "" && password != "" {
		return username, password // Return credentials if found
	} else {

		file, err := os.Open("config/admin.json")
		if err != nil {
			fmt.Println("Loading Admin Settings Failed: ")
			fmt.Println(err)
		}
		defer file.Close()

		// Decode the JSON data
		var data map[string]interface{}
		err = json.NewDecoder(file).Decode(&data)
		if err != nil {
			fmt.Println("error happened")
		}

		menuItems := data["admin"].([]interface{})

		// Create the list of links
		username := ""
		password := ""
		for _, item := range menuItems {
			linkData := item.(map[string]interface{})
			username = linkData["username"].(string)
			password = linkData["password"].(string)

		}

		return username, password
	}
}
