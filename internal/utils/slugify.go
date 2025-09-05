package utils

import (
	"regexp"
	"strings"
)

func Slugify(input string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		panic(err)
	}
	processedString := reg.ReplaceAllString(input, " ")
	processedString = strings.TrimSpace(processedString)
	slug := strings.ReplaceAll(processedString, " ", "-")
	slug = strings.ToLower(slug)

	return slug
}

// UniqueWords extracts unique words from a string, converts them to lowercase.
func UniqueWords(input string) []string {
	words := strings.Fields(strings.ToLower(input))
	uniqueMap := make(map[string]bool)
	for _, word := range words {
		uniqueMap[word] = true
	}
	uniqueSlice := make([]string, 0, len(uniqueMap))
	for word := range uniqueMap {
		uniqueSlice = append(uniqueSlice, word)
	}
	return uniqueSlice
}