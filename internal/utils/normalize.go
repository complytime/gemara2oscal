package utils

import (
	"regexp"
	"strings"
)

// Assisted by: Gemini 2.5 Flash

func NormalizeControl(input string) string {
	// Compile the regular expression to find patterns like (number).
	// \( and \) are used to match literal parentheses.
	// (\d+) captures one or more digits inside the parentheses.
	re := regexp.MustCompile(`\((\d+)\)`)

	// Replace all occurrences of the pattern.
	// ".$1" means replace with a dot followed by the content of the first captured group (the digits).
	replacedString := re.ReplaceAllString(input, ".$1")

	// Convert the entire resulting string to lowercase.
	finalString := strings.ToLower(replacedString)

	return finalString
}
