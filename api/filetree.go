package api

import (
	"errors"
	"strings"

	"github.com/juruen/rmapi/util"
)

const (
	StopVisiting     = true
	ContinueVisiting = false
)

type FileTreeCtx struct {
	root          *Node
	idToNode      map[string]*Node
	pendingParent map[string]map[string]struct{}
}

type FileTreeVistor struct {
	Visit func(node *Node, path []string) bool
}

func CreateFileTreeCtx() FileTreeCtx {
	root := CreateNode(Document{
		ID:           "1",
		Type:         "CollectionType",
		VissibleName: "/",
	})

	return FileTreeCtx{
		&root,
		make(map[string]*Node),
		make(map[string]map[string]struct{}),
	}
}

func (ctx *FileTreeCtx) Root() *Node {
	return ctx.root
}

func (ctx *FileTreeCtx) addDocument(document Document) {
	node := CreateNode(document)
	nodeId := document.ID
	parentId := document.Parent

	ctx.idToNode[nodeId] = &node

	if parentId == "" {
		// This is a node whose parent is root
		node.Parent = ctx.root
		ctx.root.Children[nodeId] = &node
	} else if parentNode, ok := ctx.idToNode[parentId]; ok {
		// Parent node already processed
		node.Parent = parentNode
		parentNode.Children[nodeId] = &node
	} else {
		// Parent node hasn't been processed yet
		if _, ok := ctx.pendingParent[parentId]; !ok {
			ctx.pendingParent[parentId] = make(map[string]struct{})
		}
		ctx.pendingParent[parentId][nodeId] = struct{}{}
	}

	// Resolve pendingChildren
	if pendingChildren, ok := ctx.pendingParent[nodeId]; ok {
		for id, _ := range pendingChildren {
			ctx.idToNode[id].Parent = &node
			node.Children[id] = ctx.idToNode[id]
		}
		delete(ctx.pendingParent, nodeId)
	}
}

func (ctx *FileTreeCtx) NodeByPath(path string, current *Node) (*Node, error) {
	if current == nil {
		current = ctx.Root()
	}

	entries := util.SplitPath(path)

	if len(entries) == 0 {
		return current, nil
	}

	i := 0
	if entries[i] == "" {
		current = ctx.Root()
		i++
	}

	for i < len(entries) {
		if entries[i] == "" || entries[i] == "." {
			i++
			continue
		}

		if entries[i] == ".." {
			if current.Parent == nil {
				current = ctx.Root()
			} else {
				current = current.Parent
			}

			i++
			continue
		}

		var err error
		current, err = current.FindByName(entries[i])

		if err != nil {
			return nil, err
		}

		i++
	}

	return current, nil
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

func (ctx *FileTreeCtx) NodeToPath(targetNode *Node) (string, error) {
	resultPath := ""
	found := false

	visitor := FileTreeVistor{
		func(currentNode *Node, path []string) bool {
			if targetNode != currentNode {
				return ContinueVisiting
			}

			found = true
			resultPath = BuildPath(path, currentNode.Name())
			return StopVisiting
		},
	}

	ctx.WalkTree(ctx.root, visitor)

	if found {
		return resultPath, nil
	} else {
		return "", errors.New("entry not found")
	}
}

func (_ *FileTreeCtx) WalkTree(node *Node, visitor FileTreeVistor) {
	doWalkTree(node, make([]string, 0), visitor)
}

func appendEntryPath(currentPath []string, entry string) []string {
	newPath := make([]string, len(currentPath))
	copy(newPath, currentPath)
	newPath = append(newPath, entry)

	return newPath
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
