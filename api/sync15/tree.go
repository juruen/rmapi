package sync15

import (
	"bufio"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"

	"github.com/juruen/rmapi/log"
	"golang.org/x/sync/errgroup"
)

const SchemaVersion = "3"
const DocType = "80000000"
const FileType = "0"
const Delimiter = ':'

func FileHashAndSize(file string) ([]byte, int64, error) {
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

func (t *HashTree) IndexReader() (io.ReadCloser, error) {
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

type HashTree struct {
	Hash         string
	Generation   int64
	Docs         []*BlobDoc
	CacheVersion int
}

func (t *HashTree) FindDoc(id string) (*BlobDoc, error) {
	//O(n)
	for _, d := range t.Docs {
		if d.DocumentID == id {
			return d, nil
		}
	}
	return nil, fmt.Errorf("doc %s not found", id)
}

func (t *HashTree) Remove(id string) error {
	docIndex := -1
	for index, d := range t.Docs {
		if d.DocumentID == id {
			docIndex = index
			break
		}
	}
	if docIndex > -1 {
		log.Trace.Printf("Removing %s", id)
		length := len(t.Docs) - 1
		t.Docs[docIndex] = t.Docs[length]
		t.Docs = t.Docs[:length]

		t.Rehash()
		return nil
	}
	return fmt.Errorf("%s not found", id)
}

func (t *HashTree) Rehash() error {
	entries := []*Entry{}
	for _, e := range t.Docs {
		entries = append(entries, &e.Entry)
	}
	hash, err := HashEntries(entries)
	if err != nil {
		return err
	}
	log.Info.Println("New root hash: ", hash)
	t.Hash = hash
	return nil
}

// / Mirror makes the tree look like the storage
func (t *HashTree) Mirror(r RemoteStorage, maxconcurrent int) error {
	rootHash, gen, err := r.GetRootIndex()
	if err != nil {
		return err
	}
	if rootHash == "" && gen == 0 {
		log.Info.Println("Empty cloud")
		t.Docs = nil
		t.Generation = 0
		return nil
	}

	if rootHash == t.Hash {
		return nil
	}
	log.Info.Printf("remote root hash different")

	rootIndexReader, err := r.GetReader(rootHash)
	if err != nil {
		return err
	}
	defer rootIndexReader.Close()

	entries, err := parseIndex(rootIndexReader)
	if err != nil {
		return err
	}

	head := make([]*BlobDoc, 0)
	current := make(map[string]*BlobDoc)
	new := make(map[string]*Entry)

	for _, e := range entries {
		new[e.DocumentID] = e
	}
	wg, ctx := errgroup.WithContext(context.TODO())
	wg.SetLimit(maxconcurrent)

	//current documents
	for _, doc := range t.Docs {
		if entry, ok := new[doc.DocumentID]; ok {
			head = append(head, doc)
			current[doc.DocumentID] = doc

			if entry.Hash != doc.Hash {
				log.Info.Println("doc updated: ", doc.DocumentID)
				e := entry
				d := doc
				wg.Go(func() error {
					return d.Mirror(e, r)
				})
			}
		}
		select {
		case <-ctx.Done():
			goto EXIT
		default:
		}
	}

	//find new entries
	for k, newEntry := range new {
		if _, ok := current[k]; !ok {
			doc := &BlobDoc{}
			log.Trace.Println("doc new: ", k)
			head = append(head, doc)
			e := newEntry
			wg.Go(func() error {
				return doc.Mirror(e, r)
			})
		}
		select {
		case <-ctx.Done():
			goto EXIT
		default:
		}
	}
EXIT:
	err = wg.Wait()
	if err != nil {
		return err
	}
	sort.Slice(head, func(i, j int) bool { return head[i].DocumentID < head[j].DocumentID })
	t.Docs = head
	t.Generation = gen
	t.Hash = rootHash
	return nil
}

func BuildTree(provider RemoteStorage) (*HashTree, error) {
	tree := HashTree{}

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

		doc := &BlobDoc{}
		doc.Entry = *e

		items, _ := parseIndex(f)
		doc.Files = items
		for _, i := range items {
			doc.ReadMetadata(i, provider)
		}
		//don't include deleted items
		if doc.Metadata.Deleted {
			continue
		}

		tree.Docs = append(tree.Docs, doc)
	}

	return &tree, nil

}
