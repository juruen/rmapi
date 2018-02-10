package api

import (
	"errors"

	"github.com/juruen/rmapi/util"
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

func (ctx *FileTreeCtx) AddDocument(document Document) {
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

func (ctx *FileTreeCtx) DeleteNode(node *Node) {
	if node.IsRoot() {
		return
	}

	delete(node.Parent.Children, node.Id())
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
				return ContinueVisiting
			}

			found = true
			resultPath = BuildPath(path, currentNode.Name())
			return StopVisiting
		},
	}

	WalkTree(ctx.root, visitor)

	if found {
		return resultPath, nil
	} else {
		return "", errors.New("entry not found")
	}
}
