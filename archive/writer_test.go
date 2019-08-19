package archive

import (
	"os"
	"testing"
)

func TestWrite(t *testing.T) {
	zip := NewZip()
	zip.Content.FileType = "pdf"
	zip.Content.PageCount = 1
	zip.Pages = append(zip.Pages, Page{Pagedata: "Blank"})
	zip.Pdf = []byte{'p', 'd', 'f'}

	// create test file
	file, err := os.Create("write.zip")
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	// read file into note
	err = zip.Write(file)
	if err != nil {
		t.Error(err)
	}
}
