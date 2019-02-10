package archive

import (
	"io"
)

// Writer is for writing a remarkable .zip
// The uuid should be unique across notes.
type Writer struct {
	w    io.Writer
	uuid string
}

// NewWriter creates a Writer from an io.Writer.
// The uuid will be used for file names as done by
// the remarkable device.
func NewWriter(w io.Writer, uuid string) *Writer {
	return &Writer{w, uuid}
}

// CreateContent is for writing a content file.
// Content will be overriden if already existing.
func (w *Writer) CreateContent() (io.Writer, error) {
	return nil, nil
}

// CreatePagedata is for writing a pagedata file.
// Pagedata will be overriden if already existing.
func (w *Writer) CreatePagedata() (io.Writer, error) {
	return nil, nil
}

// CreatePdf is for writing a pdf file.
// Pdf will be overriden if already existing.
func (w *Writer) CreatePdf() (io.Writer, error) {
	return nil, nil
}

// CreateEpub is for writing an epub file.
// Epub will be overriden if already existing.
func (w *Writer) CreateEpub() (io.Writer, error) {
	return nil, nil
}

// CreatePage is for writing a page.
// The argument idx serves for file names.
// The functions returns two io.Writer.
//  - The first one is for the data
//  - The second is for metadata
// Data and metadata will be overriden if already existing.
func (w *Writer) CreatePage(idx int) (io.Writer, io.Writer, error) {
	return nil, nil, nil
}

// CreateThumbnail is for writing a thumbnail file.
// The argument idx is used for file names.
// Thumbnail will be overriden if already existing.
func (w *Writer) CreateThumbnail(idx int) (io.Writer, error) {
	return nil, nil
}
