package api

import (
	"encoding/json"
	"log"
	"time"

	uuid "github.com/satori/go.uuid"
)

const (
	DirectoryType = "CollectionType"
	DocumentType  = "DocumentType"
)

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

type MetadataDocument struct {
	ID             string
	Parent         string
	VissibleName   string
	Type           string
	Version        int
	ModifiedClient string
}

type DeleteDocument struct {
	ID      string
	Version int
}

type UploadDocumentRequest struct {
	ID      string
	Type    string
	Version int
}

type UploadDocumentResponse struct {
	ID                string
	Version           int
	Message           string
	Success           bool
	BlobURLPut        string
	BlobURLPutExpires string
}

func CreateDirDocument(parent, name string) MetadataDocument {
	id, err := uuid.NewV4()

	if err != nil {
		log.Panic("failed to create uuid for directory")
	}

	return MetadataDocument{
		ID:             id.String(),
		Parent:         parent,
		VissibleName:   name,
		Type:           DirectoryType,
		Version:        1,
		ModifiedClient: time.Now().Format(time.RFC3339Nano),
	}
}

func CreateUploadDocumentRequest() UploadDocumentRequest {
	id, err := uuid.NewV4()

	if err != nil {
		log.Panic("failed to create uuid for upload request")
	}

	return UploadDocumentRequest{
		id.String(),
		"DocumentType",
		1,
	}
}

func CreateUploadDocumentMeta(id, parent, name string) MetadataDocument {
	if parent == "1" {
		parent = ""
	}

	return MetadataDocument{
		ID:             id,
		Parent:         parent,
		VissibleName:   name,
		Type:           DocumentType,
		Version:        1,
		ModifiedClient: time.Now().Format(time.RFC3339Nano),
	}
}

func (meta MetadataDocument) Serialize() (string, error) {
	documents := ([]MetadataDocument{meta})
	serialized, err := json.Marshal(documents)

	if err != nil {
		return "", err
	}

	return string(serialized), nil
}

func (del DeleteDocument) Serialize() (string, error) {
	documents := ([]DeleteDocument{del})
	serialized, err := json.Marshal(documents)

	if err != nil {
		return "", err
	}

	return string(serialized), nil
}

func (req UploadDocumentRequest) Serialize() (string, error) {
	documents := ([]UploadDocumentRequest{req})
	serialized, err := json.Marshal(documents)

	if err != nil {
		return "", err
	}

	return string(serialized), nil
}

func (meta MetadataDocument) ToDocument() Document {
	return Document{
		ID:             meta.ID,
		Parent:         meta.Parent,
		VissibleName:   meta.VissibleName,
		Type:           meta.Type,
		Version:        1,
		ModifiedClient: meta.ModifiedClient,
	}
}

func (doc Document) ToMetaDocument() MetadataDocument {
	return MetadataDocument{
		ID:             doc.ID,
		Parent:         doc.Parent,
		VissibleName:   doc.VissibleName,
		Type:           doc.Type,
		Version:        doc.Version,
		ModifiedClient: time.Now().Format(time.RFC3339Nano),
	}
}

func (doc Document) ToDeleteDocument() DeleteDocument {
	return DeleteDocument{
		ID:      doc.ID,
		Version: doc.Version,
	}
}
