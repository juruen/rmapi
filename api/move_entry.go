package api

import (
	"errors"

	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/transport"
)

func (ctx *ApiCtx) MoveEntry(src, dstDir *model.Node, name string) (*model.Node, error) {
	if dstDir.IsFile() {
		return nil, errors.New("destination directory is a file")
	}

	metaDoc := src.Document.ToMetaDocument()
	metaDoc.Version = metaDoc.Version + 1
	metaDoc.VissibleName = name
	metaDoc.Parent = dstDir.Id()

	err := ctx.Http.Put(transport.UserBearer, updateStatus, metaDoc, nil)

	if err != nil {
		log.Error.Println("failed to move entry", err)
		return nil, err
	}

	doc := metaDoc.ToDocument()

	return &model.Node{&doc, src.Children, dstDir}, nil
}
