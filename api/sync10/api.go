package sync10

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/juruen/rmapi/archive"
	"github.com/juruen/rmapi/config"
	"github.com/juruen/rmapi/filetree"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/transport"
	"github.com/juruen/rmapi/util"
)

// An ApiCtx allows you interact with the remote reMarkable API
type ApiCtx struct {
	Http *transport.HttpClientCtx
	ft   *filetree.FileTreeCtx
}

func (ctx *ApiCtx) Filetree() *filetree.FileTreeCtx {
	return ctx.ft
}

func (ctx *ApiCtx) Refresh() (err error) {
	return errors.New("not implemented")
}

// Nuke removes all documents from the account
func (ctx *ApiCtx) Nuke() error {
	documents := make([]model.Document, 0)

	if err := ctx.Http.Get(transport.UserBearer, config.ListDocs, nil, &documents); err != nil {
		return err
	}

	for _, d := range documents {
		log.Info.Println("Deleting: ", d.VissibleName)

		err := ctx.Http.Put(transport.UserBearer, config.DeleteEntry, util.InSlice(d), nil)
		if err != nil {
			log.Error.Println("failed to remove entry", err)
			return err
		}
	}

	return nil
}

// FetchDocument downloads a document given its ID and saves it locally into dstPath
func (ctx *ApiCtx) FetchDocument(docId, dstPath string) error {
	documents := make([]model.Document, 0)

	url := fmt.Sprintf("%s?withBlob=true&doc=%s", config.ListDocs, docId)

	if err := ctx.Http.Get(transport.UserBearer, url, nil, &documents); err != nil {
		log.Error.Println("failed to fetch document BlobURLGet", err)
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
		log.Error.Printf("failed to copy %s to %s, er: %s\n", tmpPath, dstPath, err.Error())
		return err
	}

	return nil
}

// CreateDir creates a remote directory with a given name under the parentId directory
func (ctx *ApiCtx) CreateDir(parentId, name string, notify bool) (*model.Document, error) {
	uploadRsp, err := ctx.uploadRequest("", model.DirectoryType)

	if err != nil {
		return nil, err
	}

	if !uploadRsp.Success {
		return nil, errors.New("upload request returned success := false")
	}

	zippath, err := archive.CreateZipDirectory(uploadRsp.ID)

	if err != nil {
		log.Error.Println("failed to create zip directory", err)
		return nil, err
	}

	f, err := os.Open(zippath)

	if err != nil {
		log.Error.Println("failed to read zip file to upload", zippath, err)
		return nil, err
	}

	defer f.Close()

	err = ctx.Http.PutStream(transport.UserBearer, uploadRsp.BlobURLPut, f)

	if err != nil {
		log.Error.Println("failed to upload directory", err)
		return nil, err
	}

	metaDoc := model.CreateUploadDocumentMeta(uploadRsp.ID, model.DirectoryType, parentId, name)

	err = ctx.Http.Put(transport.UserBearer, config.UpdateStatus, util.InSlice(metaDoc), nil)

	if err != nil {
		log.Error.Println("failed to move entry", err)
		return nil, err
	}

	doc := metaDoc.ToDocument()

	return &doc, err

}

// DeleteEntry removes an entry: either an empty directory or a file
func (ctx *ApiCtx) DeleteEntry(node *model.Node) error {
	if node.IsDirectory() && len(node.Children) > 0 {
		return errors.New("directory is not empty")
	}

	deleteDoc := node.Document.ToDeleteDocument()

	err := ctx.Http.Put(transport.UserBearer, config.DeleteEntry, util.InSlice(deleteDoc), nil)

	if err != nil {
		log.Error.Println("failed to remove entry", err)
		return err
	}

	return nil
}

