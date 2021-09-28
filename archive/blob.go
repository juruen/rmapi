package archive

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/model"
)

type FileMap struct {
	Name string
	Path string
}

func AddStuff(f *[]*FileMap, name, filepath string) {
	fs := &FileMap{
		Name: name,
		Path: filepath,
	}
	*f = append(*f, fs)
}

func Prepare(name, parentId, sourceDocPath, ext, tmpDir string) (files []*FileMap, id string, err error) {
	files = []*FileMap{}
	if ext == "zip" {
		var metadataPath string
		id, files, metadataPath, err = UnpackAndFix(sourceDocPath, tmpDir)
		if err != nil {
			return
		}
		if id == "" {
			return nil, "", errors.New("could not determine the Document UUID")
		}
		if metadataPath == "" {
			log.Warning.Println("missing metadata, creating...", name)
			objectName, filePath, err1 := CreateMetadata(id, name, parentId, model.DocumentType, tmpDir)
			if err1 != nil {
				err = err1
				return
			}
			AddStuff(&files, objectName, filePath)
		} else {
			err = FixMetadata(parentId, name, metadataPath)
			if err != nil {
				return
			}
		}
	} else {
		id = uuid.New().String()
		objectName := id + "." + ext
		doctype := ext
		if ext == "rm" {
			pageId := uuid.New().String()
			objectName = fmt.Sprintf("%s/%s.rm", id, pageId)
			doctype = "notebook"
		}
		AddStuff(&files, objectName, sourceDocPath)
		objectName, filePath, err1 := CreateMetadata(id, name, parentId, model.DocumentType, tmpDir)
		if err1 != nil {
			err = err1
			return
		}
		AddStuff(&files, objectName, filePath)

		objectName, filePath, err = CreateContent(id, doctype, tmpDir)
		if err != nil {
			return
		}
		AddStuff(&files, objectName, filePath)
	}
	return files, id, err
}

func FixMetadata(parentId, name, path string) error {
	meta := MetadataFile{}
	metaData, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(metaData, &meta)
	if err != nil {
		return err
	}
	meta.Parent = parentId
	meta.DocName = name
	metaData, err = json.Marshal(meta)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, metaData, 0600)
}
func UnpackAndFix(src, dest string) (id string, files []*FileMap, metadataPath string, err error) {
	log.Info.Println("Unpacking in: ", dest)
	r, err := zip.OpenReader(src)
	if err != nil {
		return
	}
	defer r.Close()
	files = []*FileMap{}

	for _, f := range r.File {
		fname := f.Name

		if strings.HasSuffix(fname, ".content") {
			id = strings.TrimSuffix(fname, path.Ext(fname))
		}
		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			err = fmt.Errorf("%s: illegal file path", fpath)
			return
		}

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		} else {
			AddStuff(&files, f.Name, fpath)
		}

		if strings.HasSuffix(fname, ".metadata") {
			metadataPath = fpath
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return
		}

		outFile, err1 := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err1 != nil {
			err = err1
			return
		}

		rc, err1 := f.Open()
		if err != nil {
			err = err1
			return
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return
		}
	}

	return id, files, metadataPath, nil
}
