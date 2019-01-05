package util

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/juruen/rmapi/log"
)

type zipDocumentContent struct {
	ExtraMetadata  map[string]string `json:"extraMetadata"`
	FileType       string            `json:"fileType"`
	LastOpenedPage int               `json:"lastOpenedPage"`
	LineHeight     int               `json:"lineHeight"`
	Margins        int               `json:"margins"`
	TextScale      int               `json:"textScale"`
	Transform      map[string]string `json:"transform"`
}

func CreateZipDocument(id, srcPath string) (string, error) {
	pdf, err := ioutil.ReadFile(srcPath)

	if err != nil {
		log.Error.Println("failed to open source document file to read", err)
		return "", err
	}

	tmp, err := ioutil.TempFile("", "rmapizip")
	log.Trace.Println("creating temp zip file:", tmp.Name())
	defer tmp.Close()

	if err != nil {
		log.Error.Println("failed to create tmpfile for zip doc", err)
		return "", err
	}

	w := zip.NewWriter(tmp)
	defer w.Close()

	// Create document (pdf or epub) file
	ext := "pdf"
	if strings.HasSuffix(srcPath, "epub") {
		ext = "epub"
	}

	f, err := w.Create(fmt.Sprintf("%s.%s", id, ext))
	if err != nil {
		log.Error.Println("failed to create doc entry in zip file", err)
		return "", err
	}

	f.Write(pdf)

	// Create pagedata file
	f, err = w.Create(fmt.Sprintf("%s.pagedata", id))
	if err != nil {
		log.Error.Println("failed to create content entry in zip file", err)
		return "", err
	}

	f.Write(make([]byte, 0))

	// Create content content
	f, err = w.Create(fmt.Sprintf("%s.content", id))
	if err != nil {
		log.Error.Println("failed to create content entry in zip file", err)
		return "", err
	}

	c, err := createZipContent(ext)
	if err != nil {
		return "", err
	}

	f.Write([]byte(c))

	return tmp.Name(), nil
}

func CreateZipDirectory(id string) (string, error) {
	tmp, err := ioutil.TempFile("", "rmapizip")
	log.Trace.Println("creating temp zip file:", tmp.Name())
	defer tmp.Close()

	if err != nil {
		log.Error.Println("failed to create tmpfile for zip dir", err)
		return "", err
	}

	w := zip.NewWriter(tmp)
	defer w.Close()

	// Create content content
	f, err := w.Create(fmt.Sprintf("%s.content", id))
	if err != nil {
		log.Error.Println("failed to create content entry in zip file", err)
		return "", err
	}

	f.Write([]byte("{}"))

	return tmp.Name(), nil
}


func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, 0700)
		} else {
			os.MkdirAll(filepath.Dir(path), 0700)
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

func createZipContent(ext string) (string, error) {
	c := zipDocumentContent{
		make(map[string]string),
		ext,
		0,
		-1,
		180,
		1,
		make(map[string]string),
	}

	cstring, err := json.Marshal(c)

	log.Trace.Println("content: ", string(cstring))

	if err != nil {
		log.Error.Println("failed to serialize content file", err)
		return "", err
	}

	return string(cstring), nil
}
