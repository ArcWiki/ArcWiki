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
	"strings"
	"time"

	"github.com/ArcWiki/ArcWiki/db"
	"github.com/gomarkdown/markdown"
	"github.com/houseme/mobiledetect"
)

type Result struct {
	Title string `sql:"title"`
	Body  string `sql:"body"`
}

type SearchData struct {
	ThemeColor template.HTML
	NavTitle   string
	Menu       template.HTML
	CTitle     string
	Results    []Result
	Size       template.HTML
}

func QueryHandler(w http.ResponseWriter, r *http.Request) {

	// Access form data from request object
	query := r.FormValue("query")
	fmt.Println("what you submitted:", query)

	// Get all titles from the database

	db, err := db.LoadDatabase()
	if err != nil {
		panic(err) // Handle errors appropriately in production
	}
	rows, err := db.Query("SELECT title, body FROM Pages WHERE title LIKE ?", "%"+query+"%")
	if err != nil {
		// Handle error
		return
	}
	defer rows.Close()

	defer rows.Close() // Close the rows after iterating
	safeMenu, err := loadMenu()
	if err != nil {
		// Handle error
		return
	}
	userAgent := ""
	size := ""
	detect := mobiledetect.New(r, nil)

	if detect.IsMobile() || detect.IsTablet() {
		fmt.Println("is either a mobile or tablet")
		userAgent = "mobile"
	} else {
		userAgent = "desktop"
	}
	if userAgent == "desktop" {
		size = "<div class=\"col-11 d-none d-sm-block\">"
	} else {
		size = "<div class=\"col-12 d-block d-sm-none\">"
	}
	var data = SearchData{
		ThemeColor: template.HTML(arcWikiLogo()),
		Size:       template.HTML(size),
		Menu:       safeMenu,
		NavTitle:   config.SiteTitle,
		CTitle:     "Search",
	}

	for rows.Next() {
		var result Result
		err := rows.Scan(&result.Title, &result.Body)
		if err != nil {
			fmt.Println("Error with searching database:", err)
			return
		}
		words := strings.Fields(result.Body) // Split on whitespace
		if len(words) > 7 {                  // Adjust limit as needed (7 words in this example)
			result.Body = strings.Join(words[:7], " ") + "..." // Join limited words and add ellipsis
		} else {
			result.Body = strings.Join(words, " ") // Join all words if under limit
		}
		data.Results = append(data.Results, result)
	}

	templates := template.New("") // Create a new template set
	templates, err = templates.ParseFiles("templates/results.html", "templates/navbar.html", "templates/header.html", "templates/footer.html")
	if err != nil {
		// Handle template parsing error
		fmt.Println("Error parsing templates:", err)
		return
	}

	// ... (your search handler logic to get results)

	// Execute the relevant template with data
	err = templates.ExecuteTemplate(w, "results.html", data) // Assuming search results are in "results"
	if err != nil {
		// Handle template execution error
		fmt.Println("Error executing template:", err)
	}

	//renderTemplate(w, "search", p)

	//return &Search{Rtitle: titles, Rbody: bodies}, nil

	//return titles, nil

}

func SearchHandler(w http.ResponseWriter, r *http.Request, title string, userAgent string) {
	//category := r.URL.Path[len("/title/"):]
	if len(r.URL.Path) >= len("/title/") && r.URL.Path[:len("/title/")] == "/title/" {
		// Path starts with "/title/" and has enough characters for slicing
		//category := r.URL.Path[len("/title/"):]

		// Process the extracted category

		switch {

		default:

			fmt.Println("No category specified, defaulting to normal view")

			// Load the page for standard title viewing
			p, err := LoadNothing(title, userAgent)
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
		p, err := LoadNothing("Main_page", userAgent)
		if err != nil {
			fmt.Println("viewHandler: Something weird happened")
			//http.Redirect(w, r, "/title/Main_page", http.StatusFound)
			return
		}
		renderTemplate(w, "search", p)
	}

}
func LoadNothing(title string, userAgent string) (*Page, error) {

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
	db, err := db.LoadDatabase()
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
		return &Page{NavTitle: config.SiteTitle, ThemeColor: template.HTML(arcWikiLogo()), CTitle: "Search", Title: "title", Body: "", Size: template.HTML(size), Menu: safeMenu, CategoryLink: categoryLink, UpdatedDate: footer}, nil
	} else if err != sql.ErrNoRows { // Handle other SQLite errors
		return nil, err
	}

	return &Page{NavTitle: config.SiteTitle, ThemeColor: template.HTML(arcWikiLogo()), CTitle: "Search", Title: "title", Body: "", Size: template.HTML(size), Menu: safeMenu, UpdatedDate: "footer"}, nil
	//return nil, fmt.Errorf("File not found: %s.txt", title) // File not found in any folder
}
