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


const easyMDE = new EasyMDE(
    {toolbar: ["bold", "italic", "heading", "|", "quote", "unordered-list", "ordered-list", "|", "link", "image", "code", "preview", {
      name: "checkboxfill",
        action: (editor) => {
  
          const cm = editor.codemirror;
  
          const output = " <i class=\"ri-checkbox-fill\"></i>";
          const selectedText = cm.getSelection();
          const text = selectedText || ''; // Use empty string if no selection
          const cursor = cm.getCursor();
          if (cursor) {
            // Insert at cursor
            cm.replaceSelection(output + text);
  
            // Move cursor inside the square brackets
            cm.setCursor({ line: cursor.line, ch: cursor.ch + 1 });
          }
          
        },
        className: "ri-checkbox-fill",
        
        title: "Checkbox",
      },
      
      
      {
      name: "internal",
      action: (editor) => {
  
          const cm = editor.codemirror;
  
          const output = "[]";
          const selectedText = cm.getSelection();
          const text = selectedText || ''; // Use empty string if no selection
          const cursor = cm.getCursor();
          if (cursor) {
            // Insert at cursor
            cm.replaceSelection(output + text);
            cm.focus();
  
            // Move cursor inside the square brackets
            cm.setCursor({ line: cursor.line, ch: cursor.ch + 4 });
          }
  
          },
        className: "ri-link",
        
        title: "Internal Link",
      },
      {
      name: "category",
        action: (editor) => {
          const cm = editor.codemirror;
  
          const output = "[Category:]";
          const selectedText = cm.getSelection();
          const text = selectedText || ''; // Use empty string if no selection
          const cursor = cm.getCursor();
          if (cursor) {
            // Insert at cursor
            cm.replaceSelection(output + text);
  
            // Move cursor inside the square brackets
            cm.setCursor({ line: cursor.line, ch: cursor.ch + 2 });
          }
        },
        
        className: "ri-file-list-line",
        
        title: "Category",
      },
    ]}
     );