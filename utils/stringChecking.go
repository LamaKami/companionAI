package utils

import "regexp"

func CheckStringAlphabet(name string) bool {
	var isStringAlphabetic = regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString
	return isStringAlphabetic(name)
}
