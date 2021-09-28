package sync15

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/juruen/rmapi/archive"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
)

type FieldReader struct {
	index  int
	fields []string
}

func (fr *FieldReader) HasNext() bool {
	return fr.index < len(fr.fields)
}

func (fr *FieldReader) Next() (string, error) {
	if fr.index >= len(fr.fields) {
		return "", fmt.Errorf("out of bounds")
	}
	res := fr.fields[fr.index]
	fr.index++
	return res, nil
}

func FileHash(file string) ([]byte, int64, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()

	hasher := sha256.New()
	io.Copy(hasher, f)
	h := hasher.Sum(nil)
	size, err := f.Seek(0, os.SEEK_CUR)
	return h, size, err

}

// func dirHash(dir string) ([]byte, error) {

// 	hashes := make([][]byte, 0)
// 	entries, err := os.ReadDir(dir)
// 	if err != nil {
// 		return nil, err
// 	}
// 	fmt.Println(entries)

// 	for _, f := range entries {
// 		fp := path.Join(dir, f.Name())
// 		fmt.Println(fp)
// 		fh, err := fileHash(fp)
// 		if err != nil {
// 			return nil, err
// 		}
// 		hashes = append(hashes, fh)
// 	}
// 	hasher := sha256.New()
// 	for _, h := range hashes {
// 		hasher.Write(h)
// 	}
// 	dirhash := hasher.Sum(nil)
// 	return dirhash, nil

// }

func NewBlobDoc(name, documentId, colType string) *Doc {
	return &Doc{
		MetadataFile: archive.MetadataFile{
			DocName:        name,
			CollectionType: colType,
		},
		Entry: Entry{
			DocumentID: documentId,
		},
	}

}

func NewFieldReader(line string) FieldReader {
	fld := strings.FieldsFunc(line, func(r rune) bool { return r == Delimiter })

	fr := FieldReader{
		index:  0,
		fields: fld,
	}
	return fr

}
func parseEntry(line string) (*Entry, error) {
	entry := Entry{}
	rdr := NewFieldReader(line)
	numFields := len(rdr.fields)
	if numFields != 5 {
		return nil, fmt.Errorf("wrong number of fields %d", numFields)

	}
	var err error
	entry.Hash, err = rdr.Next()
	if err != nil {
		return nil, err
	}
	entry.Type, err = rdr.Next()
	if err != nil {
		return nil, err
	}
	entry.DocumentID, err = rdr.Next()
	if err != nil {
		return nil, err
	}
	tmp, err := rdr.Next()
	if err != nil {
		return nil, err
	}
	entry.Subfiles, err = strconv.Atoi(tmp)
	if err != nil {
		return nil, fmt.Errorf("cannot read subfiles %s %v", line, err)
	}
	tmp, err = rdr.Next()
	if err != nil {
		return nil, err
	}
	entry.Size, err = strconv.ParseInt(tmp, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("cannot read size %s %v", line, err)
	}
	return &entry, nil
}

const SchemaVersion = "3"
const DocType = "80000000"
const FileType = "0"

func parseIndex(f io.Reader) ([]*Entry, error) {
	var entries []*Entry
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	schema := scanner.Text()

	if schema != SchemaVersion {
		return nil, errors.New("wrong schema")
	}
	for scanner.Scan() {
		line := scanner.Text()
		entry, err := parseEntry(line)
		if err != nil {
			return nil, fmt.Errorf("cant parse line '%s', %w", line, err)
		}

		entries = append(entries, entry)
	}
	return entries, nil
}

func (t *Tree) IndexReader() (io.ReadCloser, error) {
	pipeReader, pipeWriter := io.Pipe()
	w := bufio.NewWriter(pipeWriter)
	go func() {
		defer pipeWriter.Close()
		w.WriteString(SchemaVersion)
		w.WriteString("\n")
		for _, d := range t.Docs {
			w.WriteString(d.Line())
			w.WriteString("\n")
		}
		w.Flush()
	}()

	return pipeReader, nil
}

