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

	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
)

type FieldReader struct {
	index int
	slc   []string
}

func (txt *FieldReader) HasNext() bool {
	return txt.index < len(txt.slc)
}

func (txt *FieldReader) Next() (string, error) {
	if txt.index >= len(txt.slc) {
		return "", fmt.Errorf("out of bounds")
	}
	res := txt.slc[txt.index]
	txt.index++
	return res, nil

}

func fileHash(file string) ([]byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	hasher := sha256.New()
	io.Copy(hasher, f)
	h := hasher.Sum(nil)
	return h, nil

}
func dirHash(dir string) ([]byte, error) {

	hashes := make([][]byte, 0)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	fmt.Println(entries)

	for _, f := range entries {
		fp := path.Join(dir, f.Name())
		fmt.Println(fp)
		fh, err := fileHash(fp)
		if err != nil {
			return nil, err
		}
		hashes = append(hashes, fh)
	}
	hasher := sha256.New()
	for _, h := range hashes {
		hasher.Write(h)
	}
	dirhash := hasher.Sum(nil)
	return dirhash, nil

}

func parseEntry(line string) (*Entry, error) {
	entry := Entry{}
	fld := strings.FieldsFunc(line, func(r rune) bool { return r == ':' })
	rdr := FieldReader{
		index: 0,
		slc:   fld,
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
	sub, err := rdr.Next()
	if err != nil {
		return nil, err
	}
	entry.Subfiles, err = strconv.Atoi(sub)
	if err != nil {
		return nil, err
	}
	sub, err = rdr.Next()
	if err != nil {
		return nil, err
	}
	entry.Size, err = strconv.Atoi(sub)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func parseIndex(f io.Reader) ([]*Entry, error) {
	var entries []*Entry
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	schema := scanner.Text()

	if schema != "3" {
		return nil, errors.New("wrong schema")
	}
	for scanner.Scan() {
		line := scanner.Text()
		entry, err := parseEntry(line)
		if err != nil {
			return nil, err
		}

		entries = append(entries, entry)
	}
	return entries, nil
}

type Tree struct {
	Hash       string
	Generation int64
	Docs       []*Doc
}

type Doc struct {
	Files []*Entry
	Entry
	Metadata
}

type Metadata struct {
	DocName        string `json:"visibleName"`
	CollectionType string `json:"type"`
	Parent         string `json:"parent"`
	LastModified   string `json:"lastModified"`
	LastOpened     string `json:"lastOpened"`
	Version        int    `json:"version"`
}

type Entry struct {
	Hash       string
	Type       string
	DocumentID string
	Subfiles   int
	Size       int
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

		metadata := Metadata{}

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
		doc.Metadata = metadata
	}

	return nil
}

func (doc *Doc) ToDocument() *model.Document {
	return &model.Document{
		ID:           doc.DocumentID,
		VissibleName: doc.Metadata.DocName,
		Version:      doc.Metadata.Version,
		Parent:       doc.Metadata.Parent,
		Type:         doc.Metadata.CollectionType,
	}

}

func (doc *Tree) FindDoc(id string) (*Doc, error) {
	for _, d := range doc.Docs {
		if d.DocumentID == id {
			return d, nil
		}
	}
	return nil, errors.New("not found")
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
	t.Docs = append(t.Docs, d)
	return t.Rehash()
}

/// Sync makes the tree look like the storage
func (t *Tree) Sync(r RemoteStorage) error {
	rootHash, gen, err := r.GetRootIndex()
	if err != nil {
		return err
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

func main() {
	storagePath := "../../../rmfakecloud/data/users/test/sync/"
	provider := &LocalStore{storagePath}
	tree := &Tree{}
	err := tree.Sync(provider) // readTree(provider)
	// tree, err := readTree(provider)
	if err != nil {
		log.Info.Fatal(err)
	}
	err = saveTree(tree)
	if err != nil {
		log.Info.Fatal(err)
	}
}
