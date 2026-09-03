package dsasnippets

import (
	"fmt"
	"unicode"
)

// isAlphanumericUnicode checks if a character is a letter or digit in any language.
func isAlphanumericUnicode(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch)
}

func charAt() {
	str := "Go世界" // "世界" uses 3 bytes per character
	// Convert to a slice of runes
	runes := []rune(str)
	// Safely get the character at index 2
	r := runes[2]
	fmt.Println(string(r)) // Output: 世
}
