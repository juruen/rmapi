package archive

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/juruen/rmapi/log"
	"github.com/nfnt/resize"
	pdfmodel "github.com/unidoc/unipdf/v3/model"
	"github.com/unidoc/unipdf/v3/render"
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

func makeThumbnail(pdf []byte) ([]byte, error) {

	reader, err := pdfmodel.NewPdfReader(bytes.NewReader(pdf))
	if err != nil {
		return nil, err
	}
	page, err := reader.GetPage(1)
	if err != nil {
		return nil, err
	}

	device := render.NewImageDevice()

	image, err := device.Render(page)
	if err != nil {
		return nil, err
	}

	thumbnail := resize.Resize(280, 374, image, resize.Lanczos3)
	out := &bytes.Buffer{}
	jpeg.Encode(out, thumbnail, nil)

	return out.Bytes(), nil
}

// gets the Document UUID from an archive
func GetIdFromZip(srcPath string) (id string, err error) {
	file, err := os.Open(srcPath)
	if err != nil {
		return
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		return
	}
	zip := Zip{}
	err = zip.Read(file, fi.Size())
	if err != nil {
		return
	}
	id = zip.UUID
	return
}

func CreateZipDocument(id, srcPath string) (zipPath string, err error) {
	ext := strings.TrimPrefix(filepath.Ext(srcPath), ".")
	if ext == "zip" {
		zipPath = srcPath
		return
	}

	doc, err := ioutil.ReadFile(srcPath)
	if err != nil {
		log.Error.Println("failed to open source document file to read", err)
		return
	}
	// Create document (pdf or epub) file
	tmp, err := ioutil.TempFile("", "rmapizip")
	if err != nil {
		return
	}
	defer tmp.Close()

	if err != nil {
		log.Error.Println("failed to create tmpfile for zip doc", err)
		return
	}

	w := zip.NewWriter(tmp)
	defer w.Close()

	f, err := w.Create(fmt.Sprintf("%s.%s", id, ext))
	if err != nil {
		log.Error.Println("failed to create doc entry in zip file", err)
		return
	}
	f.Write(doc)

	//try to create a thumbnail
	if ext == "pdf" {
		thumbnail, err := makeThumbnail(doc)
		if err != nil {
			log.Error.Println("cannot generate thumbnail", err)
		} else {
			f, err := w.Create(fmt.Sprintf("%s.thumbnails/0.jpg", id))
			if err != nil {
				log.Error.Println("failed to create doc entry in zip file", err)
				return "", err
			}
			f.Write(thumbnail)
		}
	}

	// Create pagedata file
	f, err = w.Create(fmt.Sprintf("%s.pagedata", id))
	if err != nil {
		log.Error.Println("failed to create content entry in zip file", err)
		return
	}
	f.Write(make([]byte, 0))

	//create thumbnail

	// Create content content
	f, err = w.Create(fmt.Sprintf("%s.content", id))
	if err != nil {
		log.Error.Println("failed to create content entry in zip file", err)
		return
	}

	c, err := createZipContent(ext)
	if err != nil {
		return
	}

	f.Write([]byte(c))
	zipPath = tmp.Name()

	return
}

func CreateZipDirectory(id string) (string, error) {
	tmp, err := ioutil.TempFile("", "rmapizip")
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

	if err != nil {
		log.Error.Println("failed to serialize content file", err)
		return "", err
	}

	return string(cstring), nil
}
