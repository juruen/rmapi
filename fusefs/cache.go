package fusefs

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/juruen/rmapi/api"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
	"golang.org/x/sys/unix"
)

const (
	INDEX_FILE = "rmapi.idx"
)

type FsCache interface {
	FetchDocument(id string) (string, error)
	FetchHttpMetaDocument(id string) (model.HttpDocumentMeta, error)
}

type IndexDocument map[string]model.HttpDocumentMeta

type DiskCache struct {
	BackedDir string
	Rmapi     api.RmApi
	Index     IndexDocument
}

func CreateDiskCache(api api.RmApi, dir string) DiskCache {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Error.Fatal("directory", dir, "doesn't exist", err)
	}

	if err := unix.Access(dir, unix.W_OK); err != nil {
		log.Error.Fatal("directory", dir, "is not writtable", err)
	}

	return DiskCache{BackedDir: dir, Rmapi: api}
}

func (dcache *DiskCache) LoadIndex() {
	indexFile := dcache.indexFilePath()

	if _, err := os.Stat(indexFile); os.IsNotExist(err) {
		log.Info.Print("index file", indexFile, "doesn't exist")
		return
	}

	file, err := os.Open(indexFile)

	if err != nil {
		log.Error.Fatal("failed to open index file", indexFile, err)
	}

	decoder := json.NewDecoder(file)
	var decodedIndex IndexDocument
	err = decoder.Decode(&decodedIndex)

	if err != nil {
		log.Error.Fatal("failed to decode index file, deleting it", indexFile, err)
		os.Remove(indexFile)
	}
}

func (dcache *DiskCache) syncDiskFiles() {

}

func (dcache *DiskCache) indexFilePath() string {
	return fmt.Sprintf("%s/%s", dcache.BackedDir, INDEX_FILE)
}
