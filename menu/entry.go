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
package menu

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strings"
)

func Load() (template.HTML, error) {
	file, err := os.Open("menu/menu.json")
	if err != nil {
		fmt.Println("Error Loading menu: " + err.Error())

	}

	defer file.Close()

	// Decode the JSON data
	var data map[string]interface{}
	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		return "", err // Return the error
	}

	// Extract the menu items (now an ordered list)
	menuItems := data["menu"].([]interface{})

	// Create the list of links
	links := strings.Builder{}
	for _, item := range menuItems {
		linkData := item.(map[string]interface{})
		name := linkData["name"].(string)
		link := linkData["link"].(string)
		links.WriteString(fmt.Sprintf("<li><a href=\"%s\">%s</a></li>\n", link, name))
	}

	return template.HTML(links.String()), nil
}