func (t *Doc) IndexReader() (io.ReadCloser, error) {
	if len(t.Files) == 0 {
		return nil, errors.New("no files")
	}
	pipeReader, pipeWriter := io.Pipe()
	w := bufio.NewWriter(pipeWriter)
	go func() {
		defer pipeWriter.Close()
		w.WriteString(SchemaVersion)
		w.WriteString("\n")
		for _, d := range t.Files {
			w.WriteString(d.Line())
			w.WriteString("\n")
		}
		w.Flush()
	}()

	return pipeReader, nil
}

type Tree struct {
	Hash       string
	Generation int64
	Docs       []*Doc
}

type Doc struct {
	Files []*Entry
	Entry
	archive.MetadataFile
}

const Delimiter = ':'

func (d *Entry) Line() string {
	var sb strings.Builder
	sb.WriteString(d.Hash)
	sb.WriteRune(Delimiter)
	sb.WriteString(FileType)
	sb.WriteRune(Delimiter)
	sb.WriteString(d.DocumentID)
	sb.WriteRune(Delimiter)
	sb.WriteString("0")
	sb.WriteRune(Delimiter)
	sb.WriteString(strconv.FormatInt(d.Size, 10))
	return sb.String()
}
func (d *Doc) Line() string {
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

type Entry struct {
	Hash       string
	Type       string
	DocumentID string
	Subfiles   int
	Size       int64
}

func getCachedTreePath() (string, error) {
	cachedir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	x := path.Join(cachedir, "rmapi")
	os.MkdirAll(x, 0700)
	cacheFile := path.Join(x, ".tree")
	return cacheFile, nil
}

func loadTree() (*Tree, error) {
	tree := Tree{}
	cacheFile, err := getCachedTreePath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(cacheFile); err == nil {
		b, err := ioutil.ReadFile(cacheFile)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(b, &tree)
		if err != nil {
			log.Error.Println("cache corrupt")
			return nil, err
		}
	} else {
		os.Create(cacheFile)
	}
	log.Info.Println("cache loaded")

	return &tree, nil
}
func saveTree(tree *Tree) error {
	cacheFile, err := getCachedTreePath()
	log.Info.Println("Writing cache: ", cacheFile)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(tree, "", "")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(cacheFile, b, 0644)
	return err
}

// Extract the documentname from metadata blob
func (doc *Doc) SyncName(fileEntry *Entry, r RemoteStorage) error {
	if strings.HasSuffix(fileEntry.DocumentID, ".metadata") {
		log.Info.Println("Reading metadata: " + doc.DocumentID)

		metadata := archive.MetadataFile{}

		meta, err := r.GetReader(fileEntry.Hash)
		if err != nil {
			return err
		}
		defer meta.Close()
		content, err := ioutil.ReadAll(meta)
		if err != nil {
			return err
		}
		err = json.Unmarshal(content, &metadata)
		if err != nil {
			log.Info.Printf("cannot read metadata %v", err)
		}
		log.Info.Println(metadata.DocName)
		doc.MetadataFile = metadata
	}

	return nil
}

func (doc *Doc) ToDocument() *model.Document {
	return &model.Document{
		ID:           doc.DocumentID,
		VissibleName: doc.MetadataFile.DocName,
		Version:      doc.MetadataFile.Version,
		Parent:       doc.MetadataFile.Parent,
		Type:         doc.MetadataFile.CollectionType,
	}

}

func (t *Tree) FindDoc(id string) (*Doc, error) {
	//O(n)
	for _, d := range t.Docs {
		if d.DocumentID == id {
			return d, nil
		}
	}
	return nil, errors.New("not found")
}

func (t *Tree) Remove(id string) error {
	docIndex := -1
	for index, d := range t.Docs {
		if d.DocumentID == id {
			docIndex = index
			break
		}
	}
	if docIndex > -1 {
		log.Info.Printf("Removing %s", id)
		length := len(t.Docs) - 1
		t.Docs[docIndex] = t.Docs[length]
		t.Docs = t.Docs[:length]

		t.Rehash()
		return nil
	}
	return errors.New("not found")
}

func (doc *Doc) Sync(e *Entry, r RemoteStorage) error {
	doc.Entry = *e
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

	for _, currentEntry := range doc.Files {
		if newEntry, ok := new[currentEntry.DocumentID]; ok {
			if newEntry.Hash != currentEntry.Hash {
				err = doc.SyncName(newEntry, r)
				if err != nil {
					return err
				}
				currentEntry.Hash = newEntry.Hash
			}
			head = append(head, currentEntry)
			current[currentEntry.DocumentID] = currentEntry
		}
	}

	for k, newEntry := range new {
		if _, ok := current[k]; !ok {
			err = doc.SyncName(newEntry, r)
			if err != nil {
				return err
			}
			head = append(head, newEntry)
		}
	}
	sort.Slice(head, func(i, j int) bool { return head[i].DocumentID < head[j].DocumentID })
	doc.Files = head
	return nil

}

func Hash(entries []*Entry) (string, error) {
	sort.Slice(entries, func(i, j int) bool { return entries[i].DocumentID < entries[j].DocumentID })
	hasher := sha256.New()
	for _, d := range entries {

		bh, err := hex.DecodeString(d.Hash)
		if err != nil {
			return "", err
		}
		hasher.Write(bh)
	}
	hash := hasher.Sum(nil)
	hashStr := hex.EncodeToString(hash)
	return hashStr, nil
}

func (t *Tree) Rehash() error {
	entries := []*Entry{}
	for _, e := range t.Docs {
		entries = append(entries, &e.Entry)
	}
	hash, err := Hash(entries)
	if err != nil {
		return err
	}
	t.Hash = hash
	return nil
}

func (d *Doc) AddFile(e *Entry) error {
	d.Files = append(d.Files, e)
	hash, err := Hash(d.Files)
	if err != nil {
		return err
	}
	d.Hash = hash
	return nil
}

func (t *Tree) Add(d *Doc) error {
	if len(d.Files) == 0 {
		return errors.New("no files")
	}
	t.Docs = append(t.Docs, d)
	return t.Rehash()
}

/// Sync makes the tree look like the storage
func (t *Tree) Sync(r RemoteStorage) error {
	rootHash, gen, err := r.GetRootIndex()
	if err != nil {
		return err
	}
	if rootHash == "" && gen == 0 {
		log.Info.Println("empty cloud")
		t.Docs = nil
		t.Generation = 0
		return nil
	}

	if rootHash == t.Hash {
		return nil
	}
	log.Info.Printf("Root hash different")

	rdr, err := r.GetReader(rootHash)
	if err != nil {
		return err
	}
	defer rdr.Close()

	entries, err := parseIndex(rdr)
	if err != nil {
		return err
	}

	head := make([]*Doc, 0)
	current := make(map[string]*Doc)
	new := make(map[string]*Entry)
	for _, e := range entries {
		new[e.DocumentID] = e
	}
	//current documents
	for _, doc := range t.Docs {
		if entry, ok := new[doc.Entry.DocumentID]; ok {
			//hash different update
			if entry.Hash != doc.Hash {
				log.Info.Println("doc updated: " + doc.DocumentID)
				doc.Sync(entry, r)
			}
			head = append(head, doc)
			current[doc.DocumentID] = doc
		}

	}

	//find new entries
	for k, newEntry := range new {
		if _, ok := current[k]; !ok {
			doc := &Doc{}
			log.Info.Println("doc new: " + k)
			doc.Sync(newEntry, r)
			head = append(head, doc)
		}
	}
	sort.Slice(head, func(i, j int) bool { return head[i].DocumentID < head[j].DocumentID })
	t.Docs = head
	t.Generation = gen
	t.Hash = rootHash
	return nil
}

func readTree(provider RemoteStorage) (*Tree, error) {
	tree := Tree{}

	rootHash, gen, err := provider.GetRootIndex()

	if err != nil {
		return nil, err
	}
	tree.Hash = rootHash
	tree.Generation = gen

	rootIndex, err := provider.GetReader(rootHash)
	if err != nil {
		return nil, err
	}

	defer rootIndex.Close()
	entries, _ := parseIndex(rootIndex)

	for _, e := range entries {
		f, _ := provider.GetReader(e.Hash)
		defer f.Close()

		doc := &Doc{}
		doc.Entry = *e
		tree.Docs = append(tree.Docs, doc)

		items, _ := parseIndex(f)
		doc.Files = items
		for _, i := range items {
			doc.SyncName(i, provider)
		}
	}

	return &tree, nil

}
