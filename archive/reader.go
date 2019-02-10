// Package archive is used to parse a .zip file retrieved
// by the API.
//
// Here is the content of an archive retried on the tablet as example:
// 384327f5-133e-49c8-82ff-30aa19f3cfa4.content
// 384327f5-133e-49c8-82ff-30aa19f3cfa4//0-metadata.json
// 384327f5-133e-49c8-82ff-30aa19f3cfa4//0.rm
// 384327f5-133e-49c8-82ff-30aa19f3cfa4.pagedata
// 384327f5-133e-49c8-82ff-30aa19f3cfa4.thumbnails/0.jpg
//
// As the .zip file from remarkable is simply a normal .zip file
// containing specific file formats, this package is only a wrapper
// around the standard zip package that will follow the same
// code architecture and that will help gathering one of
// those specific files more easily.
package archive

import (
	"archive/zip"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// A Page represents a note page.
type Page struct {
	Data     *zip.File
	Metadata *zip.File
}

// Reader will parse specific files of the remarkable zip file.
type Reader struct {
	Content    *zip.File
	Pagedata   *zip.File
	Thumbnails []*zip.File
	Pages      []Page
	Pdf        *zip.File
	Epub       *zip.File
	UUID       string
}

// OpenReader opens a reader from a zip file name.
// The UUID is taken from the Content or Pagadata file name.
func OpenReader(name string) (*Reader, error) {
	r := Reader{}

	zipRead, err := zip.OpenReader(name)
	if err != nil {
		return &r, err
	}

	for _, file := range zipRead.File {
		if !file.FileInfo().IsDir() {

			filename := file.FileInfo().Name()
			ext := filepath.Ext(filename)
			name := filename[0 : len(filename)-len(ext)]

			switch ext {

			case ".content":
				r.Content = file
				r.UUID = name
				continue

			case ".pdf":
				r.Pdf = file
				continue

			case ".epub":
				r.Epub = file
				continue

			case ".json":
				idx, err := strconv.Atoi(strings.Split(name, "-")[0])
				if err != nil {
					return &r, fmt.Errorf("error in .json filename")
				}

				if len(r.Pages) <= idx {
					r.Pages = append(r.Pages, Page{})
				}
				r.Pages[idx].Metadata = file
				continue

			case ".rm":
				idx, err := strconv.Atoi(name)
				if err != nil {
					return &r, fmt.Errorf("error in .rm filename")
				}

				if len(r.Pages) <= idx {
					r.Pages = append(r.Pages, Page{})
				}
				r.Pages[idx].Data = file
				continue

			case ".pagedata":
				r.Pagedata = file
				r.UUID = name
				continue

			case ".jpg":
				r.Thumbnails = append(r.Thumbnails, file)
				continue

			default:
				return &r, fmt.Errorf("file unknown")
			}
		}
	}
	return &r, nil
}

func (r Reader) String() string {
	var o strings.Builder
	if r.Content != nil {
		fmt.Fprintf(&o, "Content: %s\n", r.Content.FileInfo().Name())
	}
	if r.Pagedata != nil {
		fmt.Fprintf(&o, "Pagedata: %s\n", r.Pagedata.FileInfo().Name())
	}
	if r.Epub != nil {
		fmt.Fprintf(&o, "Epub: %s\n", r.Epub.FileInfo().Name())
	}
	if r.Pdf != nil {
		fmt.Fprintf(&o, "Pdf: %s\n", r.Pdf.FileInfo().Name())
	}
	for i, thumb := range r.Thumbnails {
		fmt.Fprintf(&o, "Thumb %d: %s\n", i, thumb.FileInfo().Name())
	}
	for i, page := range r.Pages {
		fmt.Fprintf(&o, "Page %d Data: %s\n", i, page.Data.FileInfo().Name())
		fmt.Fprintf(&o, "Page %d Metadata: %s\n", i, page.Metadata.FileInfo().Name())
	}
	return o.String()
}
