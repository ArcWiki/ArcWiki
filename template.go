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
)

var templates = template.Must(template.ParseFiles("templates/header.html", "templates/footer.html", "templates/navbar.html", "templates/edit.html", "templates/title.html", "templates/add.html", "templates/login.html", "templates/editCategory.html", "templates/errorPage.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {

	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		fmt.Println("Error Occured in renderTemplate " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func renderEditPageTemplate(w http.ResponseWriter, tmpl string, ep *EditPage) {
	err := templates.ExecuteTemplate(w, tmpl+".html", ep)
	if err != nil {
		fmt.Println("Error Occured in renderEditPageTemplate " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderAddPageTemplate(w http.ResponseWriter, tmpl string, ap *AddPage) {
	err := templates.ExecuteTemplate(w, tmpl+".html", ap)
	if err != nil {
		fmt.Println("Error Occured in renderAddPageTemplate " + err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
