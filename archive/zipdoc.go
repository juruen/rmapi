package archive

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"io"
	"io/ioutil"
	"os"

	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/util"
	"github.com/nfnt/resize"
	uuid "github.com/satori/go.uuid"
	pdfmodel "github.com/unidoc/unipdf/v3/model"
	"github.com/unidoc/unipdf/v3/render"
)

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

// GetIdFromZip tries to get the Document UUID from an archive
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
	_, ext := util.DocPathToName(srcPath)
	fileType := ext

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

	var documentPath string

	pages := make([]string, 0)
	if ext == "rm" {
		var pageUUID uuid.UUID

		pageUUID, err = uuid.NewV4()
		if err != nil {
			return
		}
		pageID := pageUUID.String()
		documentPath = fmt.Sprintf("%s/%s.rm", id, pageID)
		fileType = "notebook"
		pages = append(pages, pageID)
	} else {
		documentPath = fmt.Sprintf("%s.%s", id, ext)
	}

	f, err := w.Create(documentPath)
	if err != nil {
		log.Error.Println("failed to create doc entry in zip file", err)
		return
	}
	f.Write(doc)

	//try to create a thumbnail
	//due to a bug somewhere in unipdf the generation is opt-in
	if ext == "pdf" && os.Getenv("RMAPI_THUMBNAILS") != "" {
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

	// Create content content
	f, err = w.Create(fmt.Sprintf("%s.content", id))
	if err != nil {
		log.Error.Println("failed to create content entry in zip file", err)
		return
	}

	c, err := createZipContent(fileType, pages)
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

func createZipContent(ext string, pageIDs []string) (string, error) {
	c := Content{
		DummyDocument: false,
		ExtraMetadata: ExtraMetadata{
			LastPen:             "Finelinerv2",
			LastTool:            "Finelinerv2",
			LastFinelinerv2Size: "1",
		},
		FileType:       ext,
		PageCount:      0,
		LastOpenedPage: 0,
		LineHeight:     -1,
		Margins:        180,
		TextScale:      1,
		Transform: Transform{
			M11: 1,
			M12: 0,
			M13: 0,
			M21: 0,
			M22: 1,
			M23: 0,
			M31: 0,
			M32: 0,
			M33: 1,
		},
		Pages: pageIDs,
	}

	cstring, err := json.Marshal(c)

	if err != nil {
		log.Error.Println("failed to serialize content file", err)
		return "", err
	}

	return string(cstring), nil
}

// ReplacePayloadInZip takes the given Payload and replaces the Payload at the
// ZIP file located at zipFileName.
func ReplacePayloadInZip(id, zipFileName, fileExt string, payload []byte) (string, error) {
	currentZip, err := zip.OpenReader(zipFileName)
	if err != nil {
		return "", err
	}
	defer currentZip.Close()

	targetPayloadName := fmt.Sprintf("%s%s", id, fileExt)

	var foundFile bool
	for _, f := range currentZip.File {
		if f.Name == targetPayloadName {
			foundFile = true
			break
		}
	}

	if !foundFile {
		return "", fmt.Errorf("expected file of type %s", fileExt)
	}

	tmp, err := ioutil.TempFile("", "rmapizip")
	defer tmp.Close()

	newZip := zip.NewWriter(tmp)
	defer newZip.Close()

	for _, file := range currentZip.File {
		newF, err := newZip.Create(file.Name)
		if err != nil {
			defer os.Remove(tmp.Name())
			return "", err
		}

		var content []byte
		if file.Name == targetPayloadName {
			content = payload
		} else {
			f, err := file.Open()
			if err != nil {
				defer os.Remove(tmp.Name())
				return "", err
			}

			content, err = io.ReadAll(f)
			f.Close()
			if err != nil {
				defer os.Remove(tmp.Name())
				return "", err
			}
		}

		newF.Write(content)
	}

	return tmp.Name(), nil
}
