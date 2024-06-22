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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ArcWiki/ArcWiki/db"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func arcWikiLogo() string {
	return `
	<nav class="navbar navbar-expand-lg navbar-dark bg-dark" style="border-bottom-color:` + config.TColor + `;
	border-width: medium medium 5px;
	border-style: none none solid;">
	  <div class="container-fluid">
		<a class="navbar-brand" href="#">
		  <svg width="30" height="24" viewBox="0 0 20.092125 32.433891" version="1.1" id="svg5"
			xmlns="http://www.w3.org/2000/svg" xmlns:svg="http://www.w3.org/2000/svg">
			<defs id="defs2" />
			<g id="layer1" transform="translate(-6.135872,-88.557874)">
			  <path style="fill:` + config.TColor + `;fill-opacity:1;stroke-width:22.5425"
				d="m 10.418081,120.99177 4.031487,-16.99612 -8.3136961,-0.1154 11.9245121,-9.624479 8.075561,-5.697896 -6.071881,15.278315 6.163934,-0.0705 z"
				id="path423" />
			</g>
		  </svg>
	
	`
}

// the foundation for including iconify
func iconifyLogo() string {

	return `
	<nav class="navbar navbar-expand-lg navbar-dark bg-dark" style="border-bottom-color:` + config.TColor + `;
	border-width: medium medium 5px;
	border-style: none none solid;">
	  <div class="container-fluid">
		<a class="navbar-brand" href="#">
		  <iconify-icon icon="formkit:bitcoin" width="14.9" height="24" style="display: inline-block; vertical-align: middle;"></iconify-icon>
		  `
}

func formatDateTime(t time.Time) string {
	return t.Format("2 January 2006, at 15:04")
}

