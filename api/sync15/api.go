package sync15

import (
	"archive/zip"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/juruen/rmapi/archive"
	"github.com/juruen/rmapi/config"
	"github.com/juruen/rmapi/filetree"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/transport"
	"github.com/juruen/rmapi/util"
)

var ErrorNotImplemented = errors.New("not implemented")

// An ApiCtx allows you interact with the remote reMarkable API
type ApiCtx struct {
	Http *transport.HttpClientCtx
	ft   *filetree.FileTreeCtx
	r    *BlobStorage
	t    *Tree
}

func (ctx *ApiCtx) Filetree() *filetree.FileTreeCtx {
	return ctx.ft
}

// Nuke removes all documents from the account
func (ctx *ApiCtx) Nuke() error {
	return ErrorNotImplemented
}

// FetchDocument downloads a document given its ID and saves it locally into dstPath
func (ctx *ApiCtx) FetchDocument(docId, dstPath string) error {

	doc, err := ctx.t.FindDoc(docId)
	if err != nil {
		return err
	}

	tmp, err := ioutil.TempFile("", "rmapizip")

	if err != nil {
		log.Error.Println("failed to create tmpfile for zip dir", err)
		return err
	}
	defer tmp.Close()

	w := zip.NewWriter(tmp)
	defer w.Close()
	for _, f := range doc.Files {
		log.Info.Println(f.DocumentID)
		blobReader, err := ctx.r.GetReader(f.Hash)
		if err != nil {
			return err
		}
		defer blobReader.Close()
		header := zip.FileHeader{}
		header.Name = f.DocumentID
		header.Modified = time.Now()
		zipWriter, err := w.CreateHeader(&header)
		if err != nil {
			return err
		}
		_, err = io.Copy(zipWriter, blobReader)

		if err != nil {
			return err
		}
	}
	w.Close()
	tmpPath := tmp.Name()
	_, err = util.CopyFile(tmpPath, dstPath)

	if err != nil {
		log.Error.Printf("failed to copy %s to %s, er: %s\n", tmpPath, dstPath, err.Error())
		return err
	}

	defer os.RemoveAll(tmp.Name())

	return nil
}

// CreateDir creates a remote directory with a given name under the parentId directory
func (ctx *ApiCtx) CreateDir(parentId, name string) (*model.Document, error) {
	return nil, ErrorNotImplemented
}

func Sync(b *BlobStorage, tree *Tree, operation func(t *Tree)) error {
	synccount := 0
	for {
		synccount++
		if synccount > 10 {
			log.Error.Println("Something is wrong")
			break
		}
		log.Info.Println("Uploading...")
		operation(tree)

		indexReader, err := tree.IndexReader()
		if err != nil {
			return err
		}
		err = b.UploadBlob(tree.Hash, indexReader)
		if err != nil {
			return err
		}
		defer indexReader.Close()

		gen, err := b.WriteRootIndex(tree.Hash, tree.Generation)

		if err == nil {
			tree.Generation = gen
			break
		}

		if err != transport.ErrWrongGeneration {
			return err
		}

		//resync and try again
		err = tree.Sync(b)
		if err != nil {
			return err
		}
	}
	err := saveTree(tree)
	if err != nil {
		return err
	}
	err = b.SyncComplete()
	if err != nil {
		log.Error.Printf("cannot send sync %v", err)
	}
	return err
}

// DeleteEntry removes an entry: either an empty directory or a file
func (ctx *ApiCtx) DeleteEntry(node *model.Node) error {
	if node.IsDirectory() && len(node.Children) > 0 {
		return errors.New("directory is not empty")
	}

	//_ := node.Document.ToDeleteDocument()

	err := Sync(ctx.r, ctx.t, func(t *Tree) {
		t.Remove(node.Document.ID)
	})
	return err

}

