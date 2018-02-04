package api

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/juruen/rmapi/log"
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
