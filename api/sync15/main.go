package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

type TextReader struct {
	index int
	slc   []string
}

func (txt *TextReader) HasNext() bool {
	return txt.index < len(txt.slc)
}

func (txt *TextReader) Next() (string, error) {
	if txt.index >= len(txt.slc) {
		return "", fmt.Errorf("out of bounds")
	}
	res := txt.slc[txt.index]
	txt.index++
	return res, nil

}

func woker(c chan int) {
	<-c
	c <- 0
	// stuf := fmt.Sprintf("%d", 10)

}

type Foo struct {
	bar *int
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
func parse(f io.Reader) {
	scanner := bufio.NewScanner(f)
	scanner.Scan()
	schema := scanner.Text()

	if schema != "3" {
		log.Fatal("wrong schema")
	}
	fmt.Println(schema)
	for scanner.Scan() {
		line := scanner.Text()
		fld := strings.FieldsFunc(line, func(r rune) bool { return r == ':' })
		rdr := TextReader{
			index: 0,
			slc:   fld,
		}
		hash1, _ := rdr.Next()
		bt, err := hex.DecodeString(hash1)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(bt)
		for rdr.HasNext() {
			fmt.Println(rdr.Next())
		}
	}
}

type Tree struct {
	RootHash RootHash
	Entries  []Entry
}

func (t *Tree) IsSame(hash string) bool {
	return hash == t.RootHash.Hash
}

func (t *Tree) MissingOrChanged(other []Entry) []Entry {
	m := map[string]Entry{}
	for _, e := range t.Entries {
		m[e.Uid] = e
	}
	return nil
}

type Entry struct {
	Uid      string
	Hash     string
	Name     string
	Subfiles int
	Leafs    []Leaf
}

type Leaf struct {
	Hash     string
	FileName string
	Size     int
}
type RootHash struct {
	Hash       string
	Generation int
}

func loadTree() error {
	cachedir, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	tree := Tree{}
	fmt.Println(cachedir)
	x := path.Join(cachedir, "rmapi")
	os.MkdirAll(x, 0700)
	cacheFile := path.Join(x, ".tree")
	if _, err := os.Stat(cacheFile); err == nil {
		b, err := ioutil.ReadFile(cacheFile)
		if err != nil {
			return err
		}
		json.Unmarshal(b, &tree)
	} else {
		os.Create(cacheFile)
	}
	fmt.Print(tree)

	return nil
}

func main() {
	fmt.Println("after 1 second")

	f, err := os.Open(path.Join("dir", "file.txt"))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	parse(f)

	drh, err := dirHash("dir")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(hex.EncodeToString(drh))

	loadTree()
}
