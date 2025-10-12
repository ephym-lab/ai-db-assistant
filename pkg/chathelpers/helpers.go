package chathelpers

import "strings"

// ToLower converts a string to lowercase
func ToLower(s string) string {
	return strings.ToLower(s)
}

// Contains checks if a string contains any of the given keywords
func Contains(s string, keywords ...string) bool {
	lowerS := strings.ToLower(s)
	for _, keyword := range keywords {
		if strings.Contains(lowerS, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// ContainsAll checks if a string contains all of the given keywords
func ContainsAll(s string, keywords ...string) bool {
	lowerS := strings.ToLower(s)
	for _, keyword := range keywords {
		if !strings.Contains(lowerS, strings.ToLower(keyword)) {
			return false
		}
	}
	return true
}

// HasAny checks if any of the keywords exist in the string
func HasAny(s string, keywords []string) bool {
	lowerS := strings.ToLower(s)
	for _, keyword := range keywords {
		if strings.Contains(lowerS, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}