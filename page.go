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
	"sort"
	"strings"
	"time"

	"github.com/ArcWiki/ArcWiki/db"
	"github.com/ArcWiki/ArcWiki/menu"

	"github.com/gomarkdown/markdown"
)

type AddPage struct {
	ThemeColor  string
	NavTitle    string
	CTitle      string
	Title       string
	Body        string
	FolderList  []string
	Menu        template.HTML
	Size        template.HTML
	UpdatedDate string
}
type Page struct {
	ThemeColor   string
	NavTitle     string
	CTitle       string
	Title        string
	Body         template.HTML
	Menu         template.HTML
	Size         template.HTML
	CategoryLink []string
	UpdatedDate  string
}

type EditPage struct {
	ThemeColor  string
	NavTitle    string
	CTitle      string
	Title       string
	Body        template.HTML
	Menu        template.HTML
	Size        template.HTML
	UpdatedDate string
}

func (p *Page) save() error {
	fmt.Println("the page saved was called " + p.CTitle)
	db, err := db.LoadDatabase()
	if err != nil {
		return fmt.Errorf("error loading database: %w", err) // Return a descriptive error
	}
	defer db.Close()

	stmt, err := db.Prepare("UPDATE Pages SET title = ?, body = ?, updated_at = CURRENT_TIMESTAMP WHERE title = ?")
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(canonicalizeTitle(p.Title), string(p.Body), p.CTitle) // Ignore RowsAffected
	if err != nil {
		return fmt.Errorf("error executing update: %w", err)
	}

	fmt.Println("Updated page with title:", canonicalizeTitle(p.Title)) // Clearer message
	return nil
}
func addPage(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	title := r.FormValue("title")
	if title != "index" {
		body := r.FormValue("body")

		freshTitle := canonicalizeTitle(title)
		dbsql("INSERT INTO Pages (title, body, user_id, created_at, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)", freshTitle, body, 1)

		http.Redirect(w, r, "/title/"+freshTitle, http.StatusFound)
	} else {
		fmt.Println("cannot be index don't be silly")
		http.Redirect(w, r, "/title/index", http.StatusFound)
	}
}

