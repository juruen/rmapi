package api

import "strings"

const (
	StopVisiting     = true
	ContinueVisiting = false
)

func WalkTree(node *Node, visitor FileTreeVistor) {
	doWalkTree(node, make([]string, 0), visitor)
}

func doWalkTree(node *Node, path []string, visitor FileTreeVistor) bool {
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

func BuildPath(path []string, entry string) string {
	if len(path) == 0 {
		return entry
	}

	path = append(path, entry)
	resultPath := strings.Join(path, "/")

	if len(path) > 1 && path[0] == "/" {
		return resultPath[1:len(resultPath)]
	}
	return resultPath
}
