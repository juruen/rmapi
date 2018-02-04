package api

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

func (meta MetadataDocument) ToDocument() Document {
	return Document{
		ID:             meta.ID,
		Parent:         meta.Parent,
		VissibleName:   meta.VissibleName,
		Type:           DirectoryType,
		Version:        1,
		ModifiedClient: meta.ModifiedClient,
	}
}
