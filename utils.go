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

func stringSliceToMap(s []string) map[string]bool {
	m := make(map[string]bool)
	for _, k := range s {
		m[k] = true
	}
	return m
}

func complement(a []string, b map[string]bool) []string {
	c := make([]string, 0, 5)
	for _, m := range a {
		if val, _ := b[m]; !val {
			c = append(c, m)
		}
	}
	return c
}
