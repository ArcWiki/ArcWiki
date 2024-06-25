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
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/ArcWiki/ArcWiki/db"

	"github.com/gomarkdown/markdown"
)

type AddPage struct {
	ThemeColor  template.HTML
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
	ID           int
	ThemeColor   template.HTML
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
	ThemeColor  template.HTML
	NavTitle    string
	CTitle      string
	Title       string
	Body        template.HTML
	Menu        template.HTML
	Size        template.HTML
	UpdatedDate string
}

func (p *Page) save() error {
	fmt.Println("the page saved was called " + canonicalizeTitle(p.Title))
	db, err := db.LoadDatabase()
	if err != nil {
		return fmt.Errorf("error loading database: %w", err) // Return a descriptive error
	}
	defer db.Close()

	tx, err := db.Begin() // Start transaction
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if err != nil { // Rollback on any error
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare("UPDATE Pages SET title = ?, body = ?, updated_at = CURRENT_TIMESTAMP WHERE title = ?")
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(canonicalizeTitle(p.Title), string(p.Body), canonicalizeTitle(p.Title)) // Ignore RowsAffected
	if err != nil {
		return fmt.Errorf("error executing update: %w", err)
	}

	// Prepare a list of category IDs to insert based on match[1]
	var categoryIDsToInsert []int
	regex := regexp.MustCompile(`\[Category:([^\]|]*)\]`)
	matches := regex.FindAllStringSubmatch(string(p.Body), -1) // Find all matches

	// Check if the page exists before fetching ID
	var pageID int
	// Check if the page exists before fetching ID
	row := tx.QueryRow("SELECT id FROM Pages WHERE title = ?", canonicalizeTitle(p.Title))
	err = row.Scan(&pageID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("page with title", canonicalizeTitle(p.Title)+" not found") // Informative error
		}
		fmt.Println("error checking for existing page:", err)
	}
	_, err = tx.Exec("DELETE FROM CategoryPages WHERE page_id = ?", pageID)
	if err != nil {
		fmt.Println("Error deleting existing category links:", err)
		// Consider returning an error or logging the error and continuing
	}
	for _, matchedCategory := range matches {
		var categoryID int
		err := tx.QueryRow("SELECT id FROM Categories WHERE title = ?", matchedCategory[1]).Scan(&categoryID)
		if err != nil {
			fmt.Println("Error fetching category ID:", err)
			continue // Skip to next category if error occurs
		}
		categoryIDsToInsert = append(categoryIDsToInsert, categoryID)
	}

	// Batch insert new category links (adjusted for current page only)
	for _, categoryID := range categoryIDsToInsert {

		_, err = tx.Exec("INSERT INTO CategoryPages (page_id, category_id) VALUES (?, ?)", pageID, categoryID)
		fmt.Println("Inserting Category links", pageID, categoryID)
		if err != nil {
			fmt.Println(err)
			_ = tx.Rollback() // Explicit rollback on error

			fmt.Println("error inserting category link:", err)
		}
	}

	// Commit the transaction only once after successful insertions
	err = tx.Commit()
	if err != nil {
		fmt.Println("error committing transaction:", err)
	}

	fmt.Println("Updated page with title:", canonicalizeTitle(p.Title)) // Clearer message
	return nil
}

func addPage(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	title := r.FormValue("title")
	if title != "index" {
		body := r.FormValue("body")
		// We Fix make the category links straight away more dev here
		regex := regexp.MustCompile(`\[Category:([^\]|]*)\]`)
		matches := regex.FindAllStringSubmatch(body, -1) // Find all matches

		freshTitle := canonicalizeTitle(title)

		db, err := db.LoadDatabase()
		if err != nil {
			fmt.Println("Database Error: " + err.Error())
			return // Handle error
		}

		stmt := `INSERT INTO Pages (title, body, user_id, created_at, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP); SELECT last_insert_rowid();`

		tx, err := db.Begin()
		if err != nil {
			fmt.Println(err)
			return // Handle error
		}
		defer tx.Rollback() // Rollback if any error occurs

		result, err := tx.Exec(stmt, freshTitle, body, 1)
		if err != nil {
			fmt.Println(err)
			_ = tx.Rollback() // rollback if error occurs
			return            // Handle error
		}

		var pageID int64
		pageID, err = result.LastInsertId()
		fmt.Println("page id is this : ", pageID)
		if err != nil {
			fmt.Println(err)
			_ = tx.Rollback() // Explicitly rollback if error occurs
			return            // Handle error
		}

		// Prepare a list of category IDs to insert based on match[1]
		var categoryIDsToInsert []int
		for _, matchedCategory := range matches {
			var categoryID int
			err := tx.QueryRow("SELECT id FROM Categories WHERE title = ?", matchedCategory[1]).Scan(&categoryID)
			if err != nil { // Handle potential error fetching category ID
				fmt.Println("Error fetching category ID:", err)
				continue // Skip to next category if error occurs
			}
			categoryIDsToInsert = append(categoryIDsToInsert, categoryID)
		}

		// Batch insert new category links (adjusted for current page only)
		for _, categoryID := range categoryIDsToInsert {
			_, err = tx.Exec("INSERT INTO CategoryPages (page_id, category_id) VALUES (?, ?)", pageID, categoryID)
			fmt.Println("Inserting Category links" + string(pageID) + " " + string(categoryID))
			if err != nil {
				fmt.Println(err)
				_ = tx.Rollback() // Explicitly
				return            // Handle error
			}
		}

		// Commit the transaction only once after successful insertions
		err = tx.Commit()
		if err != nil {
			fmt.Println(err)
			return // Handle error
		}

		http.Redirect(w, r, "/title/"+freshTitle, http.StatusFound)
	} else {
		fmt.Println("cannot be index don't be silly")
		http.Redirect(w, r, "/title/index", http.StatusFound)
	}
}