// MoveEntry moves an entry (either a directory or a file)
// - src is the source node to be moved
// - dstDir is an existing destination directory
// - name is the new name of the moved entry in the destination directory
func (ctx *ApiCtx) MoveEntry(src, dstDir *model.Node, name string) (*model.Node, error) {
	if dstDir.IsFile() {
		return nil, errors.New("destination directory is a file")
	}

	metaDoc := src.Document.ToMetaDocument()
	metaDoc.Version = metaDoc.Version + 1
	metaDoc.VissibleName = name
	metaDoc.Parent = dstDir.Id()

	err := ctx.Http.Put(transport.UserBearer, config.UpdateStatus, util.InSlice(metaDoc), nil)

	if err != nil {
		log.Error.Println("failed to move entry", err)
		return nil, err
	}

	doc := metaDoc.ToDocument()

	return &model.Node{&doc, src.Children, dstDir}, nil
}

// UploadDocument uploads a local document given by sourceDocPath under the parentId directory
func (ctx *ApiCtx) UploadDocument(parentId string, sourceDocPath string, notify bool) (*model.Document, error) {
	name, ext := util.DocPathToName(sourceDocPath)

	if name == "" {
		return nil, errors.New("file name is invalid")
	}

	if !util.IsFileTypeSupported(ext) {
		return nil, errors.New("unsupported file extension: " + ext)
	}

	id := ""
	var err error

	//restore document
	if ext == "zip" {
		id, err = archive.GetIdFromZip(sourceDocPath)
		if err != nil {
			return nil, err
		}
		if id == "" {
			return nil, errors.New("could not determine the Document UUID")
		}
	}

	uploadRsp, err := ctx.uploadRequest(id, model.DocumentType)

	if err != nil {
		return nil, err
	}

	if !uploadRsp.Success {
		return nil, errors.New("upload request returned success := false")
	}

	zipPath, err := archive.CreateZipDocument(uploadRsp.ID, sourceDocPath)

	if err != nil {
		log.Error.Println("failed to create zip doc", err)
		return nil, err
	}

	f, err := os.Open(zipPath)
	defer f.Close()

	if err != nil {
		log.Error.Println("failed to read zip file to upload", zipPath, err)
		return nil, err
	}

	err = ctx.Http.PutStream(transport.UserBearer, uploadRsp.BlobURLPut, f)

	if err != nil {
		log.Error.Println("failed to upload zip document", err)
		return nil, err
	}

	metaDoc := model.CreateUploadDocumentMeta(uploadRsp.ID, model.DocumentType, parentId, name)

	err = ctx.Http.Put(transport.UserBearer, config.UpdateStatus, util.InSlice(metaDoc), nil)

	if err != nil {
		log.Error.Println("failed to move entry", err)
		return nil, err
	}

	doc := metaDoc.ToDocument()

	return &doc, err
}

func (ctx *ApiCtx) uploadRequest(id string, entryType string) (model.UploadDocumentResponse, error) {
	uploadReq := model.CreateUploadDocumentRequest(id, entryType)
	uploadRsp := make([]model.UploadDocumentResponse, 0)

	err := ctx.Http.Put(transport.UserBearer, config.UploadRequest, util.InSlice(uploadReq), &uploadRsp)

	if err != nil {
		log.Error.Println("failed to to send upload request", err)
		return model.UploadDocumentResponse{}, err
	}

	return uploadRsp[0], nil
}

func CreateCtx(http *transport.HttpClientCtx) (*ApiCtx, error) {

	tree, err := DocumentsFileTree(http)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch document tree %v", err)
	}
	return &ApiCtx{http, tree}, nil
}

// DocumentsFileTree reads your remote documents and builds a file tree
// structure to represent them
func DocumentsFileTree(http *transport.HttpClientCtx) (*filetree.FileTreeCtx, error) {
	documents := make([]*model.Document, 0)

	if err := http.Get(transport.UserBearer, config.ListDocs, nil, &documents); err != nil {
		return nil, err
	}

	fileTree := filetree.CreateFileTreeCtx()

	for _, d := range documents {
		fileTree.AddDocument(d)
	}

	for _, d := range fileTree.Root().Children {
		log.Trace.Println(d.Name(), d.IsFile())
	}

	return &fileTree, nil
}

// SyncComplete does nothing for this version
func (ctx *ApiCtx) SyncComplete() error {
	return nil
}
