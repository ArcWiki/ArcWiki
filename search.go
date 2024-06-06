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
	"html/template"
	"net/http"
	"time"

	"github.com/gomarkdown/markdown"
)

type Search struct {
	Rtitle []string
	Rbody  []string
}

func queryHandler(w http.ResponseWriter, r *http.Request) {

	// Access form data from request object
	query := r.FormValue("query")
	fmt.Println("what you submitted:", query)

	// Get all titles from the database

	db, err := loadDatabase()
	if err != nil {
		panic(err) // Handle errors appropriately in production
	}
	var titles, bodies []string
	sql := fmt.Sprintf("SELECT title, body FROM Pages WHERE title LIKE '%%%s%%' ORDER BY title DESC", query)
	rows, err := db.Query(sql)

	defer rows.Close() // Close the rows after iterating

	for rows.Next() {
		var title, body string
		err := rows.Scan(&title)
		if err != nil {
			fmt.Println("error")
		}
		titles = append(titles, title)
		bodies = append(bodies, body)

		fmt.Println(title + body) // Print each title during iteration (for testing)
	}
	//renderTemplate(w, "search", p)

	//return &Search{Rtitle: titles, Rbody: bodies}, nil

	//return titles, nil

}

func searchHandler(w http.ResponseWriter, r *http.Request, title string, userAgent string) {
	//category := r.URL.Path[len("/title/"):]
	if len(r.URL.Path) >= len("/title/") && r.URL.Path[:len("/title/")] == "/title/" {
		// Path starts with "/title/" and has enough characters for slicing
		//category := r.URL.Path[len("/title/"):]

		// Process the extracted category

		switch {

		default:

			fmt.Println("No category specified, defaulting to normal view")

			// Load the page for standard title viewing
			p, err := loadNothing(title, userAgent)
			if err != nil {
				fmt.Println("viewHandler: Something weird happened")
				http.Redirect(w, r, "/title/Main_page", http.StatusFound)
				return
			}

			renderTemplate(w, "search", p)

		}
	} else {
		//fmt.Println("hello beautiful world")
		// Load the page for standard title viewing
		p, err := loadNothing("Main_page", userAgent)
		if err != nil {
			fmt.Println("viewHandler: Something weird happened")
			//http.Redirect(w, r, "/title/Main_page", http.StatusFound)
			return
		}
		renderTemplate(w, "search", p)
	}

}
func loadNothing(title string, userAgent string) (*Page, error) {

	safeMenu, err := loadMenu()
	if err != nil {
		return nil, err
	}
	size := ""
	if userAgent == "desktop" {
		size = "<div class=\"col-11 d-none d-sm-block\">"
	} else {
		size = "<div class=\"col-12 d-block d-sm-none\">"
	}
	db, err := loadDatabase()
	if err != nil {
		return nil, err
	}

	stmt, err := db.Prepare("SELECT title, body, updated_at FROM Pages WHERE title = ?")
	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(title)

	defer db.Close()   // Close the database connection
	defer stmt.Close() // Close the prepared statement

	var body string
	var updated_at time.Time
	err = row.Scan(&title, &body, &updated_at)
	bodyMark := markdown.ToHTML([]byte(body), nil, nil)
	parsedText := addHeadingIDs(string(bodyMark))
	happyhtml := createHeadingList(parsedText)
	//This grabs all Category links
	categoryLink := findAllCategoryLinks(happyhtml)
	//noLinks := removeCategoryLinks(happyhtml)
	//fmt.Println(noLinks)
	///perfecthtml := parseWikiText(noLinks)

	//internalLinks := convertLinksToAnchors(perfecthtml)
	//safeBodyHTML := template.HTML(internalLinks)
	footer := "This page was last modified on " + formatDateTime(updated_at)

	//need to double check this as I'm not certain why this is
	if err == nil { // Page found in database
		// ... (existing code for markdown parsing and HTML generation)
		return &Page{NavTitle: config.SiteTitle, CTitle: "removeUnderscores(title)", Title: "title", Body: "hey there", Size: template.HTML(size), Menu: safeMenu, CategoryLink: categoryLink, UpdatedDate: footer}, nil
	} else if err != sql.ErrNoRows { // Handle other SQLite errors
		return nil, err
	}

	return &Page{NavTitle: config.SiteTitle, CTitle: "removeUnderscores(title)", Title: "title", Body: "hey there", Size: template.HTML(size), Menu: safeMenu, UpdatedDate: "footer"}, nil
	//return nil, fmt.Errorf("File not found: %s.txt", title) // File not found in any folder
}
