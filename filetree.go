package main

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
