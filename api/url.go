package api

const (
	newTokenDevice string = "https://my.remarkable.com/token/device/new"
	newUserDevice  string = "https://my.remarkable.com/token/user/new"
	docHost        string = "https://document-storage-production-dot-remarkable-production.appspot.com"
	listDocs       string = docHost + "/document-storage/json/2/docs"
	updateStatus   string = docHost + "/document-storage/json/2/upload/update-status"
	upload         string = docHost + "/document-storage/json/2/upload/request"
	deleteEntry    string = docHost + "/document-storage/json/2/delete"
)
