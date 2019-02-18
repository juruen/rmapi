package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"
	"time"
)

// Writer is for writing a remarkable .zip
// The uuid should be unique across notes.
type Writer struct {
	zipw *zip.Writer
	uuid string
}

// NewWriter creates a Writer from an io.Writer.
// The uuid will be used for file names as done by
// the remarkable device.
func NewWriter(w io.Writer, uuid string) *Writer {
	archive := zip.NewWriter(w)
	return &Writer{archive, uuid}
}

// Close finishes writing the zip file.
func (w *Writer) Close() error {
	return w.zipw.Close()
}

// CreateContent is for writing a content file.
func (w *Writer) CreateContent() (io.Writer, error) {
	fn := fmt.Sprintf("%s.content", w.uuid)
	return w.create(fn)
}

// CreatePagedata is for writing a pagedata file.
func (w *Writer) CreatePagedata() (io.Writer, error) {
	fn := fmt.Sprintf("%s.pagedata", w.uuid)
	return w.create(fn)
}

// CreatePdf is for writing a pdf file.
func (w *Writer) CreatePdf() (io.Writer, error) {
	fn := fmt.Sprintf("%s.pdf", w.uuid)
	return w.create(fn)
}

// CreateEpub is for writing an epub file.
func (w *Writer) CreateEpub() (io.Writer, error) {
	fn := fmt.Sprintf("%s.epub", w.uuid)
	return w.create(fn)
}

// CreatePage is for writing a page data.
// The argument idx serves for file names.
func (w *Writer) CreatePage(idx int) (io.Writer, error) {
	name := fmt.Sprintf("%d.rm", idx)
	fn := filepath.Join(w.uuid, name)
	return w.create(fn)
}

// CreatePageMetadata is for writing a page metadata.
// The argument idx serves for file names.
func (w *Writer) CreatePageMetadata(idx int) (io.Writer, error) {
	name := fmt.Sprintf("%d-metadata.json", idx)
	fn := filepath.Join(w.uuid, name)
	return w.create(fn)
}

// CreateThumbnail is for writing a thumbnail file.
// The argument idx is used for file names.
func (w *Writer) CreateThumbnail(idx int) (io.Writer, error) {
	folder := fmt.Sprintf("%s.thumbnail", w.uuid)
	name := fmt.Sprintf("%d.jpg", idx)
	fn := filepath.Join(folder, name)
	return w.create(fn)
}

func (w *Writer) create(name string) (io.Writer, error) {
	h := &zip.FileHeader{
		Name:         name,
		Method:       zip.Store,
		ModifiedTime: uint16(time.Now().UnixNano()),
		ModifiedDate: uint16(time.Now().UnixNano()),
	}

	writer, err := w.zipw.CreateHeader(h)
	if err != nil {
		return nil, err
	}

	return writer, nil
}
