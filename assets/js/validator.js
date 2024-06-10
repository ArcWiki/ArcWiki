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
//bbq
(function () {
  'use strict'

  // Fetch the title input element
  const titleInput = document.getElementById("title");
 
  // Add input event listener for real-time validation
  if (titleInput) {
    titleInput.addEventListener("input", () => {
      const titleValue = titleInput.value;
      //const lettersOnlyPattern = /^[a-zA-Z]+$/;
      const expandedPattern = /^[a-zA-Z0-9_.,!?'" ]+$/;

      if (expandedPattern.test(titleValue)) {
        titleInput.classList.remove("is-invalid");
        titleInput.classList.add("is-valid");
      } else {
        titleInput.classList.remove("is-valid");
        titleInput.classList.add("is-invalid");
      }
    });
}


  // Code for preventing form submission with invalid fields (already provided)
  // ...
})();