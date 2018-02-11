package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/util"
)

func (httpCtx *HttpClientCtx) DocumentsFileTree() *FileTreeCtx {
	documents := make([]Document, 0)

	if err := httpCtx.httpGet(UserBearer, listDocs, EmptyBody, &documents); err != nil {
		log.Error.Println("failed to fetch documents %s", err.Error())
		return nil
	}

	fileTree := CreateFileTreeCtx()

	for _, d := range documents {
		fileTree.AddDocument(d)
	}

	for _, d := range fileTree.root.Children {
		log.Trace.Println(d.Name(), d.IsFile())
	}

	return &fileTree
}

func (httpCtx *HttpClientCtx) FetchDocument(docId, dstPath string) error {
	documents := make([]Document, 0)

	url := fmt.Sprintf("%s?withBlob=true&doc=%s", listDocs, docId)

	if err := httpCtx.httpGet(UserBearer, url, EmptyBody, &documents); err != nil {
		log.Error.Println("failed to fetch document BlobURLGet %s", err)
		return err
	}

	if len(documents) == 0 || documents[0].BlobURLGet == "" {
		log.Error.Println("BlobURLGet for document is empty")
		return errors.New("no BlobURLGet")
	}

	blobUrl := documents[0].BlobURLGet

	src, err := httpCtx.httpGetStream(UserBearer, blobUrl, EmptyBody)

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

	err = os.Rename(tmpPath, dstPath)

	return nil
}

func (httpCtx *HttpClientCtx) CreateDir(parentId, name string) (Document, error) {
	metaDoc := CreateDirDocument(parentId, name)
	metaBody, err := metaDoc.Serialize()

	if err != nil {
		return Document{}, err
	}

	_, err = httpCtx.httpPutRaw(UserBearer, updateStatus, metaBody)

	if err != nil {
		log.Error.Println("failed to create a new device directory", metaBody, err)
		return Document{}, err
	}

	return metaDoc.ToDocument(), nil
}

func (httpCtx *HttpClientCtx) DeleteEntry(node *Node) error {
	if node.IsDirectory() && len(node.Children) > 0 {
		return errors.New("directory is not empty")
	}

	del := node.Document.ToDeleteDocument()
	delBody, err := del.Serialize()

	_, err = httpCtx.httpPutRaw(UserBearer, deleteEntry, delBody)

	if err != nil {
		log.Error.Println("failed to remove entry", err)
		return err
	}

	return nil
}

func (httpCtx *HttpClientCtx) MoveEntry(src *Node, dstDir *Node, name string) (*Node, error) {
	if dstDir.IsFile() {
		return nil, errors.New("destination directory is a file")
	}

	metaDoc := src.Document.ToMetaDocument()
	metaDoc.Version = metaDoc.Version + 1
	metaDoc.VissibleName = name

	if dstDir.IsRoot() {
		metaDoc.Parent = ""
	} else {
		metaDoc.Parent = dstDir.Id()
	}

	metaBody, err := metaDoc.Serialize()

	if err != nil {
		log.Error.Println("failed to move entry", err)
		return nil, err
	}

	_, err = httpCtx.httpPutRaw(UserBearer, updateStatus, metaBody)

	if err != nil {
		log.Error.Println("failed to move entry", metaBody, err)
		return nil, err
	}

	doc := metaDoc.ToDocument()

	if doc.Parent == "" {
		doc.Parent = "1"
	}

	return &Node{&doc, src.Children, dstDir}, nil
}

func (httpCtx *HttpClientCtx) UploadDocument(parent string, pdfpath string) (*Document, error) {
	name := util.PdfPathToName(pdfpath)

	if name == "" {
		return nil, errors.New("file name is invalid")
	}

	uploadRsp, err := httpCtx.uploadRequest()

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

	_, err = httpCtx.httpPutStream(UserBearer, uploadRsp.BlobURLPut, f)

	if err != nil {
		log.Error.Println("failed to upload zip document", err)
		return nil, err
	}

	metaDoc := CreateUploadDocumentMeta(uploadRsp.ID, parent, name)
	metaBody, err := metaDoc.Serialize()

	if err != nil {
		log.Error.Println("failed to serialize uplaod meta request", err)
		return nil, err
	}

	_, err = httpCtx.httpPutRaw(UserBearer, updateStatus, metaBody)

	if err != nil {
		log.Error.Println("failed to move entry", metaBody, err)
		return nil, err
	}

	doc := metaDoc.ToDocument()

	return &doc, err
}

func (httpCtx *HttpClientCtx) uploadRequest() (UploadDocumentResponse, error) {
	uploadReq := CreateUploadDocumentRequest()
	uploadReqBody, err := uploadReq.Serialize()
	uploadRsp := make([]UploadDocumentResponse, 1)

	if err != nil {
		log.Error.Println("failed to serilize upload request", err)
		return uploadRsp[0], err
	}

	resp, err := httpCtx.httpPutRaw(UserBearer, uploadRequest, string(uploadReqBody))

	if err != nil {
		log.Error.Println("failed to to send upload request", err)
		return uploadRsp[0], err
	}

	err = json.Unmarshal([]byte(resp), &uploadRsp)

	if err != nil {
		log.Error.Println("failed to to deserialize upload request", err)
		return uploadRsp[0], err
	}

	return uploadRsp[0], nil
}
