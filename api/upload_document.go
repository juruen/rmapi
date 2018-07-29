package api

import (
	"errors"
	"os"

	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/transport"
	"github.com/juruen/rmapi/util"
)

func (ctx *ApiCtx) UploadDocument(parent string, pdfpath string) (*model.Document, error) {
	name := util.DocPathToName(pdfpath)

	if name == "" {
		return nil, errors.New("file name is invalid")
	}

	uploadRsp, err := ctx.uploadRequest(model.DocumentType)

	if err != nil {
		return nil, err
	}

	if !uploadRsp.Success {
		return nil, errors.New("upload request returned success := false")
	}

	zippath, err := util.CreateZipDocument(uploadRsp.ID, pdfpath)

	if err != nil {
		log.Error.Println("failed to create zip doc", err)
		return nil, err
	}

	f, err := os.Open(zippath)
	defer f.Close()

	if err != nil {
		log.Error.Println("failed to read zip file to upload", zippath, err)
		return nil, err
	}

	err = ctx.Http.PutStream(transport.UserBearer, uploadRsp.BlobURLPut, f)

	if err != nil {
		log.Error.Println("failed to upload zip document", err)
		return nil, err
	}

	metaDoc := model.CreateUploadDocumentMeta(uploadRsp.ID, model.DocumentType, parent, name)

	err = ctx.Http.Put(transport.UserBearer, updateStatus, metaDoc, nil)

	if err != nil {
		log.Error.Println("failed to move entry", err)
		return nil, err
	}

	doc := metaDoc.ToDocument()

	return &doc, err
}

func (ctx *ApiCtx) uploadRequest(entryType string) (model.UploadDocumentResponse, error) {
	uploadReq := model.CreateUploadDocumentRequest(entryType)
	uploadRsp := make([]model.UploadDocumentResponse, 0)

	err := ctx.Http.Put(transport.UserBearer, uploadRequest, uploadReq, &uploadRsp)

	if err != nil {
		log.Error.Println("failed to to send upload request", err)
		return model.UploadDocumentResponse{}, err
	}

	return uploadRsp[0], nil
}
