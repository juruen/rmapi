package sync15

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
)

type LocalStore struct {
	folder string
}

func (p *LocalStore) GetRootIndex() (string, int64, error) {
	rootPath := path.Join(p.folder, "root")
	root_hash, err := ioutil.ReadFile(rootPath)
	if err != nil {
		return "", 0, err
	}
	strRootHash := string(root_hash)
	rootGenPath := path.Join(p.folder, ".root.history")
	var gen int64

	if fi, err := os.Stat(rootGenPath); err == nil {
		gen = fi.Size() / 86
	}
	fmt.Println("root ->", strRootHash)
	return strRootHash, gen, nil
}

func (p *LocalStore) GetReader(hash string) (io.ReadCloser, error) {
	rootIndexPath := path.Join(p.folder, hash)
	return os.Open(rootIndexPath)
}
