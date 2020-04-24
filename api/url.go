package api

import "os"

var newTokenDevice string
var newUserDevice string
var docHost string
var listDocs string
var updateStatus string
var uploadRequest string
var deleteEntry string

func init() {
	docHost := "https://document-storage-production-dot-remarkable-production.appspot.com"
	authHost := "https://my.remarkable.com"

	host := os.Getenv("RMAPI_DOC")
	if host != "" {
		docHost = host
	}

	host = os.Getenv("RMAPI_AUTH")

	if host != "" {
		authHost = host
	}
	newTokenDevice = authHost + "/token/json/2/device/new"
	newUserDevice = authHost + "/token/json/2/user/new"
	listDocs = docHost + "/document-storage/json/2/docs"
	updateStatus = docHost + "/document-storage/json/2/upload/update-status"
	uploadRequest = docHost + "/document-storage/json/2/upload/request"
	deleteEntry = docHost + "/document-storage/json/2/delete"
}
