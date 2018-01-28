package main

type Node struct {
	Document *Document
	Children map[string]*Node
	Parent   *Node
}

func CreateNode(document Document) Node {
	return Node{&document, make(map[string]*Node, 0), nil}
}

func (node *Node) name() string {
	return node.Document.VissibleName
}

func (node *Node) isDirectory() bool {
	return node.Document.Type == "CollectionType"
}

func (node *Node) isFile() bool {
	return !node.isDirectory()
}

func (node *Node) entyExists(id string) bool {
	_, ok := node.Children[id]
	return ok
}
