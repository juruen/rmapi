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
	"sort"
	"strconv"
	"strings"
)

// A Page represents a note page.
type Page struct {
	Data      *zip.File
	Metadata  *zip.File
	Thumbnail *zip.File
}

// Reader will parse specific files of the remarkable zip file.
type Reader struct {
	Content  *zip.File
	Pagedata *zip.File
	Pages    []Page
	Pdf      *zip.File
	Epub     *zip.File
	UUID     string
}

// OpenReader opens a reader from a zip file name.
// The UUID is taken from the Content or Pagadata file name.
func OpenReader(name string) (*Reader, error) {
	r := Reader{}
	pages := make(map[int]Page)

	zipRead, err := zip.OpenReader(name)
	if err != nil {
		return &r, err
	}

	for _, file := range zipRead.File {
		if !file.FileInfo().IsDir() {

			name, ext := splitFilenameExtension(file.FileInfo().Name())

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
				pages[idx] = Page{
					Data:      pages[idx].Data,
					Metadata:  file,
					Thumbnail: pages[idx].Thumbnail,
				}
				continue

			case ".rm":
				idx, err := strconv.Atoi(name)
				if err != nil {
					return &r, fmt.Errorf("error in .rm filename")
				}
				pages[idx] = Page{
					Data:      file,
					Metadata:  pages[idx].Metadata,
					Thumbnail: pages[idx].Thumbnail,
				}
				continue

			case ".pagedata":
				r.Pagedata = file
				r.UUID = name
				continue

			case ".jpg":
				idx, err := strconv.Atoi(name)
				if err != nil {
					return &r, fmt.Errorf("error in .jpg filename")
				}
				pages[idx] = Page{
					Data:      pages[idx].Data,
					Metadata:  pages[idx].Metadata,
					Thumbnail: file,
				}
				continue

			default:
				return &r, fmt.Errorf("file unknown")
			}
		}
	}

	// Flatten page map to slices
	r.Pages = make([]Page, nbPages(pages))
	for k, v := range pages {
		r.Pages[k] = v
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
	for i, page := range r.Pages {
		if page.Data != nil {
			fmt.Fprintf(&o, "Page %d Data: %s\n", i, page.Data.FileInfo().Name())
		}
		if page.Metadata != nil {
			fmt.Fprintf(&o, "Page %d Metadata: %s\n", i, page.Metadata.FileInfo().Name())
		}
		if page.Thumbnail != nil {
			fmt.Fprintf(&o, "Page %d Thumbnail: %s\n", i, page.Thumbnail.FileInfo().Name())
		}
	}
	return o.String()
}

func splitFilenameExtension(name string) (string, string) {
	ext := filepath.Ext(name)
	return name[0 : len(name)-len(ext)], ext
}

func nbPages(pages map[int]Page) int {
	if len(pages) == 0 {
		return 0
	}

	sorted := make([]int, 0, len(pages))
	for k := range pages {
		sorted = append(sorted, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))
	return sorted[0] + 1
}
