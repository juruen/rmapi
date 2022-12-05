package model

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type Node struct {
	Document *Document
	Children map[string]*Node
	Parent   *Node
}

func (node *Node) Nodes() []*Node {
	result := make([]*Node, 0)
	for _, n := range node.Children {
		result = append(result, n)
	}
	return result

}

func CreateNode(document Document) Node {
	return Node{&document, make(map[string]*Node, 0), nil}
}

func (node *Node) Name() string {
	return node.Document.VissibleName
}

func (node *Node) Id() string {
	return node.Document.ID
}
func (node *Node) Version() int {
	return node.Document.Version
}

func (node *Node) IsRoot() bool {
	return node.Id() == ""
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

func (node *Node) LastModified() (time.Time, error) {
	return time.Parse(time.RFC3339Nano, node.Document.ModifiedClient)
}

func (node *Node) FindByName(name string) (*Node, error) {
	for _, n := range node.Children {
		if n.Name() == name {
			return n, nil
		}
	}
	return nil, fmt.Errorf("entry '%s' doesnt exist", name)
}

func (node *Node) FindByPattern(pattern string) ([]*Node, error) {
	result := make([]*Node, 0)
	if pattern == "" {
		return nil, errors.New("empty pattern")
	}

	lowerCasePattern := strings.ToLower(pattern)

	for _, n := range node.Children {
		matched, err := filepath.Match(lowerCasePattern, strings.ToLower(n.Name()))
		if err != nil {
			return nil, err
		}
		if matched {
			result = append(result, n)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no matches for '%s'", pattern)
	}
	return result, nil
}
