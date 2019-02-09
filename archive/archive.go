// Package archive is used to parse a .zip file retrieved
// by the API.
//
// Here is the content of an archive retried on the tablet as example:
// 384327f5-133e-49c8-82ff-30aa19f3cfa4.content
// 384327f5-133e-49c8-82ff-30aa19f3cfa4//0-metadata.json
// 384327f5-133e-49c8-82ff-30aa19f3cfa4//0.rm
// 384327f5-133e-49c8-82ff-30aa19f3cfa4.pagedata
// 384327f5-133e-49c8-82ff-30aa19f3cfa4.thumbnails/0.jpg
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

// File is a structured representation of all files
// located in a remarkable zip file.
type File struct {
	Content    *zip.File
	Pagedata   *zip.File
	Thumbnails []*zip.File
	Pages      []Page
	Pdf        *zip.File
	Epub       *zip.File
}

// Open a zip file and parse it to read its content.
func Open(name string) (File, error) {
	f := File{}

	zipRead, err := zip.OpenReader(name)
	if err != nil {
		return f, err
	}

	for _, file := range zipRead.File {
		if !file.FileInfo().IsDir() {

			filename := file.FileInfo().Name()
			ext := filepath.Ext(filename)
			name := filename[0 : len(filename)-len(ext)]

			switch ext {

			case ".content":
				f.Content = file
				continue

			case ".pdf":
				f.Pdf = file
				continue

			case ".epub":
				f.Epub = file
				continue

			case ".json":
				idx, err := strconv.Atoi(strings.Split(name, "-")[0])
				if err != nil {
					return f, fmt.Errorf("error in .json filename")
				}

				if len(f.Pages) <= idx {
					f.Pages = append(f.Pages, Page{})
				}
				f.Pages[idx].Metadata = file
				continue

			case ".rm":
				idx, err := strconv.Atoi(name)
				if err != nil {
					return f, fmt.Errorf("error in .rm filename")
				}

				if len(f.Pages) <= idx {
					f.Pages = append(f.Pages, Page{})
				}
				f.Pages[idx].Data = file
				continue

			case ".pagedata":
				f.Pagedata = file
				continue

			case ".jpg":
				f.Thumbnails = append(f.Thumbnails, file)
				continue

			default:
				return f, fmt.Errorf("file unknown")
			}
		}
	}
	return f, nil
}

func (f File) String() string {
	var o strings.Builder
	if f.Content != nil {
		fmt.Fprintf(&o, "Content: %s\n", f.Content.FileInfo().Name())
	}
	if f.Pagedata != nil {
		fmt.Fprintf(&o, "Pagedata: %s\n", f.Pagedata.FileInfo().Name())
	}
	if f.Epub != nil {
		fmt.Fprintf(&o, "Epub: %s\n", f.Epub.FileInfo().Name())
	}
	if f.Pdf != nil {
		fmt.Fprintf(&o, "Pdf: %s\n", f.Pdf.FileInfo().Name())
	}
	for i, thumb := range f.Thumbnails {
		fmt.Fprintf(&o, "Thumb %d: %s\n", i, thumb.FileInfo().Name())
	}
	for i, page := range f.Pages {
		fmt.Fprintf(&o, "Page %d Data: %s\n", i, page.Data.FileInfo().Name())
		fmt.Fprintf(&o, "Page %d Metadata: %s\n", i, page.Metadata.FileInfo().Name())
	}
	return o.String()
}
