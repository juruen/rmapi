package filetree

import (
	"errors"

	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/util"
)

type FileTreeCtx struct {
	root          *model.Node
	idToNode      map[string]*model.Node
	pendingParent map[string]map[string]struct{}
}

type FileTreeVistor struct {
	Visit func(node *model.Node, path []string) bool
}

func CreateFileTreeCtx() FileTreeCtx {
	root := model.CreateNode(model.Document{
		ID:           "",
		Type:         "CollectionType",
		VissibleName: "/",
	})

	return FileTreeCtx{
		&root,
		make(map[string]*model.Node),
		make(map[string]map[string]struct{}),
	}
}

func (ctx *FileTreeCtx) Root() *model.Node {
	return ctx.root
}

func (ctx *FileTreeCtx) NodeById(id string) *model.Node {
	if len(id) == 0 {
		return ctx.Root()
	}

	if n, ok := ctx.idToNode[id]; ok {
		return n
	} else {
		return nil
	}
}

func (ctx *FileTreeCtx) AddDocument(document model.Document) {
	node := model.CreateNode(document)
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

func (ctx *FileTreeCtx) DeleteNode(node *model.Node) {
	if node.IsRoot() {
		return
	}

	delete(node.Parent.Children, node.Id())
}

func (ctx *FileTreeCtx) MoveNode(src, dst *model.Node) {
	if src.IsRoot() {
		return
	}

	src.Document.VissibleName = dst.Document.VissibleName
	src.Document.Version = dst.Document.Version
	src.Document.ModifiedClient = dst.Document.ModifiedClient

	if src.Parent != dst.Parent {
		delete(src.Parent.Children, src.Id())
		src.Parent = dst.Parent
		dst.Parent.Children[src.Id()] = src
	}
}

func (ctx *FileTreeCtx) NodeByPath(path string, current *model.Node) (*model.Node, error) {
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

func (ctx *FileTreeCtx) NodeToPath(targetNode *model.Node) (string, error) {
	resultPath := ""
	found := false

	visitor := FileTreeVistor{
		func(currentNode *model.Node, path []string) bool {
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
