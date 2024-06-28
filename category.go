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
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/ArcWiki/ArcWiki/db"
)

type Category struct {
	Title string
	Body  string
	//user_id int
	Size template.HTML
	Menu template.HTML
	ID   int
}

func (p *Category) deleteCategory() error {
	db, err := db.LoadDatabase()

	if err != nil {
		panic(err)
	}
	defer db.Close()

	stmt, err := db.Prepare("DELETE FROM Categories WHERE title = ?")
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
func (p *Page) saveCat() error {
	fmt.Println("titlehere.." + p.Title + ".. " + string(p.Body))

	err := dbsql("UPDATE Categories SET body = ? WHERE title = ?", string(p.Body), p.Title)
	if err != nil {
		fmt.Println("Page Save Error: " + err.Error())

	}

	return nil

}
func getCategoryIDByName(categoryName string) (int, error) {
	//new
	db, err := db.LoadDatabase()
	if err != nil {
		return 0, err
	}
	defer db.Close()

	stmt, err := db.Prepare("SELECT id FROM Categories WHERE title = ?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(categoryName)
	var categoryID int
	err = row.Scan(&categoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // Category not found
		}
		return 0, err
	}

	return categoryID, nil
}

func addCat(w http.ResponseWriter, r *http.Request) {
	//new
	categoryName := r.URL.Path[len("/category/"):]
	db, err := db.LoadDatabase()
	if err != nil {
		fmt.Println("Database Error: " + err.Error())

	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO Categories (title, body, user_id) VALUES (?, ?, ?)")
	if err != nil {
		fmt.Println("Database Error: " + err.Error())

	}
	defer stmt.Close()

	res, err := stmt.Exec(canonicalizeTitle(categoryName), "", 1)
	if err != nil {
		fmt.Println("Database Error: " + err.Error())

	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		// Handle error
	} else if rowsAffected != 1 {
		fmt.Println("Unexpected number of rows affected:", rowsAffected)
	} else {
		fmt.Println("Category inserted successfully!")
		http.Redirect(w, r, "/title/Special:Categories", http.StatusFound)
	}
	fmt.Println(categoryName)
}
func checkCategoryExistence(categoryName string) bool {

	db, err := db.LoadDatabase()
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Define the category name to check

	// Prepare the SQL query
	stmt, err := db.Prepare("SELECT EXISTS(SELECT 1 FROM Categories WHERE title = ?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	// Execute the query with the category name
	var exists bool
	err = stmt.QueryRow(categoryName).Scan(&exists)
	if err != nil {
		panic(err)
	}
	return exists

}

// displays categories and sub-categories on the Category:somename page
func loadPageCategory(title string, categoryName string, userAgent string) (*Page, error) {
	size := ""
	if userAgent == Desktop {
		size = "<div class=\"col-11 d-none d-sm-block\">"
	} else {
		size = "<div class=\"col-12 d-block d-sm-none\">"
	}
	safeMenu, err := loadMenu() // Replace with your menu loading logic
	if err != nil {
		return nil, err
	}

	// Gather matching pages
	matchingPages := findPagesInCategory(categoryName)
	matchingSubCatPages := loadLinksFromSubCategoryFile(categoryName)
	fmt.Println(matchingSubCatPages)
	// Format bodyHTML based on matching pages
	categories := formatPageList(matchingPages)
	subcategories := formatSubCatList(matchingSubCatPages)
	if checkCategoryExistence(categoryName) {
		if subcategories != "" {
			subcategories = "<h2 class=\"wikih2\"> Subcategories </h2>" + subcategories
		}
		if categories != "" {
			categories = "<h2 class=\"wikih2\"> Pages in category </h2>" + categories
		} else {
			categories = "<p>This category currently contains no pages or media. </p>"
		}
	} else {
		categories = "<a style=\"color:red\" href=\"/category/" + categoryName + "\"> Add This Category</a>"

	}
	return &Page{
		NavTitle:   config.SiteTitle,
		ThemeColor: template.HTML(arcWikiLogo()),
		CTitle:     removeUnderscores(title + ":" + categoryName),
		Title:      title + ":" + categoryName,
		Body:       template.HTML(subcategories + categories),
		Size:       template.HTML(size), // Set Size to the number of matching pages
		Menu:       template.HTML(safeMenu),
	}, nil
}

// looks for a file using the same name as the current category loaded prefixed with sub-category name it then generates the the Subcategory pages for this
func loadLinksFromSubCategoryFile(categoryName string) []string {
	db, err := db.LoadDatabase()
	if err != nil {
		return nil
	}
	defer db.Close()

	// Get the category ID
	categoryID, err := getCategoryIDByName(categoryName)
	if err != nil {
		return nil
	}
	fmt.Println(categoryID)

	// Retrieve page paths from CategoryPages
	rows, err := db.Query("SELECT Categories.title FROM SubCategoryPages JOIN Categories ON Categories.id = SubCategoryPages.category_id WHERE SubCategoryPages.subcategory_id = ?", categoryID)

	if err != nil {
		return nil
	}
	defer rows.Close()

	var matchingPages []string
	for rows.Next() {
		var path string
		err := rows.Scan(&path)
		if err != nil {
			return nil
		}
		matchingPages = append(matchingPages, path)
	}

	return matchingPages
}
func findPagesInCategory(categoryName string) []string {
	db, err := db.LoadDatabase()
	if err != nil {
		return nil
	}
	defer db.Close()

	// Get the category ID
	categoryID, err := getCategoryIDByName(categoryName)
	if err != nil {
		return nil
	}

	// Retrieve page paths from CategoryPages
	rows, err := db.Query("SELECT Pages.title FROM Pages JOIN CategoryPages ON Pages.id = CategoryPages.page_id WHERE CategoryPages.category_id = ?", categoryID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var matchingPages []string
	for rows.Next() {
		var path string
		err := rows.Scan(&path)
		if err != nil {
			return nil
		}
		matchingPages = append(matchingPages, path)
	}

	return matchingPages
}

// Formats the links for displaying on web
func formatPageList(pages []string) string {
	baseURL := "/title/"
	// Construct the HTML content for the page list
	// (Implement your desired formatting here)
	linksByLetter := make(map[rune][]string)
	for _, page := range pages {

		linkTitle := filepath.Base(page)
		linkTitleWithoutExt := strings.TrimSuffix(linkTitle, ".txt")
		firstLetter, _ := utf8.DecodeRuneInString(linkTitleWithoutExt)
		firstLetter = unicode.ToUpper(firstLetter)
		// Help title changed to Help:somepage
		if strings.HasPrefix(linkTitleWithoutExt, "Help-") {
			linkTitleWithoutExt = "Help:" + linkTitleWithoutExt[5:]
		}
		//fmt.Println(linkTitle)
		linksByLetter[firstLetter] = append(linksByLetter[firstLetter], fmt.Sprintf("<li><a href=\"%s%s\">%s</a></li>", baseURL, linkTitleWithoutExt, linkTitleWithoutExt))

	}
	keys := make([]string, 0, len(linksByLetter))
	for k := range linksByLetter {
		keys = append(keys, string(k))
	}
	sort.Strings(keys) // Sort keys alphabetically
	var html string
	//var html string
	for _, letter := range keys {
		linkList := linksByLetter[rune(letter[0])] // Access list using converted string
		// ... (rest of your code for generating HTML list)
		html += fmt.Sprintf("<h3>%s</h3><ul>\n%s\n</ul>", letter, strings.Join(linkList, "\n"))
	}
	// for letter, linkList := range linksByLetter {

	// 	html += fmt.Sprintf("<h3>%c</h3><ul>\n%s\n</ul>", letter, strings.Join(linkList, "\n"))
	// }

	return html
}
func formatSubCatList(pages []string) string {
	baseURL := "/title/"
	// Construct the HTML content for the page list

	linksByLetter := make(map[rune][]string)
	for _, page := range pages {

		linkTitle := filepath.Base(page)
		linkTitleWithoutExt := strings.TrimSuffix(linkTitle, ".txt")
		firstLetter, _ := utf8.DecodeRuneInString(linkTitleWithoutExt)
		firstLetter = unicode.ToUpper(firstLetter)
		if strings.HasPrefix(linkTitleWithoutExt, "help-") {
			linkTitleWithoutExt = "Help:" + linkTitleWithoutExt[5:]
		}
		linksByLetter[firstLetter] = append(linksByLetter[firstLetter], fmt.Sprintf("<li><a href=\"%sCategory:%s\">%s</a></li>", baseURL, linkTitleWithoutExt, linkTitleWithoutExt))
	}
	keys := make([]string, 0, len(linksByLetter))
	for k := range linksByLetter {
		keys = append(keys, string(k))
	}
	sort.Strings(keys) // Sort keys alphabetically
	var html string
	//var html string
	for _, letter := range keys {
		linkList := linksByLetter[rune(letter[0])] // Access list using converted string
		// ... (rest of your code for generating HTML list)
		html += fmt.Sprintf("<h3>%s</h3><ul>\n%s\n</ul>", letter, strings.Join(linkList, "\n"))
	}
	// var html string
	// for letter, linkList := range linksByLetter {
	// 	html += fmt.Sprintf("<h3>%c</h3><ul>\n%s\n</ul>", letter, strings.Join(linkList, "\n"))
	// }
	return html
}
func loadCategoryNoHtml(title string, userAgent string) (*EditPage, error) {
	size := ""
	safeMenu, err := loadMenu()
	if err != nil {
		return nil, err
	}
	if userAgent == Desktop {
		size = "<div class=\"col-11 d-none d-sm-block\">"
	} else {
		size = "<div class=\"col-12 d-block d-sm-none\">"
	}
	db, err := db.LoadDatabase()
	if err != nil {
		return nil, err
	}
	stmt, err := db.Prepare("SELECT title, body FROM Categories WHERE title = ?")
	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(title)
	defer db.Close()   // Close the database connection
	defer stmt.Close() // Close the prepared statement

	var body string
	err = row.Scan(&title, &body)
	if err != nil {
		return nil, err
	}

	return &EditPage{NavTitle: config.SiteTitle, ThemeColor: template.HTML(arcWikiLogo()), CTitle: removeUnderscores(title), Title: title, Body: template.HTML(body), Menu: template.HTML(safeMenu), Size: template.HTML(size)}, nil
}
