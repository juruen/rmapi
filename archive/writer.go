package archive

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// Write writes an archive file from a Zip struct.
// It automatically generates a uuid if not already
// defined in the struct.
func (z *Zip) Write(w io.Writer) error {
	// generate random uuid if not defined
	if z.UUID == "" {
		z.UUID = uuid.New().String()
	}

	archive := zip.NewWriter(w)

	if err := z.writeContent(archive); err != nil {
		return err
	}

	if err := z.writePdf(archive); err != nil {
		return err
	}

	if err := z.writeEpub(archive); err != nil {
		return err
	}

	if err := z.writePagedata(archive); err != nil {
		return err
	}

	if err := z.writeThumbnails(archive); err != nil {
		return err
	}

	if err := z.writeMetadata(archive); err != nil {
		return err
	}

	if err := z.writeData(archive); err != nil {
		return err
	}

	archive.Close()

	return nil
}

// writeContent writes the .content file to the archive.
func (z *Zip) writeContent(zw *zip.Writer) error {
	bytes, err := json.MarshalIndent(&z.Content, "", "    ")
	if err != nil {
		return err
	}

	name := fmt.Sprintf("%s.content", z.UUID)

	w, err := addToZip(zw, name)
	if err != nil {
		return err
	}

	if _, err := w.Write(bytes); err != nil {
		return err
	}

	return nil
}

// writePdf writes a pdf file to the archive if existing in the struct.
func (z *Zip) writePdf(zw *zip.Writer) error {
	// skip if no pdf
	if z.Pdf == nil {
		return nil
	}

	name := fmt.Sprintf("%s.pdf", z.UUID)

	w, err := addToZip(zw, name)
	if err != nil {
		return err
	}

	if _, err := w.Write(z.Pdf); err != nil {
		return err
	}

	return nil
}

// writeEpub writes an epub file to the archive if existing in the struct.
func (z *Zip) writeEpub(zw *zip.Writer) error {
	// skip if no epub
	if z.Epub == nil {
		return nil
	}

	name := fmt.Sprintf("%s.epub", z.UUID)

	w, err := addToZip(zw, name)
	if err != nil {
		return err
	}

	if _, err := w.Write(z.Epub); err != nil {
		return err
	}

	return nil
}

// writePagedata writes a .pagedata file containing
// the name of background templates for each page (one per line).
func (z *Zip) writePagedata(zw *zip.Writer) error {
	// don't add pagedata file if no pages
	if len(z.Pages) == 0 {
		return nil
	}

	name := fmt.Sprintf("%s.pagedata", z.UUID)

	w, err := addToZip(zw, name)
	if err != nil {
		return err
	}

	bw := bufio.NewWriter(w)
	for _, page := range z.Pages {
		template := page.Pagedata

		// set default if empty
		if template == "" {
			template = defaultPagadata
		}

		bw.WriteString(template + "\n")
	}

	// write to the underlying io.Writer
	bw.Flush()

	return nil
}

// writeThumbnails writes thumbnail files for each page
// in the archive.
func (z *Zip) writeThumbnails(zw *zip.Writer) error {
	for idx, page := range z.Pages {
		if page.Thumbnail == nil {
			continue
		}

		folder := fmt.Sprintf("%s.thumbnail", z.UUID)
		name := fmt.Sprintf("%d.jpg", idx)
		fn := filepath.Join(folder, name)

		w, err := addToZip(zw, fn)
		if err != nil {
			return err
		}

		if _, err := w.Write(page.Thumbnail); err != nil {
			return err
		}

	}

	return nil
}

// writeMetadata writes .json metadata files for each page
// in the archive.
func (z *Zip) writeMetadata(zw *zip.Writer) error {
	for idx, page := range z.Pages {
		// if no layers available, don't write the metadata file
		if len(page.Metadata.Layers) == 0 {
			continue
		}

		name := fmt.Sprintf("%d-metadata.json", idx)
		fn := filepath.Join(z.UUID, name)

		w, err := addToZip(zw, fn)
		if err != nil {
			return err
		}

		bytes, err := json.MarshalIndent(&page.Metadata, "", "    ")
		if err != nil {
			return err
		}

		if _, err := w.Write(bytes); err != nil {
			return err
		}

	}

	return nil
}

// writeData writes .rm data files for each page
// in the archive.
func (z *Zip) writeData(zw *zip.Writer) error {
	for idx, page := range z.Pages {
		if page.Data == nil {
			continue
		}

		name := fmt.Sprintf("%d.rm", idx)
		fn := filepath.Join(z.UUID, name)

		w, err := addToZip(zw, fn)
		if err != nil {
			return err
		}

		bytes, err := page.Data.MarshalBinary()
		if err != nil {
			return err
		}

		if _, err := w.Write(bytes); err != nil {
			return err
		}

	}

	return nil
}

// addToZip takes a zip.Writer in parameter and creates an io.Writer
// to write the content of a file to add to the zip.
func addToZip(zw *zip.Writer, name string) (io.Writer, error) {
	h := &zip.FileHeader{
		Name:         name,
		Method:       zip.Store,
		ModifiedTime: uint16(time.Now().UnixNano()),
		ModifiedDate: uint16(time.Now().UnixNano()),
	}

	writer, err := zw.CreateHeader(h)
	if err != nil {
		return nil, err
	}

	return writer, nil
}
