package util

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path"
	"strings"
)

const (
	PDF  = "pdf"
	ZIP  = "zip"
	RM   = "rm"
	EPUB = "epub"
)

var supportedExt = map[string]bool{
	EPUB: true,
	PDF:  true,
	ZIP:  true,
	RM:   true,
}

func IsFileTypeSupported(ext string) bool {
	return supportedExt[ext]
}

// DocPathToName extracts the file name and file extension (without .) from a given path
func DocPathToName(p string) (name string, ext string) {
	tmpExt := path.Ext(p)
	name = strings.TrimSuffix(path.Base(p), tmpExt)
	ext = strings.ToLower(strings.TrimPrefix(tmpExt, "."))
	return
}

func ToIOReader(source interface{}) (io.Reader, error) {
	var content []byte
	var err error

	if source == nil {
		return bytes.NewReader(nil), nil
	}

	content, err = json.Marshal(source)

	return bytes.NewReader(content), err
}

func CopyFile(src, dst string) (int64, error) {
	r, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	w, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer w.Close()

	n, err := io.Copy(w, r)
	if err != nil {
		return 0, err
	}

	return n, nil
}

// Wraps a request in a slice (serialize as json array)
func InSlice(req interface{}) []interface{} {
	slice := make([]interface{}, 0)
	slice = append(slice, req)
	return slice
}
