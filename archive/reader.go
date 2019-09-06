package archive

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/juruen/rmapi/encoding/rm"
)

// Read fills a Zip parsing a Remarkable archive file.
func (z *Zip) Read(r io.Reader) error {
	zr, err := zipReaderFromIOReader(r)
	if err != nil {
		return err
	}

	// reading content first because it contains the number of pages
	if err := z.readContent(zr); err != nil {
		return err
	}

	// instanciate the slice of pages
	if z.Content.PageCount == 0 {
		return errors.New("document does not have any pages")
	}
	z.Pages = make([]Page, z.Content.PageCount)

	if err := z.readPagedata(zr); err != nil {
		return err
	}

	if err := z.readPdf(zr); err != nil {
		return err
	}

	if err := z.readEpub(zr); err != nil {
		return err
	}

	if err := z.readData(zr); err != nil {
		return err
	}

	if err := z.readThumbnails(zr); err != nil {
		return err
	}

	if err := z.readMetadata(zr); err != nil {
		return err
	}

	return nil
}

// readContent reads the .content file contained in an archive
func (z *Zip) readContent(zr *zip.Reader) error {
	files, err := zipExtFinder(zr, ".content")
	if err != nil {
		return err
	}

	if len(files) != 1 {
		return errors.New("archive does not contain a unique content file")
	}

	file, err := files[0].Open()
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

// readPdf tries to extract a pdf from an archive if it exists.
func (z *Zip) readPdf(zr *zip.Reader) error {
	files, err := zipExtFinder(zr, ".pdf")
	if err != nil {
		return err
	}

	// return if no pdf
	if len(files) != 1 {
		return nil
	}

	file, err := files[0].Open()
	if err != nil {
		return err
	}
	defer file.Close()

	z.Pdf, err = ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	return nil
}

// readEpub tries to extract an epub from an archive if it exists.
func (z *Zip) readEpub(zr *zip.Reader) error {
	files, err := zipExtFinder(zr, ".epub")
	if err != nil {
		return err
	}

	// return if no epub
	if len(files) != 1 {
		return nil
	}

	file, err := files[0].Open()
	if err != nil {
		return err
	}
	defer file.Close()

	z.Epub, err = ioutil.ReadAll(file)
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

		idx, err := strconv.Atoi(name)
		if err != nil {
			return errors.New("error in .rm filename")
		}

		if len(z.Pages) < idx {
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

		if len(z.Pages) < idx {
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

// readMetadata extracts existing .json metadata files from an archive.
func (z *Zip) readMetadata(zr *zip.Reader) error {
	files, err := zipExtFinder(zr, ".json")
	if err != nil {
		return err
	}

	for _, file := range files {
		name, _ := splitExt(file.FileInfo().Name())

		// name is 0-metadata.json
		idx, err := strconv.Atoi(strings.Split(name, "-")[0])
		if err != nil {
			return errors.New("error in metadata .json filename")
		}

		if len(z.Pages) < idx {
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

// zipReaderFromIOReader transforms a io.Reader to a zip.Reader.
func zipReaderFromIOReader(r io.Reader) (*zip.Reader, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	// use a bytes.Reader as it implements io.ReadAt
	br := bytes.NewReader(b)

	zr, err := zip.NewReader(br, br.Size())
	if err != nil {
		return nil, err
	}

	return zr, nil
}
