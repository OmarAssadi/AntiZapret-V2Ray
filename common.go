package main

import "strings"

func isBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

func isEmpty(s string) bool {
	return s == ""
}

func trimStrings(arr []string) {
	for i := range arr {
		arr[i] = strings.TrimSpace(arr[i])
	}
}

func splitAndTrim(s string, d string) (int, []string) {
	parts := strings.Split(s, d)
	trimStrings(parts)
	return len(parts), parts
}

func splitAfterAndCount(str string, delimiter string) (int, []string) {
	parts := strings.SplitAfter(str, delimiter)
	return len(parts), parts
}
