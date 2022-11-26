package config

import "os"

var NewTokenDevice string
var NewUserDevice string
var DocHost string
var ListDocs string
var UpdateStatus string
var UploadRequest string
var DeleteEntry string
var UploadBlob string
var DownloadBlob string
var SyncComplete string

func init() {
	docHost := "https://document-storage-production-dot-remarkable-production.appspot.com"
	authHost := "https://webapp-prod.cloud.remarkable.engineering"
	syncHost := "https://internal.cloud.remarkable.com"

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
		authHost = host
		docHost = host
		syncHost = host
	}

	NewTokenDevice = authHost + "/token/json/2/device/new"
	NewUserDevice = authHost + "/token/json/2/user/new"
	ListDocs = docHost + "/document-storage/json/2/docs"
	UpdateStatus = docHost + "/document-storage/json/2/upload/update-status"
	UploadRequest = docHost + "/document-storage/json/2/upload/request"
	DeleteEntry = docHost + "/document-storage/json/2/delete"

	UploadBlob = syncHost + "/sync/v2/signed-urls/uploads"
	DownloadBlob = syncHost + "/sync/v2/signed-urls/downloads"
	SyncComplete = syncHost + "/sync/v2/sync-complete"
}
