package api

type Node struct {
	Document *Document
	Children map[string]*Node
	Parent   *Node
}

func CreateNode(document Document) Node {
	return Node{&document, make(map[string]*Node, 0), nil}
}

func (node *Node) Name() string {
	return node.Document.VissibleName
}

func (node *Node) IsDirectory() bool {
	return node.Document.Type == "CollectionType"
}

func (node *Node) IsFile() bool {
	return !node.IsDirectory()
}

func (node *Node) EntyExists(id string) bool {
	_, ok := node.Children[id]
	return ok
}
