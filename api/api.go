package api

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/juruen/rmapi/archive"
	"github.com/juruen/rmapi/filetree"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/transport"
	"github.com/juruen/rmapi/util"
	uuid "github.com/satori/go.uuid"
)

// An ApiCtx allows you interact with the remote reMarkable API
type ApiCtx struct {
	Http     *transport.HttpClientCtx
	Filetree *filetree.FileTreeCtx
}

// CreateApiCtx initializes an instance of ApiCtx
func CreateApiCtx(http *transport.HttpClientCtx) (*ApiCtx, error) {
	fileTree, err := DocumentsFileTree(http)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch document tree %v", err)
	}
	return &ApiCtx{http, fileTree}, nil
}

// DocumentsFileTree reads your remote documents and builds a file tree
// structure to represent them
func DocumentsFileTree(http *transport.HttpClientCtx) (*filetree.FileTreeCtx, error) {
	documents := make([]model.Document, 0)

	if err := http.Get(transport.UserBearer, listDocs, nil, &documents); err != nil {
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

// FetchDocument downloads a document given its ID and saves it locally into dstPath
func (ctx *ApiCtx) FetchDocument(docId, dstPath string) error {
	documents := make([]model.Document, 0)

	url := fmt.Sprintf("%s?withBlob=true&doc=%s", listDocs, docId)

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
func (ctx *ApiCtx) CreateDir(parentId, name string) (model.Document, error) {
	uploadRsp, err := ctx.uploadRequest("", model.DirectoryType, 1)

	if err != nil {
		return model.Document{}, err
	}

	if !uploadRsp.Success {
		return model.Document{}, errors.New("upload request returned success := false")
	}

	zippath, err := archive.CreateZipDirectory(uploadRsp.ID)

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

// DeleteEntry removes an entry: either an empty directory or a file
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

	err := ctx.Http.Put(transport.UserBearer, updateStatus, metaDoc, nil)

	if err != nil {
		log.Error.Println("failed to move entry", err)
		return nil, err
	}

	doc := metaDoc.ToDocument()

	return &model.Node{&doc, src.Children, dstDir}, nil
}

// UploadDocument uploads a local document given by sourceDocPath under the parentId directory
func (ctx *ApiCtx) UploadDocument(parent *model.Node, sourceDocPath string) (*model.Document, error) {
	name, ext := util.DocPathToName(sourceDocPath)

	if name == "" {
		return nil, errors.New("file name is invalid")
	}

	if !util.IsFileTypeSupported(ext) {
		return nil, errors.New("unsupported file extension: " + ext)
	}

	var metaDoc model.MetadataDocument

	docName, _ := util.DocPathToName(sourceDocPath)

	if node, err := ctx.Filetree.NodeByPath(docName, parent); err == nil {
		metaDoc = model.CreateUpdateDocumentMeta(
			node.Id(),
			model.DocumentType,
			parent.Id(),
			name,
			node.Version()+1,
		)
	} else {
		var id string
		if ext == "zip" {
			id, err = archive.GetIdFromZip(sourceDocPath)
			if err != nil {
				return nil, err
			}
			if id == "" {
				return nil, errors.New("could not determine the Document UUID")
			}
		} else {
			newId, err := uuid.NewV4()
			if err != nil {
				panic("failed to create uuid for directory")
			}

			id = newId.String()
		}
		metaDoc = model.CreateUploadDocumentMeta(id, model.DocumentType, parent.Id(), name)
	}

	//restore document
	uploadRsp, err := ctx.uploadRequest(metaDoc.ID, model.DocumentType, metaDoc.Version)

	if err != nil {
		return nil, err
	}

	if !uploadRsp.Success {
		return nil, fmt.Errorf("upload request failed: %s", uploadRsp.Message)
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

	err = ctx.Http.Put(transport.UserBearer, updateStatus, metaDoc, nil)

	if err != nil {
		log.Error.Println("failed to move entry", err)
		return nil, err
	}

	doc := metaDoc.ToDocument()

	return &doc, err
}

func (ctx *ApiCtx) uploadRequest(id string, entryType string, version int) (model.UploadDocumentResponse, error) {
	uploadReq := model.CreateUploadDocumentRequest(id, entryType, version)
	uploadRsp := make([]model.UploadDocumentResponse, 0)

	err := ctx.Http.Put(transport.UserBearer, uploadRequest, uploadReq, &uploadRsp)

	if err != nil {
		log.Error.Println("failed to to send upload request", err)
		return model.UploadDocumentResponse{}, err
	}

	return uploadRsp[0], nil
}
