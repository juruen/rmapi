package archive

import (
	"os"
	"testing"
)

func TestWrite(t *testing.T) {
	note := NewFile()
	note.Content.FileType = "pdf"
	note.Content.PageCount = 1
	note.Pages = append(note.Pages, Page{Pagedata: "Blank"})
	note.Pdf = []byte{'p', 'd', 'f'}

	// create test file
	zip, err := os.Create("write.zip")
	if err != nil {
		t.Error(err)
	}
	defer zip.Close()

	// read file into note
	err = note.Write(zip)
	if err != nil {
		t.Error(err)
	}
}
