package util

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/juruen/rmapi/log"
)

type zipDocumentContent struct {
	ExtraMetadata  map[string]string `json:"extraMetadata"`
	FileType       string            `json:"fileType"`
	PageCount      int               `json:"pageCount"`
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

func createZipContent(ext string) (string, error) {
	c := zipDocumentContent{
		make(map[string]string),
		ext,
		0,
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
