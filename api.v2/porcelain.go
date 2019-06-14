package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	// DirectoryType is used as directory type.
	DirectoryType = "CollectionType"
	// DocumentType is used a regular document type.
	DocumentType = "DocumentType"
)

// Document represents a human readable format of a Remarkable document.
// It is transformed into a more technical object and used as a
// payload or response during http calls with the Remarkable API.
type Document struct {
	ID          string
	Version     int
	Type        string
	Name        string
	CurrentPage int
	Bookmarked  bool
	Parent      string
}

// String dictates how to print a Document object.
func (d Document) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%s (%s)\n", d.Name, d.ID))
	buffer.WriteString(fmt.Sprintf("\tVersion: %d\n", d.Version))
	buffer.WriteString(fmt.Sprintf("\tType: %s\n", d.Type))
	buffer.WriteString(fmt.Sprintf("\tCurrent Page: %d\n", d.CurrentPage))
	buffer.WriteString(fmt.Sprintf("\tBookmarked: %t\n", d.Bookmarked))
	buffer.WriteString(fmt.Sprintf("\tParent: %s", d.Parent))
	return buffer.String()
}

// toRawDocument returns a technical rawDocument created from a public Document
func (d Document) toRawDocument() rawDocument {
	return rawDocument{
		ID:           d.ID,
		Version:      d.Version,
		Type:         d.Type,
		VissibleName: d.Name,
		CurrentPage:  d.CurrentPage,
		Bookmarked:   d.Bookmarked,
		Parent:       d.Parent,
	}
}

// Get is a first class method used to fetch information of a document.
//
// It takes a uuid as parameter and a Document is returned to the end user.
func (c *Client) Get(uuid string) (Document, error) {
	rdoc, err := c.getDoc(uuid)
	if err != nil {
		return Document{}, errors.Wrap(err, "can't get document")
	}

	return rdoc.toDocument(), nil
}

// List is a first class method used to fetch the list of all documents on a device.
//
// It returns a list of Documents.
func (c *Client) List() ([]Document, error) {
	// use empty uuid to have them all
	rdocs, err := c.getDocs("")
	if err != nil {
		return nil, errors.Wrap(err, "can't get documents")
	}

	var docs []Document
	for _, rdoc := range rdocs {
		docs = append(docs, rdoc.toDocument())
	}

	// sort by name
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].Name < docs[j].Name
	})

	return docs, nil
}

// Download is a first class method used to download the actual content of a document.
//
// It takes a uuid as parameter and an io.Writer that will be used to write the content.
// By using an io.Writer, the content can be committed to a file
// but also in response to an http call for example.
//
// The content received will be a zip file containing all the document files.
// To make use of it, you can have a look to the archive package.
func (c *Client) Download(uuid string, w io.Writer) error {
	rdoc, err := c.getDoc(uuid)
	if err != nil {
		return errors.Wrap(err, "can't get document")
	}

	// direct call to the http.Client.Get because different domain
	resp, err := c.httpClient.Get(rdoc.BlobURLGet)
	if err != nil {
		return errors.Wrap(err, "download failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("wrong http return code: %d", resp.StatusCode)
	}

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return errors.Wrap(err, "can't write file")
	}

	return nil
}

// Upload is a first class method used to upload content to a document.
//
// The Document given as parameter will identify which real document to
// target. If the Document uuid already exists, the content will be
// updated. If not, a new document will be created.
// For creating a new document however, you may want to use the Upload method instead.
//
// An io.Reader is expected as parameter to provide the content of the
// document. This way, we can upload a content not only from a file
// but as well from other sources.
//
// The content should be shaped as a zip file as expected by the Remarkable.
// You can have a look to the archive package to help easily creating
// a correctly formatted file.
func (c *Client) UploadDocument(doc Document, r io.Reader) error {
	if doc.ID == "" {
		return errors.New("undefined document id")
	}

	url, err := c.uploadRequest(doc.toRawDocument())
	if err != nil {
		return errors.Wrap(err, "can't create upload request")
	}

	// direct upload to url as endpoint is different
	req, err := http.NewRequest("PUT", url, r)
	if err != nil {
		return errors.Wrap(err, "can't initiate upload")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "can't upload document")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("wrong http return code: %d", resp.StatusCode)
	}

	if err := c.Metadata(doc); err != nil {
		return errors.Wrap(err, "can't update metadata")
	}

	return nil
}

