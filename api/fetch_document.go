package api

import (
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/transport"
	"github.com/juruen/rmapi/util"
)

func (ctx *ApiCtx) FetchDocument(docId, dstPath string) error {
	blobUrl, err := ctx.getDocBlobUrl(docId)

	if err != nil {
		return err
	}

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

func (ctx *ApiCtx) FetchHttpMetaDocument(docId string) (model.HttpDocumentMeta, error) {
	blobUrl, err := ctx.getDocBlobUrl(docId)

	if err != nil {
		return model.HttpDocumentMeta{}, err
	}

	response, err := ctx.Http.GetHeadResponse(transport.UserBearer, blobUrl)

	var md5Hash [md5.Size]byte
	for _, s := range response.Header["X-Goog-Hash"] {
		if strings.HasPrefix(s, "md5=") {
			decoded, err := base64.StdEncoding.DecodeString(s[4:])
			if err != nil {
				return model.HttpDocumentMeta{}, err
			}
			copy(md5Hash[:], decoded[:])
		}
	}

	httpMeta := model.HttpDocumentMeta{
		LastModified: response.Header.Get("Last-Modified"),
		Md5Hash:      md5Hash,
		Size:         response.ContentLength,
	}

	return httpMeta, nil
}

func (ctx *ApiCtx) getDocBlobUrl(docId string) (string, error) {
	documents := make([]model.Document, 0)

	url := fmt.Sprintf("%s?withBlob=true&doc=%s", listDocs, docId)

	if err := ctx.Http.Get(transport.UserBearer, url, nil, &documents); err != nil {
		log.Error.Println("failed to fetch document BlobURLGet %s", err)
		return "", err
	}

	if len(documents) == 0 || documents[0].BlobURLGet == "" {
		log.Error.Println("BlobURLGet for document is empty")
		return "", errors.New("no BlobURLGet")
	}

	return documents[0].BlobURLGet, nil
}
