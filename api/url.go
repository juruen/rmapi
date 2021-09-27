package api

import "os"

var newTokenDevice string
var newUserDevice string
var docHost string
var listDocs string
var updateStatus string
var uploadRequest string
var deleteEntry string
var uploadBlob string
var downloadBlob string
var syncComplete string

//NOT USED!!!
func init() {
	docHost := "https://document-storage-production-dot-remarkable-production.appspot.com"
	authHost := "https://webapp-production-dot-remarkable-production.appspot.com"
	blobHost := "https://rm-storage.appspot.com"

	host := os.Getenv("RMAPI_DOC")
	if host != "" {
		docHost = host
	}

	host = os.Getenv("RMAPI_AUTH")

	if host != "" {
		authHost = host
	}

	host = os.Getenv("RMAPI_HOST")
	if host != "" {
		docHost = host
		authHost = host
		blobHost = host
	}

	newTokenDevice = authHost + "/token/json/2/device/new"
	newUserDevice = authHost + "/token/json/2/user/new"
	listDocs = docHost + "/document-storage/json/2/docs"
	updateStatus = docHost + "/document-storage/json/2/upload/update-status"
	uploadRequest = docHost + "/document-storage/json/2/upload/request"
	deleteEntry = docHost + "/document-storage/json/2/delete"

	uploadBlob = docHost + "/api/v1/signed-urls/upload"
	downloadBlob = docHost + "/api/v1/signed-urls/download"
	syncComplete = blobHost + "/api/v1/sync-complete"

}
