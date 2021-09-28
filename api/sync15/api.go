package sync15

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/google/uuid"
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
	r    RemoteStorage
	t    *Tree
}

func (ctx *ApiCtx) Filetree() *filetree.FileTreeCtx {
	return ctx.ft
}

// Nuke removes all documents from the account
func (ctx *ApiCtx) Nuke() error {
	documents := make([]model.Document, 0)

	if err := ctx.Http.Get(transport.UserBearer, config.ListDocs, nil, &documents); err != nil {
		return err
	}

	for _, d := range documents {
		log.Info.Println("Deleting: ", d.VissibleName)

		err := ctx.Http.Put(transport.UserBearer, config.DeleteEntry, d, nil)
		if err != nil {
			log.Error.Println("failed to remove entry", err)
			return err
		}
	}

	return nil
}

// FetchDocument downloads a document given its ID and saves it locally into dstPath
func (ctx *ApiCtx) FetchDocument(docId, dstPath string) error {

	doc, err := ctx.t.FindDoc(docId)
	if err != nil {
		return err
	}

	dst, err := ioutil.TempDir("", "rmapifile")
	if err != nil {
		return err
	}
	log.Info.Print("got files in ", dst)
	for _, f := range doc.Files {
		log.Info.Println(f.DocumentID)
		rc, err := ctx.r.GetReader(f.Hash)
		if err != nil {
			return err
		}
		defer rc.Close()
		fp := path.Join(dst, f.DocumentID)
		root := path.Dir(fp)
		log.Info.Println(root)
		err = os.MkdirAll(root, 0744)
		if err != nil {
			return fmt.Errorf("cant create forld %v", err)
		}
		nf, err := os.Create(fp)
		if err != nil {
			return err
		}
		_, err = io.Copy(nf, rc)
		if err != nil {
			return err
		}

	}
	// defer os.RemoveAll(dst)
	log.Info.Print("got files in ", dst)

	// _, err = util.CopyFile(tmpPath, dstPath)

	// if err != nil {
	// 	log.Error.Printf("failed to copy %s to %s, er: %s\n", tmpPath, dstPath, err.Error())
	// 	return err
	// }

	return nil
}

// CreateDir creates a remote directory with a given name under the parentId directory
func (ctx *ApiCtx) CreateDir(parentId, name string) (model.Document, error) {
	uploadRsp, err := ctx.uploadRequest("", model.DirectoryType)

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

	if err != nil {
		log.Error.Println("failed to read zip file to upload", zippath, err)
		return model.Document{}, err
	}

	defer f.Close()

	err = ctx.Http.PutStream(transport.UserBearer, uploadRsp.BlobURLPut, f)

	if err != nil {
		log.Error.Println("failed to upload directory", err)
		return model.Document{}, err
	}

	metaDoc := model.CreateUploadDocumentMeta(uploadRsp.ID, model.DirectoryType, parentId, name)

	err = ctx.Http.Put(transport.UserBearer, config.UpdateStatus, metaDoc, nil)

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

	err := ctx.Http.Put(transport.UserBearer, config.DeleteEntry, deleteDoc, nil)

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

	err := ctx.Http.Put(transport.UserBearer, config.UpdateStatus, metaDoc, nil)

	if err != nil {
		log.Error.Println("failed to move entry", err)
		return nil, err
	}

	doc := metaDoc.ToDocument()

	return &model.Node{&doc, src.Children, dstDir}, nil
}

// UploadDocument uploads a local document given by sourceDocPath under the parentId directory
func (ctx *ApiCtx) UploadDocument(parentId string, sourceDocPath string) (*model.Document, error) {
	name, ext := util.DocPathToName(sourceDocPath)

	if name == "" {
		return nil, errors.New("file name is invalid")
	}

	if !util.IsFileTypeSupported(ext) {
		return nil, errors.New("unsupported file extension: " + ext)
	}

	id := ""
	var err error

	//TODO extract
	if ext == "zip" {
		id, err = archive.GetIdFromZip(sourceDocPath)
		if err != nil {
			return nil, err
		}
		if id == "" {
			return nil, errors.New("could not determine the Document UUID")
		}
	} else {
		id = uuid.New().String()
	}

	// zipPath, err := archive.CreateZipDocument(id, sourceDocPath)

	doc := model.Document{}

	//update root
	//notify

	return &doc, err
}

func (ctx *ApiCtx) uploadRequest(id string, entryType string) (model.UploadDocumentResponse, error) {
	uploadReq := model.CreateUploadDocumentRequest(id, entryType)
	uploadRsp := make([]model.UploadDocumentResponse, 0)

	err := ctx.Http.Put(transport.UserBearer, config.UploadRequest, uploadReq, &uploadRsp)

	if err != nil {
		log.Error.Println("failed to to send upload request", err)
		return model.UploadDocumentResponse{}, err
	}

	return uploadRsp[0], nil
}

func CreateCtx(http *transport.HttpClientCtx) (*ApiCtx, error) {
	apiStorage := &BlobStorage{http}
	cacheTree, err := loadTree()
	if err != nil {
		fmt.Print(err)
		return nil, err
	}
	err = cacheTree.Sync(apiStorage)
	if err != nil {
		return nil, err
	}
	saveTree(cacheTree)
	tree, err := DocumentsFileTree(cacheTree)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch document tree %v", err)
	}
	return &ApiCtx{http, tree, apiStorage, cacheTree}, nil
}

// DocumentsFileTree reads your remote documents and builds a file tree
// structure to represent them
func DocumentsFileTree(tree *Tree) (*filetree.FileTreeCtx, error) {

	documents := make([]model.Document, 0)
	for _, d := range tree.Docs {
		doc := *d.ToDocument()
		documents = append(documents, doc)
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
