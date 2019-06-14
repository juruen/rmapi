package api

import (
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

// rawDocument represents a real object expected in most calls
// by the Remarkable API.
type rawDocument struct {
	ID                string `json:"ID"`
	Version           int    `json:"Version"`
	Message           string `json:"Message"`
	Success           bool   `json:"Success"`
	BlobURLGet        string `json:"BlobURLGet"`
	BlobURLGetExpires string `json:"BlobURLGetExpires"`
	BlobURLPut        string `json:"BlobURLPut"`
	BlobURLPutExpires string `json:"BlobURLPutExpires"`
	ModifiedClient    string `json:"ModifiedClient"`
	Type              string `json:"Type"`
	VissibleName      string `json:"VissibleName"`
	CurrentPage       int    `json:"CurrentPage"`
	Bookmarked        bool   `json:"Bookmarked"`
	Parent            string `json:"Parent"`
}

// toDocument transforms a rawDocument to a
// cleaner public Document
func (r rawDocument) toDocument() Document {
	return Document{
		ID:          r.ID,
		Version:     r.Version,
		Type:        r.Type,
		Name:        r.VissibleName,
		CurrentPage: r.CurrentPage,
		Bookmarked:  r.Bookmarked,
		Parent:      r.Parent,
	}
}

// getDocs makes a call to the Remarkable API in order to get
// a list of documents present on a device.
// urlParams is a string representing optional query string parameters.
// uuid can be used to filter the request to a single document.
// withBlob can be used to indicate that a download url should be given as return.
func (c *Client) getDocs(urlParams string) ([]rawDocument, error) {
	req, err := c.newRequest("GET", "document-storage/json/2/docs", nil)
	if err != nil {
		return nil, err
	}

	// add query string parameters
	req.URL.RawQuery = urlParams

	var docs []rawDocument
	resp, err := c.do(req, &docs)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("wrong http return code: %d", resp.StatusCode)
	}

	return docs, err
}

// getDoc calls getDocs by filtering to a precise uuid and
// by including a withBlob=true parameter to include the download url as return.
func (c *Client) getDoc(uuid string) (rawDocument, error) {
	v := url.Values{}
	v.Add("doc", uuid)
	// assume we always want to have the download url in response
	v.Add("withBlob", "true")

	rdocs, err := c.getDocs(v.Encode())
	if err != nil {
		return rawDocument{}, errors.Wrap(err, "can't retrieve documents")
	}

	if len(rdocs) == 0 {
		return rawDocument{}, errors.Wrap(err, "empty document list received")
	}

	if !rdocs[0].Success {
		return rawDocument{}, errors.Errorf("success false received: %s", rdocs[0].Message)
	}

	return rdocs[0], nil
}

// uploadRequest makes an initial request to the Remarkable API to start a
// document upload.
// The doc parameter is used to configure the upload.
// If it contains a new uuid, it will create an upload for a new document.
// If it contains an existing uuid, it will try to upload another version
// of a document. For the latter to work, the Version parameter of the doc should
// be increased.
// As return, uploadRequest will give a URL that can be used for uploading the actual
// content of the document.
func (c *Client) uploadRequest(doc rawDocument) (string, error) {
	payload := []rawDocument{doc}

	req, err := c.newRequest("PUT", "document-storage/json/2/upload/request", payload)
	if err != nil {
		return "", err
	}

	var rdocs []rawDocument
	resp, err := c.do(req, &rdocs)
	if err != nil {
		return "", errors.Wrap(err, "request failed")
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("wrong http return code: %d", resp.StatusCode)
	}

	if len(rdocs) == 0 {
		return "", errors.Wrap(err, "empty document list received")
	}

	if !rdocs[0].Success {
		return "", errors.Errorf("success false received: %s", rdocs[0].Message)
	}

	if rdocs[0].BlobURLPut == "" {
		return "", errors.New("empty upload url received")
	}

	return rdocs[0].BlobURLPut, nil
}

// getCurrentVersion makes an http call to the Remarkable API to
// fetch a document from a uuid and return its current version.
func (c *Client) getCurrentVersion(uuid string) (int, error) {
	rdoc, err := c.getDoc(uuid)
	if err != nil {
		return 0, errors.Wrap(err, "can't get document")
	}
	return rdoc.Version, nil
}
