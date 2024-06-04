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
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

func adminHandler(w http.ResponseWriter, r *http.Request, title string, userAgent string) {
	baseURL := "/title/"
	if title == "page" {
		managePages(userAgent, baseURL, w) // Pass only the body string

	} else if title == "category" {
		manageCategory(userAgent, baseURL, w)
	} else {
		size := ""
		if userAgent == "desktop" {
			size = "<div class=\"col-11 d-none d-sm-block\">"
		} else {
			size = "<div class=\"col-12 d-block d-sm-none\">"
		}

		bodyMark :=
			`
	<a class="btn btn-sm btn-outline-secondary" href="/add"> Add Page </a><br />
	<a class="btn btn-sm btn-outline-secondary" href="/admin/page"> Manage Pages </a><br />
	<a class="btn btn-sm btn-outline-secondary" href="/admin/category"> Manage Category </a><br />
	<a class="btn btn-sm btn-outline-secondary" href="/logout"> Logout </a>
	`
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
		safeMenu, err := loadMenu()
		if err != nil {
			fmt.Println("error loading menu")
		}

		p := Page{NavTitle: config.SiteTitle, CTitle: "Admin panel", Title: "admin", Body: safeBodyHTML, Size: template.HTML(size), Menu: safeMenu, CategoryLink: categoryLink}

		// Assuming renderTemplate accepts a string for body content:
		renderTemplate(w, "title", &p) // Pass only the body string
	}

}
func manageCategory(userAgent string, baseURL string, w http.ResponseWriter) {
	size := ""
	if userAgent == "desktop" {
		size = "<div class=\"col-11 d-none d-sm-block\">"
	} else {
		size = "<div class=\"col-12 d-block d-sm-none\">"
	}

	db, err := loadDatabase()
	if err != nil {

	}
	defer db.Close()

	rows, err := db.Query("SELECT title FROM Categories")
	if err != nil {

	}
	defer rows.Close()

	var pageLinks []string
	for rows.Next() {
		var title string
		err := rows.Scan(&title)
		if err != nil {

		}
		pageLinks = append(pageLinks, fmt.Sprintf("<li><a href=\"%s%s\">%s</a> <a href=\"%s\"> Edit Category</a> <a href=\"%s\"> Delete Category</a></li>", baseURL, "Category:"+title, title, "/edit/Category:"+title, "/delete/category/"+title))
	}

	bodyHTML := fmt.Sprintf("<h2 class=\"wikih2\">Manage Pages</h2><ul>\n%s\n</ul>", strings.Join(pageLinks, "\n"))

	safeBodyHTML := template.HTML(bodyHTML)

	safeMenu, err := loadMenu()
	if err != nil {
		fmt.Println("error loading menu")
	}

	p := Page{NavTitle: config.SiteTitle, CTitle: "Manage Category", Title: "admin", Body: safeBodyHTML, Size: template.HTML(size), Menu: safeMenu}

	renderTemplate(w, "title", &p)
}

func managePages(userAgent string, baseURL string, w http.ResponseWriter) {
	size := ""
	if userAgent == "desktop" {
		size = "<div class=\"col-11 d-none d-sm-block\">"
	} else {
		size = "<div class=\"col-12 d-block d-sm-none\">"
	}

	db, err := loadDatabase()
	if err != nil {

	}
	defer db.Close()

	rows, err := db.Query("SELECT title FROM Pages")
	if err != nil {

	}
	defer rows.Close()

	var pageLinks []string
	for rows.Next() {
		var title string
		err := rows.Scan(&title)
		if err != nil {

		}
		pageLinks = append(pageLinks, fmt.Sprintf("<li><a href=\"%s%s\">%s</a> <a href=\"%s\"> Edit Page</a> <a href=\"%s\"> Delete Page</a></li>", baseURL, title, title, "/edit/"+title, "/delete/page/"+title))
	}

	bodyHTML := fmt.Sprintf("<h2 class=\"wikih2\">Manage Pages</h2><ul>\n%s\n</ul>", strings.Join(pageLinks, "\n"))

	safeBodyHTML := template.HTML(bodyHTML)

	safeMenu, err := loadMenu()
	if err != nil {
		fmt.Println("error loading menu")
	}

	p := Page{NavTitle: config.SiteTitle, CTitle: "Manage Pages", Title: "admin", Body: safeBodyHTML, Size: template.HTML(size), Menu: safeMenu}

	renderTemplate(w, "title", &p)
}
