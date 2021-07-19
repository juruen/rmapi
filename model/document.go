package model

import (
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
		ModifiedClient: time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func CreateUploadDocumentRequest(id string, entryType string, version int) UploadDocumentRequest {
	if id == "" {
		newId, err := uuid.NewV4()

		if err != nil {
			log.Panic("failed to create uuid for directory")
		}
		id = newId.String()
	}

	return UploadDocumentRequest{
		id,
		entryType,
		version,
	}
}

func CreateUploadDocumentMeta(id string, entryType, parent, name string) MetadataDocument {
	return MetadataDocument{
		ID:             id,
		Parent:         parent,
		VissibleName:   name,
		Type:           entryType,
		Version:        1,
		ModifiedClient: time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func CreateUpdateDocumentMeta(id string, entryType, parent, name string, version int) MetadataDocument {
	return MetadataDocument{
		ID:             id,
		Parent:         parent,
		VissibleName:   name,
		Type:           entryType,
		Version:        version,
		ModifiedClient: time.Now().UTC().Format(time.RFC3339Nano),
	}
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
		ModifiedClient: time.Now().UTC().Format(time.RFC3339Nano),
	}
}

func (doc Document) ToDeleteDocument() DeleteDocument {
	return DeleteDocument{
		ID:      doc.ID,
		Version: doc.Version,
	}
}