func updateSubCategoryLinks() error {
	fmt.Println("checking for subcategory links...")
	db, err := db.LoadDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	// Fetch all pages and their content
	rows, err := db.Query("SELECT id, body FROM Categories")
	if err != nil {
		return err
	}
	defer rows.Close()

	var pagesToUpdate []struct {
		pageID     int
		categories []string
	}

	for rows.Next() {
		var pageID int
		var body string
		if err := rows.Scan(&pageID, &body); err != nil {
			return err
		}

		// Extract categories from content
		//re := regexp.MustCompile("\\[Category:(.*)\\]")
		re := regexp.MustCompile(`\[Category:(.*)\]`)

		matches := re.FindAllStringSubmatch(string(body), -1)

		if len(matches) > 0 {
			categories := []string{}
			if len(matches[0]) > 1 {
				categories = matches[0][1:] // Extract category from first match
			}
			pagesToUpdate = append(pagesToUpdate, struct {
				pageID     int
				categories []string
			}{pageID, categories})
		}

	}

	// Perform batch operations in a single transaction
	tx, err := db.Begin()
	if err != nil {

		return err
	}
	defer tx.Rollback() // Rollback if any error occurs

	for _, page := range pagesToUpdate {

		// Batch delete existing links for the page
		fmt.Println("debuger")
		fmt.Println("Cat pageID:", page.pageID)
		_, err = tx.Exec("DELETE FROM SubCategoryPages WHERE subcategory_id = ?", page.pageID)
		if err != nil {
			return err
		}

		// Batch insert new category links
		for _, category := range page.categories {
			categoryName := category

			var categoryID int
			err := tx.QueryRow("SELECT id FROM Categories WHERE title = ?", categoryName).Scan(&categoryID)
			if err == nil {
				_, err = tx.Exec("INSERT INTO SubCategoryPages (subcategory_id, category_id) VALUES (?, ?)", page.pageID, categoryID)
				if err != nil {
					return err
				}
			} else {
				// Handle category not found (e.g., log error, create category, or skip)
				fmt.Println("Category not found in Categories table:", categoryName)
			}
		}
	}

	// Commit the transaction if no errors occurred
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
func updateCategoryLinks() error {
	fmt.Println("checking for category links...")
	db, err := db.LoadDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	// Fetch all pages and their content
	rows, err := db.Query("SELECT id, body FROM pages")
	if err != nil {
		return err
	}
	defer rows.Close()

	var pagesToUpdate []struct {
		pageID     int
		categories []string
	}

	for rows.Next() {
		var pageID int
		var body string
		if err := rows.Scan(&pageID, &body); err != nil {
			return err
		}

		// Extract categories from content
		//re := regexp.MustCompile("\\[Category:(.*)\\]")
		re := regexp.MustCompile(`\[Category:(.*)\]`)

		matches := re.FindAllStringSubmatch(string(body), -1)

		if len(matches) > 0 {
			categories := []string{}
			if len(matches[0]) > 1 {
				categories = matches[0][1:] // Extract category from first match
			}
			pagesToUpdate = append(pagesToUpdate, struct {
				pageID     int
				categories []string
			}{pageID, categories})
		}

	}

	// Perform batch operations in a single transaction
	tx, err := db.Begin()
	if err != nil {

		return err
	}
	defer tx.Rollback() // Rollback if any error occurs

	for _, page := range pagesToUpdate {

		// Batch delete existing links for the page
		////fmt.Println("debugger")
		////fmt.Println("Struct pageID:", page.pageID)
		_, err = tx.Exec("DELETE FROM CategoryPages WHERE page_id = ?", page.pageID)
		if err != nil {
			return err
		}

		// Batch insert new category links
		for _, category := range page.categories {
			categoryName := category

			var categoryID int
			err := tx.QueryRow("SELECT id FROM Categories WHERE title = ?", categoryName).Scan(&categoryID)
			if err == nil {
				_, err = tx.Exec("INSERT INTO CategoryPages (page_id, category_id) VALUES (?, ?)", page.pageID, categoryID)
				if err != nil {
					return err
				}
			} else {
				// Handle category not found (e.g., log error, create category, or skip)
				fmt.Println("Category not found in Categories table:", categoryName)
			}
		}
	}

	// Commit the transaction if no errors occurred
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func canonicalizeTitle(text string) string {
	// Language-specific processing (optional)
	// - If handling non-Latin scripts, use Transliterate() (or similar) from a library
	// - Adjust title case conversion based on language if necessary

	// Lowercase and normalize whitespace
	//title = strings.ToLower(strings.Trim(title, " \t\n"))
	caser := cases.Title(language.English)
	title := caser.String(text)
	// Handle '&' based on context
	title = regexp.MustCompile(`(?i)(&\w+)`).ReplaceAllStringFunc(title, func(match string) string {
		word := strings.TrimLeft(match, "&")
		if isMinorWord(word) || isPreposition(word) { // Check minor words and prepositions
			return "&" + word
		}
		return "And" + strings.ToLower(word) // Replace for other words
	})

	// Handle apostrophes within words
	// Replace leading and trailing apostrophes with underscores
	if title[0] == '\'' {
		title = "" + title[1:]
	}
	if title[len(title)-1] == '\'' {
		title = title[:len(title)-1] + ""
	}

	// Replace spaces with underscores
	title = strings.ReplaceAll(title, " ", "_")

	// Remove other punctuation (except hyphens)
	title = regexp.MustCompile(`[^\w\-'_]+`).ReplaceAllString(title, "")

	// Convert non-alphanumeric characters to underscores (optional)
	// ReplaceAllString(title, unicode.ReplacementChar, "_")

	// Title case conversion

	return title
}

func isMinorWord(word string) bool {
	minorWords := map[string]bool{
		"a": true, "an": true, "and": true, "as": true,
		"at": true, "be": true, "by": true, "but": true,
		"for": true, "from": true, "he": true, "her": true,
		"his": true, "in": true, "it": true, "its": true,
		"my": true, "of": true, "on": true, "or": true,
		"our": true, "such": true, "that": true, "the": true,
		"their": true, "them": true, "there": true, "to": true,
		"was": true, "we": true, "were": true, "what": true,
		"where": true, "which": true, "who": true, "will": true,
		"with": true, "would": true, "you": true, "your": true,
	}
	return minorWords[word]
}

func isPreposition(word string) bool {
	prepositions := map[string]bool{
		"after": true, "against": true, "at": true,
		"before": true, "behind": true, "below": true,
		"beside": true, "between": true, "by": true,
		"despite": true, "during": true, "for": true,
		"from": true, "in": true, "into": true,
		"near": true, "of": true, "off": true,
		"on": true, "onto": true, "out": true,
		"over": true, "since": true, "through": true,
		"to": true, "under": true, "until": true,
		"upon": true, "with": true, "within": true,
		"without": true,
	}
	return prepositions[word]
}

// find all CategoryLinks on a page
func findAllCategoryLinks(text string) []string {
	regex := regexp.MustCompile(`\[Category:([^\]|]*)\]`)
	matches := regex.FindAllStringSubmatch(text, -1) // Find all matches

	links := make([]string, 0)
	for _, match := range matches {
		links = append(links, match[1]) // Extract and append category names
	}

	return links
}

// removes the Links off the page
func removeCategoryLinks(inputString string) string {
	categoryLinkRegex := regexp.MustCompile(`\[Category:([^\]|]*)\]`)
	return categoryLinkRegex.ReplaceAllString(inputString, "")
}

// creates a table of contents List
func createHeadingList(htmlText string) string {
	regex := regexp.MustCompile(`(?s)<h([1-6]) id="([^"]+)">(.*?)</h[1-6]>`)
	headingCount := len(regex.FindAllString(htmlText, -1))
	if headingCount < 2 {
		return htmlText // No TOC if less than 2 headings
	} else {
		tocList := ""
		currentLevel := 0
		counter := 0
		levelCounters := map[int]int{}
		var level int
		htmlText = regex.ReplaceAllStringFunc(htmlText, func(match string) string {

			level, _ = strconv.Atoi(regex.ReplaceAllString(match, "$1"))
			id := strings.ToLower(regex.ReplaceAllString(match, "$2"))
			headingText := regex.ReplaceAllString(match, "$3")

			if level > currentLevel {
				counter = 1 // Reset counter for new level
				levelCounters[level] = 1
				tocList += strings.Repeat("  ", level-currentLevel-1) + "<ul>"
			} else if level < currentLevel {
				for l := level + 1; l <= currentLevel; l++ {
					tocList += strings.Repeat("  ", currentLevel-l) + "</ul>"
				}
				counter = levelCounters[level] + 1
				levelCounters[level] = counter
			} else {
				counter = levelCounters[level] + 1
				levelCounters[level] = counter
			}

			currentLevel = level
			numberPrefix := ""
			if level > 1 {
				for l := 1; l < level; l++ {
					numberPrefix += fmt.Sprintf("%d.", levelCounters[l])
				}
			}
			formattedNumber := fmt.Sprintf("%s%d", numberPrefix, counter)

			// Wrap number in a span to prevent link highlighting
			tocList += fmt.Sprintf("<li><span>%s  </span><a href=\"#%s\">%s</a></li>", formattedNumber, id, headingText)
			return fmt.Sprintf("<h%d id=\"%s\">%s</h%d>", level, id, headingText, level) // Keep original heading
		})

		if currentLevel > 0 {
			for l := level + 1; l <= currentLevel; l++ {
				tocList += strings.Repeat("  ", currentLevel-l) + "</ul>"
			}
		}

		tocList = fmt.Sprintf("<div class=\"toc\"><h6 class=\"text-center\">Contents</h6>%s</div>\n", tocList)
		return tocList + htmlText // Prepend TOC to HTML
	}
}

// adds ids to heading tags so they can be scrolled to
func addHeadingIDs(htmlText string) string {
	regex := regexp.MustCompile(`(?s)<h([1-6])>(.+?)<\/h[1-6]>`)
	counter := 0

	return regex.ReplaceAllStringFunc(htmlText, func(match string) string {
		level, headingText := parseHeadingMatch(match)
		counter++

		id := strings.ReplaceAll(strings.ToLower(headingText), " ", "-")
		return fmt.Sprintf("<h%d id=\"%s\">%s</h%d>", level, id, headingText, level)
	})
}

// finds headings on a page ready for adding ids to them
func parseHeadingMatch(match string) (int, string) {
	matches := regexp.MustCompile(`(?s)<h([1-6])>(.+?)<\/h[1-6]>`).FindStringSubmatch(match)
	level, _ := strconv.Atoi(matches[1])
	headingText := strings.TrimSpace(matches[2])
	return level, headingText
}

// used for removing the underscrores in titles for pages to make them look smart
func removeUnderscores(s string) string {
	return strings.ReplaceAll(s, "_", " ")
}

// updates headings with styling
func parseWikiText(wikiText string) string {

	regex_patterns := map[string]string{
		"h1": `(?m)^(<h1(.+?)>)(.+?)(<\/h1>)`, // Then process h1
		"h2": `(?m)^(<h2(.+?)>)(.+?)(<\/h2>)`, // Process h2 first
		"h3": `(?m)^(<h3(.+?)>)(.+?)(<\/h3>)`, // Process h2 first
		"h4": `(?m)^(<h4(.+?)>)(.+?)(<\/h4>)`, // Process h2 first

	}

	styles := map[string]string{
		"h1": `class="wikih1"`,
		"h2": `class="wikih2"`,
		"h3": `class="wikih3"`,
		"h4": `class="wikih4"`,
	}

	parsedText := wikiText

	for level, regex_pattern := range regex_patterns {
		re := regexp.MustCompile(regex_pattern)
		parsedText = re.ReplaceAllStringFunc(parsedText, func(match string) string {
			// make sure we use match and look at group 2
			// headingText := match[2 : len(match)-len(level)]
			//headingText := match[2 : len(match)-len(level)]
			matches := re.FindStringSubmatch(match)
			//fmt.Println(matches[2])
			////fmt.Println("Match found:", matches[1])
			return fmt.Sprintf("<%s %s %s>%s</%s>", level, matches[2], styles[level], matches[3], level)
		})
	}
	//fmt.Println("Match found:", parsedText)
	return parsedText
}
func convertLinksToAnchors(text string) string {
	baseURL := "http://localhost:8080/title/"
	linkRegex := regexp.MustCompile(`\[(.*?)\]`)
	return linkRegex.ReplaceAllStringFunc(text, func(match string) string {
		linkText := match[1 : len(match)-1] // Extract link text without brackets
		return `<a href="` + baseURL + linkText + `">` + linkText + `</a>`
	})
}