// MoveEntry moves an entry (either a directory or a file)
// - src is the source node to be moved
// - dstDir is an existing destination directory
// - name is the new name of the moved entry in the destination directory
func (ctx *ApiCtx) MoveEntry(src, dstDir *model.Node, name string) (*model.Node, error) {
	if dstDir.IsFile() {
		return nil, errors.New("destination directory is a file")
	}

	// metaDoc := src.Document.ToMetaDocument()
	// metaDoc.Version = metaDoc.Version + 1
	// metaDoc.VissibleName = name
	// metaDoc.Parent = dstDir.Id()

	// err := ctx.Http.Put(transport.UserBearer, config.UpdateStatus, metaDoc, nil)

	// if err != nil {
	// 	log.Error.Println("failed to move entry", err)
	// 	return nil, err
	// }

	// doc := metaDoc.ToDocument()

	// return &model.Node{&doc, src.Children, dstDir}, nil
	return nil, ErrorNotImplemented
}

type FileStuff struct {
	Name string
	Path string
}

func AddStuff(f *[]FileStuff, name, filepath string) {
	fs := FileStuff{
		Name: name,
		Path: filepath,
	}
	*f = append(*f, fs)
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

	files := []FileStuff{}

	tmpDir, err := ioutil.TempDir("", "rmupload")
	if err != nil {
		return nil, err
	}
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
		objectId := id + "." + ext
		doctype := ext
		if ext == "rm" {
			pageId := uuid.New().String()
			objectId = fmt.Sprintf("%s/%s.rm", id, pageId)
			doctype = "notebook"
		}
		AddStuff(&files, objectId, sourceDocPath)
		objectId, filePath, err := archive.CreateMetadata(id, name, parentId, tmpDir)
		if err != nil {
			return nil, err
		}
		AddStuff(&files, objectId, filePath)

		objectId, filePath, err = archive.CreateContent(id, doctype, tmpDir)
		if err != nil {
			return nil, err
		}
		AddStuff(&files, objectId, filePath)
	}

	d := &Doc{
		MetadataFile: archive.MetadataFile{
			DocName: name,
		},
		Entry: Entry{
			DocumentID: id,
		},
	}
	for _, f := range files {
		log.Info.Printf("File %s, path: %s", f.Name, f.Path)
		hash, err := FileHash(f.Path)
		if err != nil {
			return nil, err
		}
		hashStr := hex.EncodeToString(hash)
		e := &Entry{
			DocumentID: f.Name,
			Hash:       hashStr,
			Type:       FileType,
			Size:       123,
		}
		reader, err := os.Open(f.Path)
		if err != nil {
			return nil, err
		}
		err = ctx.r.UploadBlob(hashStr, reader)

		if err != nil {
			return nil, err
		}

		d.AddFile(e)
	}
	indexReader, err := d.IndexReader()
	if err != nil {
		return nil, err
	}
	defer indexReader.Close()
	log.Info.Println("Uploading doc index...", d.Hash)
	err = ctx.r.UploadBlob(d.Hash, indexReader)
	if err != nil {
		return nil, err
	}
	tree := ctx.t

	synccount := 0
	for {
		synccount++
		if synccount > 10 {
			log.Error.Println("Something is wrong")
			break
		}
		log.Info.Println("Uploading...")
		tree.Add(d)

		indexReader, err := tree.IndexReader()
		if err != nil {
			return nil, err
		}
		err = ctx.r.UploadBlob(tree.Hash, indexReader)
		if err != nil {
			return nil, err
		}
		defer indexReader.Close()

		gen, err := ctx.r.WriteRootIndex(tree.Hash, tree.Generation)

		if err == nil {
			tree.Generation = gen
			break
		}

		if err != transport.ErrWrongGeneration {
			return nil, err
		}

		//resync and try again
		err = tree.Sync(ctx.r)
		if err != nil {
			return nil, err
		}
	}
	err = saveTree(ctx.t)
	if err != nil {
		return nil, err
	}
	err = ctx.r.SyncComplete()
	if err != nil {
		log.Error.Printf("cannot send sync %v", err)
	}

	return d.ToDocument(), err
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

	documents := make([]*model.Document, 0)
	for _, d := range tree.Docs {
		doc := d.ToDocument()
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
