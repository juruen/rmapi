package api

const (
	newTokenDevice string = "https://my.remarkable.com/token/json/2/device/new"
	newUserDevice  string = "https://my.remarkable.com/token/json/2/user/new"
	docHost        string = "https://document-storage-production-dot-remarkable-production.appspot.com"
	listDocs       string = docHost + "/document-storage/json/2/docs"
	updateStatus   string = docHost + "/document-storage/json/2/upload/update-status"
	uploadRequest  string = docHost + "/document-storage/json/2/upload/request"
	deleteEntry    string = docHost + "/document-storage/json/2/delete"
)
