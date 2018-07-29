package api

import (
	"github.com/juruen/rmapi/filetree"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/transport"
)

type RmApi interface {
	FetchDocument(docId, dstPath string) error
	FetchHttpMetaDocument(docId string) (model.HttpDocumentMeta, error)
	CreateDir(parentId, name string) (model.Document, error)
	DeleteEntry(node *model.Node) error
	MoveEntry(src, dstDir *model.Node, name string) (*model.Node, error)
	UploadDocument(parent string, pdfPath string) (*model.Document, error)
	DocumentsFileTree() *filetree.FileTreeCtx
}

type ApiCtx struct {
	Http     *transport.HttpClientCtx
	Filetree *filetree.FileTreeCtx
}

func CreateApiCtx(http *transport.HttpClientCtx) *ApiCtx {
	ctx := ApiCtx{http, DocumentsFileTree(http)}
	return &ctx
}
