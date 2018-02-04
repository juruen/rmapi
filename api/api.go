package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/juruen/rmapi/log"
	"github.com/satori/go.uuid"
)

type deviceTokenRequest struct {
	Code       string `json:"code"`
	DeviceDesc string `json:"deviceDesc"`
	DeviceId   string `json:"deviceID"`
}

type Document struct {
	ID                string
	Version           int
	Message           string
	Success           bool
	BlobURLGet        string
	BlobURLGetExpires string
	ModifiedClient    string
	Type              string
	VissibleName      string
	CurrentPage       int
	Bookmarked        bool
	Parent            string
}

const (
	defaultDeviceDesc string = "desktop-linux"
)

func (httpCtx *HttpClientCtx) newDeviceToken(code string) (string, error) {
	uuid, err := uuid.NewV4()

	if err != nil {
		panic(err)
	}

	body, err := json.Marshal(deviceTokenRequest{code, defaultDeviceDesc, uuid.String()})

	log.Trace.Println("body: ", string(body))

	if err != nil {
		panic(err)
	}

	resp, err := httpCtx.httpPostRaw(EmptyBearer, newTokenDevice, string(body))

	if err != nil {
		log.Error.Fatal("failed to create a new device token")

		return "", err
	}

	return resp, nil
}

func (httpCtx *HttpClientCtx) newUserToken() (string, error) {
	resp, err := httpCtx.httpPostRaw(DeviceBearer, newUserDevice, "")

	if err != nil {
		log.Error.Fatal("failed to create a new user token")

		return "", err
	}

	return resp, nil
}

func (httpCtx *HttpClientCtx) DocumentsFileTree() *FileTreeCtx {
	documents := make([]Document, 0)

	if err := httpCtx.httpGet(UserBearer, listDocs, EmptyBody, &documents); err != nil {
		log.Error.Println("failed to fetch documents %s", err.Error())
		return nil
	}

	fileTree := CreateFileTreeCtx()

	for _, d := range documents {
		fileTree.addDocument(d)
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
