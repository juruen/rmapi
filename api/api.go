package api

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/juruen/rmapi/filetree"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/transport"
	"github.com/juruen/rmapi/util"
)

type ApiCtx struct {
	Http     *transport.HttpClientCtx
	Filetree *filetree.FileTreeCtx
}

func CreateApiCtx(http *transport.HttpClientCtx) *ApiCtx {
	ctx := ApiCtx{http, DocumentsFileTree(http)}
	return &ctx
}

func DocumentsFileTree(http *transport.HttpClientCtx) *filetree.FileTreeCtx {
	documents := make([]model.Document, 0)

	if err := http.Get(transport.UserBearer, listDocs, nil, &documents); err != nil {
		log.Error.Println("failed to fetch documents %s", err.Error())
		return nil
	}

	fileTree := filetree.CreateFileTreeCtx()

	for _, d := range documents {
		fileTree.AddDocument(d)
	}

	for _, d := range fileTree.Root().Children {
		log.Trace.Println(d.Name(), d.IsFile())
	}

	return &fileTree
}

func (ctx *ApiCtx) FetchDocument(docId, dstPath string) error {
	documents := make([]model.Document, 0)

	url := fmt.Sprintf("%s?withBlob=true&doc=%s", listDocs, docId)

	if err := ctx.Http.Get(transport.UserBearer, url, nil, &documents); err != nil {
		log.Error.Println("failed to fetch document BlobURLGet %s", err)
		return err
	}

	if len(documents) == 0 || documents[0].BlobURLGet == "" {
		log.Error.Println("BlobURLGet for document is empty")
		return errors.New("no BlobURLGet")
	}

	blobUrl := documents[0].BlobURLGet

	src, err := ctx.Http.GetStream(transport.UserBearer, blobUrl)

	if src != nil {
		defer src.Close()
	}

	if err != nil {
		log.Error.Println("Error fetching blob")
		return err
	}

	dst, err := ioutil.TempFile("", "rmapifile")

	if err != nil {
		log.Error.Println("failed to create temp fail to download blob")
		return err
	}

	tmpPath := dst.Name()
	defer dst.Close()
	defer os.Remove(tmpPath)

	_, err = io.Copy(dst, src)

	if err != nil {
		log.Error.Println("failed to download blob")
		return err
	}

	_, err = util.CopyFile(tmpPath, dstPath)

	if err != nil {
		log.Error.Printf("failgied to copy %s to %s, er: %s\n", tmpPath, dstPath, err.Error())
		return err
	}

	return nil
}

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
