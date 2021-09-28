package sync15

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"sort"

	"github.com/juruen/rmapi/log"
)

func HashEntries(entries []*Entry) (string, error) {
	sort.Slice(entries, func(i, j int) bool { return entries[i].DocumentID < entries[j].DocumentID })
	hasher := sha256.New()
	for _, d := range entries {
		//TODO: back and forth converting
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
