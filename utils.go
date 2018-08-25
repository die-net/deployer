package main

import (
	"strings"
)

// Simplified strings.SplitN() that always returns two strings.
func splitTwo(s, sep string) (one, two string) {
	if part := strings.SplitN(s, sep, 2); len(part) == 2 {
		return part[0], part[1]
	}

	return s, ""
}
