package api

import (
	"errors"

	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/transport"
)

func (ctx *ApiCtx) DeleteEntry(node *model.Node) error {
	if node.IsDirectory() && len(node.Children) > 0 {
		return errors.New("directory is not empty")
	}

	deleteDoc := node.Document.ToDeleteDocument()

	err := ctx.Http.Put(transport.UserBearer, deleteEntry, deleteDoc, nil)

	if err != nil {
		log.Error.Println("failed to remove entry", err)
		return err
	}

	return nil
}
