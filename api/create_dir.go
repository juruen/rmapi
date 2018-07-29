package api

import (
	"errors"
	"os"

	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/transport"
	"github.com/juruen/rmapi/util"
)

func (ctx *ApiCtx) CreateDir(parentId, name string) (model.Document, error) {
	uploadRsp, err := ctx.uploadRequest(model.DirectoryType)

	if err != nil {
		return model.Document{}, err
	}

	if !uploadRsp.Success {
		return model.Document{}, errors.New("upload request returned success := false")
	}

	zippath, err := util.CreateZipDirectory(uploadRsp.ID)

	if err != nil {
		log.Error.Println("failed to create zip directory", err)
		return model.Document{}, err
	}

	f, err := os.Open(zippath)
	defer f.Close()

	if err != nil {
		log.Error.Println("failed to read zip file to upload", zippath, err)
		return model.Document{}, err
	}

	err = ctx.Http.PutStream(transport.UserBearer, uploadRsp.BlobURLPut, f)

	if err != nil {
		log.Error.Println("failed to upload directory", err)
		return model.Document{}, err
	}

	metaDoc := model.CreateUploadDocumentMeta(uploadRsp.ID, model.DirectoryType, parentId, name)

	err = ctx.Http.Put(transport.UserBearer, updateStatus, metaDoc, nil)

	if err != nil {
		log.Error.Println("failed to move entry", err)
		return model.Document{}, err
	}

	doc := metaDoc.ToDocument()

	return doc, err

}
