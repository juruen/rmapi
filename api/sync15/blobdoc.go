package sync15

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/juruen/rmapi/archive"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
)

type BlobDoc struct {
	Files []*Entry
	Entry
	Metadata archive.MetadataFile
}

func NewBlobDoc(name, documentId, colType, parentId string) *BlobDoc {
	return &BlobDoc{
		Metadata: archive.MetadataFile{
			DocName:        name,
			CollectionType: colType,
			LastModified:   archive.UnixTimestamp(),
			Parent:         parentId,
		},
		Entry: Entry{
			DocumentID: documentId,
		},
	}

}

func (d *BlobDoc) Rehash() error {

	hash, err := HashEntries(d.Files)
	if err != nil {
		return err
	}
	log.Trace.Println("New doc hash: ", hash)
	d.Hash = hash
	return nil
}

func (d *BlobDoc) MetadataHashAndReader() (hash string, reader io.Reader, err error) {
	jsn, err := json.Marshal(d.Metadata)
	if err != nil {
		return
	}
	sha := sha256.New()
	sha.Write(jsn)
	hash = hex.EncodeToString(sha.Sum(nil))
	log.Trace.Println("new hash", hash)
	reader = bytes.NewReader(jsn)
	found := false
	for _, f := range d.Files {
		if strings.HasSuffix(f.DocumentID, ".metadata") {
			f.Hash = hash
			found = true
			break
		}
	}
	if !found {
		err = errors.New("metadata not found")
	}

	return
}

func (d *BlobDoc) AddFile(e *Entry) error {
	d.Files = append(d.Files, e)
	return d.Rehash()
}

func (t *HashTree) Add(d *BlobDoc) error {
	if len(d.Files) == 0 {
		return errors.New("no files")
	}
	t.Docs = append(t.Docs, d)
	return t.Rehash()
}

func (d *BlobDoc) IndexReader() (io.ReadCloser, error) {
	if len(d.Files) == 0 {
		return nil, errors.New("no files")
	}
	pipeReader, pipeWriter := io.Pipe()
	w := bufio.NewWriter(pipeWriter)
	go func() {
		defer pipeWriter.Close()
		w.WriteString(SchemaVersion)
		w.WriteString("\n")
		for _, d := range d.Files {
			w.WriteString(d.Line())
			w.WriteString("\n")
		}
		w.Flush()
	}()

	return pipeReader, nil
}

// ReadMetadata the document metadata from remote blob
func (d *BlobDoc) ReadMetadata(fileEntry *Entry, r RemoteStorage) error {
	if strings.HasSuffix(fileEntry.DocumentID, ".metadata") {
		log.Trace.Println("Reading metadata: " + d.DocumentID)

		metadata := archive.MetadataFile{}

		meta, err := r.GetReader(fileEntry.Hash)
		if err != nil {
			return err
		}
		defer meta.Close()
		content, err := io.ReadAll(meta)
		if err != nil {
			return err
		}
		err = json.Unmarshal(content, &metadata)
		if err != nil {
			log.Error.Printf("cannot read metadata %s %v", fileEntry.DocumentID, err)
		}
		log.Trace.Println("name from metadata: ", metadata.DocName)
		d.Metadata = metadata
	}

	return nil
}

func (d *BlobDoc) Line() string {
	var sb strings.Builder
	if d.Hash == "" {
		log.Error.Print("missing hash for: ", d.DocumentID)
	}
	sb.WriteString(d.Hash)
	sb.WriteRune(Delimiter)
	sb.WriteString(DocType)
	sb.WriteRune(Delimiter)
	sb.WriteString(d.DocumentID)
	sb.WriteRune(Delimiter)

	numFilesStr := strconv.Itoa(len(d.Files))
	sb.WriteString(numFilesStr)
	sb.WriteRune(Delimiter)
	sb.WriteString("0")
	return sb.String()
}

// Mirror updates the document to be the same as the remote
func (d *BlobDoc) Mirror(e *Entry, r RemoteStorage) error {
	d.Entry = *e
	entryIndex, err := r.GetReader(e.Hash)
	if err != nil {
		return err
	}
	defer entryIndex.Close()
	entries, err := parseIndex(entryIndex)
	if err != nil {
		return err
	}

	head := make([]*Entry, 0)
	current := make(map[string]*Entry)
	new := make(map[string]*Entry)

	for _, e := range entries {
		new[e.DocumentID] = e
	}

	//updated and existing
	for _, currentEntry := range d.Files {
		if newEntry, ok := new[currentEntry.DocumentID]; ok {
			if newEntry.Hash != currentEntry.Hash {
				err = d.ReadMetadata(newEntry, r)
				if err != nil {
					return err
				}
				currentEntry.Hash = newEntry.Hash
			}
			head = append(head, currentEntry)
			current[currentEntry.DocumentID] = currentEntry
		}
	}

	//add missing
	for k, newEntry := range new {
		if _, ok := current[k]; !ok {
			err = d.ReadMetadata(newEntry, r)
			if err != nil {
				return err
			}
			head = append(head, newEntry)
		}
	}
	sort.Slice(head, func(i, j int) bool { return head[i].DocumentID < head[j].DocumentID })
	d.Files = head
	return nil

}
func (d *BlobDoc) ToDocument() *model.Document {
	var lastModified string
	unixTime, err := strconv.ParseInt(d.Metadata.LastModified, 10, 64)
	if err == nil {
		//HACK: convert wrong nano timestamps to millis
		if len(d.Metadata.LastModified) > 18 {
			unixTime /= 1000000
		}

		t := time.Unix(unixTime/1000, 0)
		lastModified = t.UTC().Format(time.RFC3339Nano)
	}
	return &model.Document{
		ID:             d.DocumentID,
		VissibleName:   d.Metadata.DocName,
		Version:        d.Metadata.Version,
		Parent:         d.Metadata.Parent,
		Type:           d.Metadata.CollectionType,
		CurrentPage:    d.Metadata.LastOpenedPage,
		ModifiedClient: lastModified,
	}
}
