package main

import (
	"encoding/json"
	"fmt"

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

const (
	newTokenDevice string = "https://my.remarkable.com/token/device/new"
	newUserDevice  string = "https://my.remarkable.com/token/user/new"
	docHost        string = "https://document-storage-production-dot-remarkable-production.appspot.com"
	listDocs       string = docHost + "/document-storage/json/2/docs"
)

func (httpCtx *HttpClientCtx) newDeviceToken(code string) (string, error) {
	uuid, err := uuid.NewV4()

	if err != nil {
		panic(err)
	}

	body, err := json.Marshal(deviceTokenRequest{code, defaultDeviceDesc, uuid.String()})

	Warning.Println("body: ", string(body))

	if err != nil {
		panic(err)
	}

	resp, err := httpCtx.httpPostRaw(EmptyBearer, newTokenDevice, string(body))

	if err != nil {
		Error.Fatal("failed to create a new device token")

		return "", err
	}

	return resp, nil
}

func (httpCtx *HttpClientCtx) newUserToken() (string, error) {
	resp, err := httpCtx.httpPostRaw(DeviceBearer, newUserDevice, "")

	if err != nil {
		Error.Fatal("failed to create a new user token")

		return "", err
	}

	return resp, nil
}

func (httpCtx *HttpClientCtx) listDocuments() {
	documents := make([]Document, 0)

	if err := httpCtx.httpGet(UserBearer, listDocs, EmptyBody, &documents); err != nil {
		Error.Println("failed to fetch documents")
		return
	}

	for _, d := range documents {
		fmt.Printf("%s = {%s, %s}\n", d.VissibleName, d.ID, d.Parent)
	}
}
