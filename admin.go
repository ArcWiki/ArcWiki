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

	"github.com/ArcWiki/ArcWiki/db"
	log "github.com/sirupsen/logrus"
)

func adminHandler(w http.ResponseWriter, r *http.Request, title string, userAgent string) {
	baseURL := "/title/"

	switch title {
	case "page":
		managePages(userAgent, baseURL, w)
		return

	case "category":
		manageCategory(userAgent, baseURL, w)
		return

	default:
		// Mobile/Desktop wrapper size
		size := ""
		if userAgent == Desktop {
			size = "<div class=\"col-11 d-none d-sm-block\">"
		} else {
			size = "<div class=\"col-12 d-block d-sm-none\">"
		}

		// Load menu
		safeMenu, err := loadMenu()
		if err != nil {
			log.Error("error loading menu")
		}

		p := Page{
			NavTitle:     config.SiteTitle,
			ThemeColor:   template.HTML(arcWikiLogo()),
			CTitle:       "Admin Panel",
			Title:        "admin",
			Size:         template.HTML(size),
			Menu:         safeMenu,
			CategoryLink: nil, // Set if needed
		}

		renderTemplate(w, "admin", &p)
	}
}

func manageCategory(userAgent string, baseURL string, w http.ResponseWriter) {
	size := ""
	if userAgent == Desktop {
		size = "<div class=\"col-11 d-none d-sm-block\">"
	} else {
		size = "<div class=\"col-12 d-block d-sm-none\">"
	}

	db, err := db.LoadDatabase()
	if err != nil {

		log.Error("Database Error", err)
		return

	}
	defer db.Close()

	rows, err := db.Query("SELECT title FROM Categories")
	if err != nil {

		log.Error("Database Error", err)
		return
	}
	defer rows.Close()

	var pageLinks []string
	for rows.Next() {
		var title string
		err := rows.Scan(&title)
		if err != nil {

			log.Error("Database Error", err)
			return

		}
		pageLinks = append(pageLinks, fmt.Sprintf("<li><a href=\"%s%s\">%s</a> <a href=\"%s\"> Edit Category</a> <a href=\"%s\"> Delete Category</a></li>", baseURL, "Category:"+title, title, "/edit/Category:"+title, "/delete/category/"+title))
	}

	bodyHTML := fmt.Sprintf("<h2 class=\"wikih2\">Manage Pages</h2><ul>\n%s\n</ul>", strings.Join(pageLinks, "\n"))

	safeBodyHTML := template.HTML(bodyHTML)

	safeMenu, err := loadMenu()
	if err != nil {
		log.Info("error loading menu")
	}

	p := Page{NavTitle: config.SiteTitle, ThemeColor: template.HTML(arcWikiLogo()), CTitle: "Manage Category", Title: "admin", Body: safeBodyHTML, Size: template.HTML(size), Menu: safeMenu}

	renderTemplate(w, "title", &p)
}

func managePages(userAgent string, baseURL string, w http.ResponseWriter) {
	size := ""
	if userAgent == Desktop {
		size = "<div class=\"col-11 d-none d-sm-block\">"
	} else {
		size = "<div class=\"col-12 d-block d-sm-none\">"
	}

	db, err := db.LoadDatabase()
	if err != nil {
		log.Error("Error loading database", err)
		return

	}
	defer db.Close()

	rows, err := db.Query("SELECT title FROM Pages")
	if err != nil {

		log.Error("Database Error", err)
		return
	}
	defer rows.Close()

	var pageLinks []string
	for rows.Next() {
		var title string
		err := rows.Scan(&title)
		if err != nil {
			log.Error("Database Error", err)
			return

		}
		pageLinks = append(pageLinks, fmt.Sprintf("<li><a href=\"%s%s\">%s</a> <a href=\"%s\"> Edit Page</a> <a href=\"%s\"> Delete Page</a></li>", baseURL, title, title, "/edit/"+title, "/delete/page/"+title))
	}

	bodyHTML := fmt.Sprintf("<h2 class=\"wikih2\">Manage Pages</h2><ul>\n%s\n</ul>", strings.Join(pageLinks, "\n"))

	safeBodyHTML := template.HTML(bodyHTML)

	safeMenu, err := loadMenu()
	if err != nil {
		log.Error("error loading menu")
	}

	p := Page{NavTitle: config.SiteTitle, ThemeColor: template.HTML(arcWikiLogo()), CTitle: "Manage Pages", Title: "admin", Body: safeBodyHTML, Size: template.HTML(size), Menu: safeMenu}

	renderTemplate(w, "title", &p)
}
