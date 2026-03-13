package validator

import (
	"regexp"
	"strings"
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

func ValidateUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

func ValidatePassword(password string) bool {
	return len(password) >= 6 && len(password) <= 128
}

func SanitizeString(s string) string {
	return strings.TrimSpace(s)
}
