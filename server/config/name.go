package config

import (
	"regexp"
)

var (
	namePattern = regexp.MustCompile(`[a-z][a-zA-Z0-9]*`)
)

//ValidName ensures the name matches the name regex
func ValidName(name string) bool {
	return namePattern.MatchString(name)
}
