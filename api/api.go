package api

import (
	"github.com/juruen/rmapi/api/sync10"
	"github.com/juruen/rmapi/filetree"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/transport"
)

type ApiCtx interface {
	Filetree() *filetree.FileTreeCtx
	FetchDocument(docId, dstPath string) error
	CreateDir(parentId, name string) (model.Document, error)
	UploadDocument(parentId string, sourceDocPath string) (*model.Document, error)
	MoveEntry(src, dstDir *model.Node, name string) (*model.Node, error)
	DeleteEntry(node *model.Node) error
	Nuke() error
}

// CreateApiCtx initializes an instance of ApiCtx
func CreateApiCtx(http *transport.HttpClientCtx) (ApiCtx, error) {

	return sync10.CreateCtx(http)

}
