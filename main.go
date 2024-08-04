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
	"regexp"
	"strings"
	"time"

	"github.com/ArcWiki/ArcWiki/db"
	"github.com/houseme/mobiledetect"
	log "github.com/sirupsen/logrus"

	_ "github.com/mattn/go-sqlite3"
)

const Desktop = "desktop"
const Mobile = "mobile"

func viewHandler(w http.ResponseWriter, r *http.Request, title string, userAgent string) {
	//.stats.getStats()
	//category := r.URL.Path[len("/title/"):]
	if len(r.URL.Path) >= len("/title/") && r.URL.Path[:len("/title/")] == "/title/" {
		// Path starts with "/title/" and has enough characters for slicing
		category := r.URL.Path[len("/title/"):]

		// Process the extracted category

		switch {
		//Uses the prefix Help: making navigating to Help easier
		case strings.HasPrefix(category, "Help:"):
			specialPageName := strings.TrimPrefix(category, "Help:")
			specialPageName = strings.TrimSpace(specialPageName)
			log.Info("Help page accessed:", specialPageName)

			p, err := loadPage("Help-"+specialPageName, userAgent)
			if err != nil {
				log.Error("Error Occured in:", specialPageName)

				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
			renderTemplate(w, "title", p)
			//Gets a list of pages and Random lands on one
		case strings.HasPrefix(category, "Special:Random"):
			log.Info("Random page accessed:", category)
			db, err := db.LoadDatabase()
			if err != nil {
				log.Error("Error Loading Database:", err)

			}
			defer db.Close()

			// Fetch a random page title from the Pages table
			// Fetch a random page title from the Pages table securely
			var title string
			stmt, err := db.Prepare("SELECT title FROM Pages ORDER BY RANDOM() LIMIT 1")
			if err != nil {

				log.Error("Error preparing statement:", err)
				return
			}
			defer stmt.Close() // Close the statement after use

			row := stmt.QueryRow()
			err = row.Scan(&title)
			if err != nil {

				log.Info("No pages found in database")

				return
			}

			// Use the retrieved title securely
			log.Info("Random Page Title:", title) // Or use the title for your purpose

			http.Redirect(w, r, "/title/"+title, http.StatusFound) // Redirect to the randomly selected page

			// Assuming you have a renderTemplate function for rendering the Page struct
			//renderTemplate(w, "page", page, safeMenu)
		// Allows getting to Special pages which are basically functional pages
		case strings.HasPrefix(category, "Special:"):
			specialPageName := strings.TrimPrefix(category, "Special:")
			specialPageName = strings.TrimSpace(specialPageName)
			log.Info("Special page accessed:", specialPageName)

			p, err := loadPageSpecial(specialPageName, userAgent)
			if err != nil {
				log.Error("Error Occurred in:", err)

				//fmt.Println("Error Occured in:", specialPageName)
				//http.Redirect(w, r, "/edit/"+title, http.StatusFound)

			}
			renderTemplate(w, "title", p)
		//possibly for Editing Categories forgotten
		case strings.Contains(category, ":"):
			categoryParts := strings.Split(category, ":")
			categoryName := strings.TrimSpace(categoryParts[1])
			log.Debug("Category: ", categoryName)
			p, err := loadPageCategory(title, categoryName, userAgent)
			if err != nil {
				log.Error("Error Occurred in:", err)

				//http.Redirect(w, r, "/edit/"+title, http.StatusFound)
				return
			}
			renderTemplate(w, "title", p)
		// Normal View Normal View Normal View
		default:
			log.Info("Showing Page: ", title)

			// Load the page for standard title viewing
			p, err := loadPage(title, userAgent)
			if err != nil {
				log.Error("viewHandler: Something weird happened")
				http.Redirect(w, r, "/title/Main_page", http.StatusFound)
				return
			}

			renderTemplate(w, "title", p)

		}
	} else {
		//fmt.Println("hello beautiful world")
		// Load the page for standard title viewing
		p, err := loadPage("Main_page", userAgent)
		if err != nil {
			log.Error("viewHandler: Something weird happened")
			http.Redirect(w, r, "/title/Main_page", http.StatusFound)
			return
		}
		renderTemplate(w, "title", p)
	}

}

// Edit Handler with a switch for editing Categories
func editHandler(w http.ResponseWriter, r *http.Request, title string, userAgent string) {
	updated_at := "Not Available"
	log.Debug(title)
	size := ""
	if userAgent == Desktop {
		size = "<div class=\"col-11 d-none d-sm-block\">"
	} else {
		size = "<div class=\"col-12 d-block d-sm-none\">"
	}
	category := r.URL.Path[len("/title/"):]

	switch {
	case strings.Contains(category, ":"):
		categoryParts := strings.Split(category, ":")
		categoryName := strings.TrimSpace(categoryParts[1])
		log.Debug("Category:", categoryName)

		session, _ := store.Get(r, "cookie-name")
		auth, ok := session.Values["authenticated"].(bool)

		if !ok || !auth {
			//http.Error(w, "Forbidden", http.StatusForbidden)
			http.Redirect(w, r, "/error", http.StatusFound)
			return
		} else {

			ep, err := loadCategoryNoHtml(categoryName, userAgent)

			if err != nil {
				ep = &EditPage{CTitle: categoryName, Title: categoryName, Size: template.HTML(size), UpdatedDate: updated_at}
			}
			renderEditPageTemplate(w, "editCategory", ep)
		}
	default:

		// check our user is logged in
		session, _ := store.Get(r, "cookie-name")
		auth, ok := session.Values["authenticated"].(bool)

		if !ok || !auth {
			//http.Error(w, "Forbidden", http.StatusForbidden)
			http.Redirect(w, r, "/error", http.StatusFound)
			return
		} else {

			ep, err := loadPageNoHtml(title, userAgent)

			if err != nil {
				ep = &EditPage{CTitle: removeUnderscores(title), Title: title, Size: template.HTML(size)}
			}
			renderEditPageTemplate(w, "edit", ep)
		}
	}
}

func saveCatHandler(w http.ResponseWriter, r *http.Request, title string, userAgent string) {
	body := r.FormValue("body")

	p := &Page{Title: title, Body: template.HTML(body)}
	err := p.saveCat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/title/Special:Categories", http.StatusFound)
}

// var validPath = regexp.MustCompile("^/(?:(add|addpage|cat|edit|save|title|Category|Special)/([a-zA-Z0-9]+)|)")
// var validPath = regexp.MustCompile("^/(?:(add|addpage|cat|edit|save|title|Category|Special)/([a-zA-Z0-9_-]+)|)")
var validPath = regexp.MustCompile("^/(?:(search|results|admin|add|addpage|edit|delete|savecat|save|title|login|Category|Special)/([a-zA-Z0-9'_-]+)|)")

//var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string, string)) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		//requestHeaders := r.Header
		detect := mobiledetect.New(r, nil)
		userAgent := ""
		if detect.IsMobile() || detect.IsTablet() {
			//log.Debug("Responsive Mode Activated")
			userAgent = Mobile
		} else {
			userAgent = Desktop
			//log.Debug("Desktop Detected")
		}

		//userAgent := requestHeaders.Get("User-Agent")
		//fmt.Println("passing through")

		m := validPath.FindStringSubmatch(r.URL.Path)

		if m == nil {
			http.NotFound(w, r)
			log.Error("Handler Error")
			return
		}

		fn(w, r, m[2], userAgent)
		//handleSSEUpdates(w, r)
	}
}
func deleteHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the resource type and title using strings.SplitN

	session, _ := store.Get(r, "cookie-name")
	auth, ok := session.Values["authenticated"].(bool)

	if !ok || !auth {
		//http.Error(w, "Forbidden", http.StatusForbidden)
		http.Redirect(w, r, "/error", http.StatusFound)
		return
	} else {
		parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/delete/"), "/", 2)
		if len(parts) != 2 {
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			return
		}
		resourceType := parts[0]
		title := parts[1]

		// Handle deletion based on resource type
		if resourceType == "page" {
			// Handle deletion of a page
			p := &Page{Title: title}
			err := p.deletePage()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/admin/manage", http.StatusFound)
		} else if resourceType == "category" {
			// Handle deletion of a category
			cat := &Category{Title: title}
			err := cat.deleteCategory()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/admin/manage", http.StatusFound)
		} else {
			// Handle invalid resource type
			//http.Error(w, "Invalid resource type", http.StatusBadRequest)
			http.Redirect(w, r, "/error", http.StatusFound)
		}
	}
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	// check our user is logged in
	session, _ := store.Get(r, "cookie-name")
	auth, ok := session.Values["authenticated"].(bool)

	if !ok || !auth {
		http.Redirect(w, r, "/error", http.StatusFound)
		//http.Error(w, "Forbidden", http.StatusForbidden)
		return
	} else {
		detect := mobiledetect.New(r, nil)
		size := ""
		if detect.IsMobile() || detect.IsTablet() {
			//fmt.Println("is either a mobile or tablet")

			size = "<div class=\"col-12 d-block d-sm-none\">"
		} else {
			size = "<div class=\"col-11 d-none d-sm-block\">"
		}

		title := ""
		safeMenu, err := loadMenu()
		if err != nil {
			log.Error("Error Loading Menu:", err)
		}
		// Create an AddPage instance directly (no loading from file)
		ap := &AddPage{NavTitle: config.SiteTitle, ThemeColor: template.HTML(arcWikiLogo()), CTitle: "Add Page", Title: title, Menu: safeMenu, Size: template.HTML(size), UpdatedDate: ""}

		// Populate other fields of ap as needed (e.g., from session data, user input, etc.)

		renderAddPageTemplate(w, "add", ap)
	}
}

