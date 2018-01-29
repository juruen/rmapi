package util

import "strings"

func SplitPath(path string) []string {
	return strings.Split(path, "/")
}
