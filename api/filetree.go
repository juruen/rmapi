package api

import (
	"errors"
	"strings"

	"github.com/juruen/rmapi/util"
)

const (
	stopVisiting     = true
	continueVisiting = false
)

type FileTreeCtx struct {
	root          *Node
	idToNode      map[string]*Node
	pendingParent map[string]map[string]struct{}
}

type FileTreeVistor struct {
	visit func(node *Node, path []string) bool
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

func (ctx *FileTreeCtx) NodeToPath(targetNode *Node) (string, error) {
	resultPath := ""
	found := false

	visitor := FileTreeVistor{
		func(currentNode *Node, path []string) bool {
			if targetNode != currentNode {
				return continueVisiting
			}

			found = true

			if len(path) == 0 {
				resultPath = currentNode.Name()
				return stopVisiting
			}

			path = append(path, currentNode.Name())
			resultPath = strings.Join(path, "/")
			resultPath = resultPath[1:len(resultPath)]

			return stopVisiting
		},
	}

	WalkTree(ctx.root, visitor)

	if found {
		return resultPath, nil
	} else {
		return "", errors.New("entry not found")
	}
}

func WalkTree(node *Node, visitor FileTreeVistor) {
	doWalkTree(node, make([]string, 0), visitor)
}

func appendEntryPath(currentPath []string, entry string) []string {
	newPath := make([]string, len(currentPath))
	copy(newPath, currentPath)
	newPath = append(newPath, entry)

	return newPath
}

func doWalkTree(node *Node, path []string, visitor FileTreeVistor) bool {
	if visitor.visit(node, path) {
		return stopVisiting
	}

	newPath := appendEntryPath(path, node.Name())

	for _, c := range node.Children {
		if doWalkTree(c, newPath, visitor) {
			return stopVisiting
		}
	}

	return continueVisiting
}