func (p *Page) deletePage() error {

	db, err := db.LoadDatabase()

	if err != nil {
		panic(err)
	}
	defer db.Close()

	stmt, err := db.Prepare("DELETE FROM Pages WHERE title = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(p.Title)
	if err != nil {
		panic(err)
	}

	rowsDeleted, err := result.RowsAffected()
	if err != nil {
		panic(err)
	}

	if rowsDeleted > 0 {
		fmt.Println("Deleted", rowsDeleted, "category with title:", p.Title)
	} else {
		fmt.Println("No category found with title:", p.Title)
	}
	return nil
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string, userAgent string) {
	titleSave := r.FormValue("title")
	body := r.FormValue("body")

	p := &Page{CTitle: title, Title: titleSave, Body: template.HTML(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/title/"+canonicalizeTitle(titleSave), http.StatusFound)
}
func loadPage(title string, userAgent string) (*Page, error) {

	safeMenu, err := menu.Load()
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
	noLinks := removeCategoryLinks(happyhtml)
	//fmt.Println(noLinks)
	perfecthtml := parseWikiText(noLinks)

	internalLinks := convertLinksToAnchors(perfecthtml)
	safeBodyHTML := template.HTML(internalLinks)
	footer := "This page was last modified on " + formatDateTime(updated_at)

	//need to double check this as I'm not certain why this is
	if err == nil { // Page found in database
		// ... (existing code for markdown parsing and HTML generation)
		return &Page{NavTitle: config.SiteTitle, ThemeColor: arcWikiLogo(), CTitle: removeUnderscores(title), Title: title, Body: safeBodyHTML, Size: template.HTML(size), Menu: safeMenu, CategoryLink: categoryLink, UpdatedDate: footer}, nil
	} else if err != sql.ErrNoRows { // Handle other SQLite errors
		return nil, err
	}

	return &Page{NavTitle: config.SiteTitle, ThemeColor: arcWikiLogo(), CTitle: removeUnderscores(title), Title: title, Body: safeBodyHTML, Size: template.HTML(size), Menu: safeMenu, UpdatedDate: footer}, nil
	//return nil, fmt.Errorf("File not found: %s.txt", title) // File not found in any folder
}

// Loads page with no html applied useful for editing markdown in the edit view
func loadPageNoHtml(title string, userAgent string) (*EditPage, error) {
	size := ""

	safeMenu, err := menu.Load()
	if err != nil {
		return nil, err
	}
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
	var updated_at time.Time
	var body string
	err = row.Scan(&title, &body, &updated_at)
	if err != nil {
		return nil, err
	}
	footer := "This page was last modified on " + formatDateTime(updated_at)
	return &EditPage{NavTitle: config.SiteTitle, ThemeColor: arcWikiLogo(), CTitle: removeUnderscores(title), Title: title, Body: template.HTML(body), Menu: template.HTML(safeMenu), Size: template.HTML(size), UpdatedDate: footer}, nil
}
func loadPageSpecial(title string, categoryName string, userAgent string) (*Page, error) {
	//func loadPageSpecial(title string, categoryName string, userAgent string) (*Page, error) {
	//size := "w-full max-w-7xl mx-auto px-4 py-8"

	size := ""
	if userAgent == "desktop" {
		size = "<div class=\"col-11 d-none d-sm-block\">"
	} else {
		size = "<div class=\"col-12 d-block d-sm-none\">"
	}
	baseURL := "/title/"

	if categoryName == "Categories" {

		db, err := db.LoadDatabase()
		if err != nil {
			return nil, err
		}
		defer db.Close()

		stmt, err := db.Prepare("SELECT title FROM Categories")
		if err != nil {
			return nil, err
		}
		defer stmt.Close()

		rows, err := stmt.Query()
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var categories []string // Slice to store category names
		for rows.Next() {
			var name string
			err := rows.Scan(&name) // Scan the "name" column into the variable
			if err != nil {
				return nil, err
			}
			//categories = append(categories, name)
			categories = append(categories, fmt.Sprintf("<li><a href=\"%sCategory:%s\">%s</a></li>", baseURL, name, name))
		}

		sort.Strings(categories) // Sort alphabetically

		bodyHTML := fmt.Sprintf("<h2 class=\"wikih2\">All Categories</h2><ul>\n%s\n</ul>", strings.Join(categories, "\n"))
		safeMenu, err := menu.Load()
		if err != nil {
			return nil, err // Return error if menu file reading fails
		}
		return &Page{

			NavTitle:   config.SiteTitle,
			ThemeColor: arcWikiLogo(),
			CTitle:     "Special:AllCategories",
			Title:      "Special:AllCategories",
			Body:       template.HTML(bodyHTML),
			Size:       template.HTML(size),
			Menu:       template.HTML(safeMenu),
		}, nil
	} else if categoryName == "AllPages" {
		db, err := db.LoadDatabase()
		if err != nil {
			return nil, err
		}
		defer db.Close()
		// List all pages from the database
		rows, err := db.Query("SELECT title FROM Pages") // Assuming you have a 'Pages' table with a 'title' column
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var pageLinks []string
		for rows.Next() {
			var title string
			err := rows.Scan(&title)
			if err != nil {
				return nil, err
			}
			pageLinks = append(pageLinks, fmt.Sprintf("<li><a href=\"%s%s\">%s</a></li>", baseURL, title, title))
		}

		bodyHTML := fmt.Sprintf("<h2 class=\"wikih2\">All Pages</h2><ul>\n%s\n</ul>", strings.Join(pageLinks, "\n"))
		safeMenu, err := menu.Load()
		if err != nil {
			return nil, err // Return error if menu file reading fails
		}
		return &Page{
			NavTitle:   config.SiteTitle,
			ThemeColor: arcWikiLogo(),
			CTitle:     "Special:AllPages",
			Title:      "Special:AllPages",
			Body:       template.HTML(bodyHTML),
			Size:       template.HTML(size),
			Menu:       template.HTML(safeMenu),
		}, nil
	} else {

		safeMenu, err := menu.Load()
		if err != nil {
			return nil, err // Return error if menu file reading fails
		}
		return &Page{
			NavTitle:   config.SiteTitle,
			ThemeColor: arcWikiLogo(),
			Title:      "Special:AllCategories",
			Body:       template.HTML("nothing here"),
			Size:       template.HTML(size),
			Menu:       template.HTML(safeMenu),
		}, nil
	}
}