// Error page needs to be used
func errorPage(w http.ResponseWriter, r *http.Request) {
	detect := mobiledetect.New(r, nil)
	userAgent := ""
	if detect.IsMobile() || detect.IsTablet() {
		//fmt.Println("is either a mobile or tablet")
		userAgent = Mobile
	} else {
		userAgent = Desktop
	}
	p, err := loadPageSpecial("specialPageName", userAgent)
	if err != nil {
		http.Error(w, "Error loading HTML file", http.StatusInternalServerError)
		return
	}
	renderTemplate(w, "errorPage", p)

}
func dbsql(stater string, args ...interface{}) error {
	db, err := db.LoadDatabase()
	if err != nil {
		log.Error("Error Loading Database:", err)

	}
	defer db.Close() // Ensure database closure

	stmt, err := db.Prepare(stater)
	if err != nil {
		log.Error("Database Error: ", err)
	}
	defer stmt.Close() // Close the prepared statement

	_, err = stmt.Exec(args...) // Execute the statement with provided arguments
	if err != nil {
		log.Error("Database Error: ", err)
	}

	return nil // Indicate successful execution
}

// moved here for ease
var templates = template.Must(template.ParseFiles("templates/search.html", "templates/header.html", "templates/footer.html", "templates/navbar.html", "templates/edit.html", "templates/title.html", "templates/add.html", "templates/login.html", "templates/editCategory.html", "templates/errorPage.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {

	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		log.Error("Error Occured in renderTemplate: ", err)

		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func renderEditPageTemplate(w http.ResponseWriter, tmpl string, ep *EditPage) {
	err := templates.ExecuteTemplate(w, tmpl+".html", ep)
	if err != nil {
		log.Error("Error Occured in renderEditPageTemplate: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderAddPageTemplate(w http.ResponseWriter, tmpl string, ap *AddPage) {
	err := templates.ExecuteTemplate(w, tmpl+".html", ap)
	if err != nil {
		log.Error("Error Occured in renderEditPageTemplate: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type Config struct {
	Admin     []Admin    `json:"admin"`
	SiteTitle string     `json:"siteTitle"`
	TColor    string     `json:"TColor"`
	SecretKey string     `json:"secretKey"`
	Menu      []MenuItem `json:"menu"`
}

type Admin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type MenuItem struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

var config Config

func loadMenu() (template.HTML, error) {
	var links strings.Builder

	for _, menuItem := range config.Menu {
		links.WriteString(fmt.Sprintf("<li><a href=\"%s\">%s</a></li>\n", menuItem.Link, menuItem.Name))
	}

	return template.HTML(links.String()), nil
}

func main() {
	//start log
	log.SetFormatter(&log.TextFormatter{
		DisableColors:   false,
		FullTimestamp:   true,
		TimestampFormat: "15:04:05",
	})

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
	// A common pattern is to re-use fields between logging statements by re-using
	// the logrus.Entry returned from WithFields()

	log.Info("Starting your instance of ArcWiki")

	// log.WithFields(log.Fields{
	// 	"animal": "walrus",
	// 	"size":   10,
	// }).Info("A group of walrus emerges from the ocean")

	// log.WithFields(log.Fields{
	// 	"omg":    true,
	// 	"number": 122,
	// }).Warn("The group's number increased tremendously!")

	// log.WithFields(log.Fields{
	// 	"omg":    true,
	// 	"number": 100,
	// }).Fatal("The ice breaks!")

	db.DBSetup()

	configBytes, err := os.ReadFile("config/config.json")
	if err != nil {

		log.Panic("Error Reading File:", err)

	}

	err = json.Unmarshal(configBytes, &config)
	if err != nil {
		panic(err)
		log.Warn("Error unmarshalling File:", err)
	}

	if os.Getenv("COLOR") != "" {
		config.TColor = os.Getenv("COLOR")
	}
	if os.Getenv("SITENAME") != "" {
		config.SiteTitle = os.Getenv("SITENAME")
	}
	//fmt.Println(config.Admin[0].Username)

	//	log.Info("Starting your instance of ArcWiki called:", config.SiteTitle)

	go func() {
		for {

			if err := updateSubCategoryLinks(); err != nil {
				// Handle error
				log.Error("Error updating subcategories:", err)
				//fmt.Println("Error updating subcategories:", err)
			}
			time.Sleep(60 * time.Second)
		}
	}()

	http.HandleFunc("/", makeHandler(viewHandler))
	http.HandleFunc("/search", makeHandler(SearchHandler))
	http.HandleFunc("/query", QueryHandler)
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/addpage", addPage)
	http.HandleFunc("/delete/", deleteHandler)
	http.HandleFunc("/category/", addCat)
	http.HandleFunc("/savecat/", makeHandler(saveCatHandler))
	http.HandleFunc("/admin/", makeHandler(adminHandler))
	http.HandleFunc("/logout", makeHandler(logout))
	http.HandleFunc("/login", makeHandler(loginFormHandler))
	http.HandleFunc("/loginPost", makeHandler(loginHandler))
	http.HandleFunc("/title/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/error", errorPage)

	fs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	log.Fatal(http.ListenAndServe(":8080", nil))

}
