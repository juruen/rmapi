package archive

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/juruen/rmapi/encoding/rm"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/util"
)

// Read fills a Zip parsing a Remarkable archive file.
func (z *Zip) Read(r io.ReaderAt, size int64) error {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return err
	}

	// reading content first because it contains the number of pages
	if err := z.readContent(zr); err != nil {
		return err
	}

	if err := z.readPayload(zr); err != nil {
		return err
	}

	//uploading and then downloading a file results in 0 pages
	if z.Content.PageCount <= 0 {
		log.Warning.Printf("PageCount is 0")
		return nil
	}

	if err := z.readMetadata(zr); err != nil {
		return err
	}

	if err := z.readPagedata(zr); err != nil {
		return err
	}

	if err := z.readData(zr); err != nil {
		return err
	}

	if err := z.readThumbnails(zr); err != nil {
		return err
	}

	return nil
}

// readContent reads the .content file contained in an archive and the UUID
func (z *Zip) readContent(zr *zip.Reader) error {
	files, err := zipExtFinder(zr, ".content")
	if err != nil {
		return err
	}

	if len(files) != 1 {
		return errors.New("archive does not contain a unique content file")
	}

	contentFile := files[0]
	file, err := contentFile.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(bytes, &z.Content); err != nil {
		return err
	}
	p := contentFile.FileInfo().Name()
	id, _ := util.DocPathToName(p)
	z.UUID = id

	if z.Content.Pages != nil && len(z.Content.Pages) > 0 {
		z.pageMap = make(map[string]int)
		log.Info.Println("Pages is not empty, using that")
		z.Pages = make([]Page, len(z.Content.Pages))
		for i, p := range z.Content.Pages {
			z.pageMap[p] = i
		}
	} else {
		// instantiate the slice of pages
		z.Pages = make([]Page, z.Content.PageCount)
	}
	return nil
}

// readPagedata reads the .pagedata file contained in an archive
// and iterate to gather which template was used for each page.
func (z *Zip) readPagedata(zr *zip.Reader) error {
	files, err := zipExtFinder(zr, ".pagedata")
	if err != nil {
		return err
	}

	if len(files) != 1 {
		return errors.New("archive does not contain a unique pagedata file")
	}

	file, err := files[0].Open()
	if err != nil {
		return err
	}
	defer file.Close()

	// iterate pagedata file lines
	sc := bufio.NewScanner(file)
	var i int = 0
	for sc.Scan() {
		line := sc.Text()
		z.Pages[i].Pagedata = line
		i++
	}

	if err := sc.Err(); err != nil {
		return err
	}

	return nil
}

// readPayload tries to extract the payload from an archive if it exists.
func (z *Zip) readPayload(zr *zip.Reader) error {
	ext := z.Content.FileType
	files, err := zipExtFinder(zr, "."+ext)
	if err != nil {
		return err
	}

	// return if not found
	if len(files) != 1 {
		return nil
	}

	file, err := files[0].Open()
	if err != nil {
		return err
	}
	defer file.Close()

	z.Payload, err = ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	return nil
}

// readData extracts existing .rm files from an archive.
func (z *Zip) readData(zr *zip.Reader) error {
	files, err := zipExtFinder(zr, ".rm")
	if err != nil {
		return err
	}

	for _, file := range files {
		name, _ := splitExt(file.FileInfo().Name())

		idx, err := z.pageIndex(name)
		if err != nil {
			return err
		}

		if len(z.Pages) <= idx {
			return errors.New("page not found")
		}

		r, err := file.Open()
		if err != nil {
			return err
		}

		bytes, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}

		z.Pages[idx].Data = rm.New()
		z.Pages[idx].Data.UnmarshalBinary(bytes)
		if err != nil {
			return err
		}
	}

	return nil
}

// readThumbnails extracts existing thumbnails from an archive.
func (z *Zip) readThumbnails(zr *zip.Reader) error {
	files, err := zipExtFinder(zr, ".jpg")
	if err != nil {
		return err
	}

	for _, file := range files {
		name, _ := splitExt(file.FileInfo().Name())

		idx, err := strconv.Atoi(name)
		if err != nil {
			return errors.New("error in .jpg filename")
		}

		if len(z.Pages) <= idx {
			return errors.New("page not found")
		}

		r, err := file.Open()
		if err != nil {
			return err
		}

		z.Pages[idx].Thumbnail, err = ioutil.ReadAll(r)
		if err != nil {
			return err
		}
	}

	return nil
}

func (z *Zip) pageIndex(namePart string) (idx int, err error) {
	if z.pageMap != nil {
		_, err = uuid.Parse(namePart)
		if err != nil {
			return -1, errors.New("pages not in uuid format " + namePart)
		}
		var ok bool
		idx, ok = z.pageMap[namePart]
		if !ok {
			log.Warning.Println("Page not found in map: ", namePart)
		}

	} else {
		idx, err = strconv.Atoi(namePart)
		if err != nil {
			return -1, errors.New("error in metadata .json filename")
		}
	}
	return
}

// readMetadata extracts existing .json metadata files from an archive.
func (z *Zip) readMetadata(zr *zip.Reader) error {
	files, err := zipExtFinder(zr, ".json")
	if err != nil {
		return err
	}

	for _, file := range files {
		name, _ := splitExt(file.FileInfo().Name())

		// name is 0-metadata.json or uuid-metadata
		namePart := strings.TrimSuffix(name, "-metadata")
		idx, err := z.pageIndex(namePart)
		if err != nil {
			return err
		}

		if len(z.Pages) <= idx {
			return errors.New("page not found")
		}

		r, err := file.Open()
		if err != nil {
			return err
		}

		bytes, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}

		err = json.Unmarshal(bytes, &z.Pages[idx].Metadata)
		if err != nil {
			return err
		}
	}

	return nil
}

// splitExt splits the extension from a filename
func splitExt(name string) (string, string) {
	ext := filepath.Ext(name)
	return name[0 : len(name)-len(ext)], ext
}

// zipExtFinder searches for a file matching the substr pattern
// in a zip file.
func zipExtFinder(zr *zip.Reader, ext string) ([]*zip.File, error) {
	var files []*zip.File

	for _, file := range zr.File {
		if _, e := splitExt(file.FileInfo().Name()); e == ext {
			files = append(files, file)
		}
	}

	return files, nil
}
