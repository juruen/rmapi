package filetree

import (
	"path"
	"strings"

	"github.com/juruen/rmapi/model"
)

const (
	StopVisiting     = true
	ContinueVisiting = false
)

func WalkTree(node *model.Node, visitor FileTreeVistor) {
	doWalkTree(node, make([]string, 0), visitor)
}

func doWalkTree(node *model.Node, path []string, visitor FileTreeVistor) bool {
	if visitor.Visit(node, path) {
		return StopVisiting
	}

	newPath := appendEntryPath(path, node.Name())

	for _, c := range node.Children {
		if doWalkTree(c, newPath, visitor) {
			return StopVisiting
		}
	}

	return ContinueVisiting
}

func appendEntryPath(currentPath []string, entry string) []string {
	newPath := make([]string, len(currentPath))
	copy(newPath, currentPath)
	newPath = append(newPath, entry)

	return newPath
}

func BuildPath(_path []string, entry string) string {
	if len(_path) == 0 {
		return entry
	}

	resultPath := path.Join(path.Join(_path...), entry)

	return strings.TrimPrefix(resultPath, "/")
}
