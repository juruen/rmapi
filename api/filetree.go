package api

import (
	"errors"
	"strings"

	"github.com/juruen/rmapi/util"
)

type FileTreeCtx struct {
	root          *Node
	idToNode      map[string]*Node
	pendingParent map[string]map[string]struct{}
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

func (ctx *FileTreeCtx) NodeToPath(node *Node) (string, error) {
	path, err := dfs(ctx.Root(), node, make([]string, 0))

	if err != nil {
		return "", err
	}

	if len(path) == 1 {
		return "/", nil
	}

	return path[1:len(path)], nil
}

func dfs(node, target *Node, pathResult []string) (string, error) {
	pathResult = append(pathResult, node.Name())

	if target == node {
		return strings.Join(pathResult, "/"), nil
	}

	newPathResult := make([]string, len(pathResult))
	copy(newPathResult, pathResult)

	for _, c := range node.Children {
		if n, err := dfs(c, target, newPathResult); err == nil {
			return n, nil
		}
	}

	return "", errors.New("node not found")
}
