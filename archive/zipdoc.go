package archive

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"time"

	uuid "github.com/google/uuid"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/util"
	"github.com/nfnt/resize"
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

	if ext == util.ZIP {
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
	if ext == util.RM {
		pageUUID := uuid.New()
		pageID := pageUUID.String()
		documentPath = fmt.Sprintf("%s/%s.rm", id, pageID)
		fileType = "notebook"
		pages = append(pages, pageID)
	} else {
		documentPath = fmt.Sprintf("%s.%s", id, ext)
		pages = append(pages, "")
	}

	f, err := w.Create(documentPath)
	if err != nil {
		log.Error.Println("failed to create doc entry in zip file", err)
		return
	}
	f.Write(doc)

	//try to create a thumbnail
	//due to a bug somewhere in unipdf the generation is opt-in
	if ext == util.PDF && os.Getenv("RMAPI_THUMBNAILS") != "" {
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

	if err != nil {
		log.Error.Println("failed to create tmpfile for zip dir", err)
		return "", err
	}
	defer tmp.Close()

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

func CreateContent(id, ext, fpath string, pageIds []string) (fileName, filePath string, err error) {
	fileName = id + ".content"
	filePath = path.Join(fpath, fileName)
	content := "{}"

	if ext != "" {
		content, err = createZipContent(ext, pageIds)
		if err != nil {
			return
		}
	}

	err = ioutil.WriteFile(filePath, []byte(content), 0600)
	return
}

func UnixTimestamp() string {
	t := time.Now().UnixMilli()
	tf := strconv.FormatInt(t, 10)
	return tf
}

func CreateMetadata(id, name, parent, colType, fpath string) (fileName string, filePath string, err error) {
	fileName = id + ".metadata"
	filePath = path.Join(fpath, fileName)
	meta := MetadataFile{
		DocName:        name,
		Version:        0,
		CollectionType: colType,
		Parent:         parent,
		Synced:         true,
		LastModified:   UnixTimestamp(),
	}

	c, err := json.Marshal(meta)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(filePath, c, 0600)
	return
}