// Upload is a first class method used to upload a new document.
//
// A newly generated uuid should be provided and a name should be given.
//
// An io.Reader is expected as parameter to provide the content of the
// document. This way, we can upload a content not only from a file
// but as well from other sources.
//
// The content should be shaped as a zip file as expected by the Remarkable.
// You can have a look to the archive package to help easily creating
// a correctly formatted file.
func (c *Client) Upload(uuid string, name string, r io.Reader) error {
	doc := Document{
		ID:      uuid,
		Type:    DocumentType,
		Name:    name,
		Version: 1,
	}

	return c.UploadDocument(doc, r)
}

// CreateFolder is a first class method used to create a new folder.
//
// It takes a folder name as parameter and an optional parent UUID to
// create the folder in a subdirectory. If an empty string is provided,
// the folder will be created at the root.
//
// The UUID of the created folder is returned.
func (c *Client) CreateFolder(name string, parent string) (string, error) {
	id := uuid.New().String()
	doc := Document{
		ID:      id,
		Parent:  parent,
		Type:    DirectoryType,
		Version: 1,
		Name:    name,
	}

	if err := c.Metadata(doc); err != nil {
		return "", errors.Wrap(err, "can't create folder")
	}

	return id, nil
}

// Metadata is a first class method used to update metadata of
// an existing document.
//
// It takes a Document as input containing the metadata changes
// that will be sent to the API.
//
// This method can be used to make a document visible, to move,
// rename, bookmark a document.
//
// Calling this method will reset the ModifiedClient time to now.
// If the document Version is not explicitly defined, it will be
// defined by fetching the current document version and adding one.
func (c *Client) Metadata(doc Document) error {
	rdoc := doc.toRawDocument()

	if rdoc.ID == "" {
		return errors.New("undefined document id")
	}

	// set modified to now
	rdoc.ModifiedClient = time.Now().Format(time.RFC3339Nano)

	// set Version to current version +1 if not defined
	if rdoc.Version == 0 {
		cur, err := c.getCurrentVersion(doc.ID)
		if err != nil {
			return errors.Wrap(err, "can't get current version of document")
		}
		rdoc.Version = cur + 1
	}

	payload := []rawDocument{rdoc}

	req, err := c.newRequest("PUT", "document-storage/json/2/upload/update-status", payload)
	if err != nil {
		return err
	}

	var rdocs []rawDocument
	resp, err := c.do(req, &rdocs)
	if err != nil {
		return errors.Wrap(err, "request failed")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("wrong http return code: %d", resp.StatusCode)
	}

	if len(rdocs) == 0 {
		return errors.Wrap(err, "empty document list received")
	}

	if !rdocs[0].Success {
		return errors.Errorf("success false received: %s", rdocs[0].Message)
	}

	return nil
}

// Delete is a first class method used to delete a document or folder.
func (c *Client) Delete(uuid string) error {
	cur, err := c.getCurrentVersion(uuid)
	if err != nil {
		return errors.Wrap(err, "can't get current version of document")
	}

	doc := Document{
		ID:      uuid,
		Version: cur,
	}

	payload := []rawDocument{doc.toRawDocument()}

	req, err := c.newRequest("PUT", "/document-storage/json/2/delete", payload)
	if err != nil {
		return err
	}

	resp, err := c.do(req, nil)
	if err != nil {
		return errors.Wrap(err, "request failed")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("wrong http return code: %d", resp.StatusCode)
	}

	return nil
}