func (p *Page) deletePage() error {
	db, err := db.LoadDatabase()
	if err != nil {
		return fmt.Errorf("error loading database: %w", err) // Return a descriptive error
	}
	defer db.Close()

	tx, err := db.Begin() // Start transaction
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if err != nil { // Rollback on any error
			_ = tx.Rollback()
		}
	}()
	var pageID int
	// Check if the page exists before deleting
	row := tx.QueryRow("SELECT id FROM Pages WHERE title = ?", canonicalizeTitle(p.Title))
	// Placeholder variable to eliminate unnecessary scan
	err = row.Scan(&pageID)

	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("page with title", canonicalizeTitle(p.Title)+" not found") // Informative error
		}
		fmt.Println("error checking for existing page: %w", err)
	}

	// Delete category links first (assuming foreign key constraints exist)
	_, err = tx.Exec("DELETE FROM CategoryPages WHERE page_id = ?", pageID) // Use title for efficiency (assuming unique constraint)
	if err != nil {
		fmt.Println("Error deleting category links:", err)
		// Consider logging the error and continuing with page deletion (optional)
	}

	// Delete the page
	result, err := tx.Exec("DELETE FROM Pages WHERE title = ?", canonicalizeTitle(p.Title))
	if err != nil {
		fmt.Println("error deleting page:", err)
	}

	rowsDeleted, err := result.RowsAffected()
	if err != nil {
		fmt.Println("error checking rows affected:", err)
	}

	if rowsDeleted > 0 {
		fmt.Println("Deleted page with title:", canonicalizeTitle(p.Title))
	} else {
		fmt.Println("No page found with title:", canonicalizeTitle(p.Title)) // May indicate a race condition
	}

	err = tx.Commit() // Commit the transaction
	if err != nil {
		fmt.Println("error committing transaction:", err)
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
	noLinks := removeCategoryLinks(happyhtml)
	//fmt.Println(noLinks)
	perfecthtml := parseWikiText(noLinks)

	internalLinks := convertLinksToAnchors(perfecthtml)
	safeBodyHTML := template.HTML(internalLinks)
	footer := "This page was last modified on " + formatDateTime(updated_at)

	//need to double check this as I'm not certain why this is
	if err == nil { // Page found in database
		// ... (existing code for markdown parsing and HTML generation)
		return &Page{NavTitle: config.SiteTitle, ThemeColor: template.HTML(arcWikiLogo()), CTitle: removeUnderscores(title), Title: title, Body: safeBodyHTML, Size: template.HTML(size), Menu: safeMenu, CategoryLink: categoryLink, UpdatedDate: footer}, nil
	} else if err != sql.ErrNoRows { // Handle other SQLite errors
		return nil, err
	}

	return &Page{NavTitle: config.SiteTitle, ThemeColor: template.HTML(arcWikiLogo()), CTitle: removeUnderscores(title), Title: title, Body: safeBodyHTML, Size: template.HTML(size), Menu: safeMenu, UpdatedDate: footer}, nil
	//return nil, fmt.Errorf("File not found: %s.txt", title) // File not found in any folder
}

// Loads page with no html applied useful for editing markdown in the edit view
func loadPageNoHtml(title string, userAgent string) (*EditPage, error) {
	size := ""

	safeMenu, err := loadMenu()
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
	return &EditPage{NavTitle: config.SiteTitle, ThemeColor: template.HTML(arcWikiLogo()), CTitle: removeUnderscores(title), Title: title, Body: template.HTML(body), Menu: template.HTML(safeMenu), Size: template.HTML(size), UpdatedDate: footer}, nil
}
func loadPageSpecial(categoryName string, userAgent string) (*Page, error) {
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
		safeMenu, err := loadMenu()
		if err != nil {
			return nil, err // Return error if menu file reading fails
		}
		return &Page{

			NavTitle:   config.SiteTitle,
			ThemeColor: template.HTML(arcWikiLogo()),
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
		safeMenu, err := loadMenu()
		if err != nil {
			return nil, err // Return error if menu file reading fails
		}
		return &Page{
			NavTitle:   config.SiteTitle,
			ThemeColor: template.HTML(arcWikiLogo()),
			CTitle:     "Special:AllPages",
			Title:      "Special:AllPages",
			Body:       template.HTML(bodyHTML),
			Size:       template.HTML(size),
			Menu:       template.HTML(safeMenu),
		}, nil
	} else {

		safeMenu, err := loadMenu()
		if err != nil {
			return nil, err // Return error if menu file reading fails
		}
		return &Page{
			NavTitle:   config.SiteTitle,
			ThemeColor: template.HTML(arcWikiLogo()),
			Title:      "Special:AllCategories",
			Body:       template.HTML("nothing here"),
			Size:       template.HTML(size),
			Menu:       template.HTML(safeMenu),
		}, nil
	}
}
