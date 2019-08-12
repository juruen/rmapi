package archive

import (
	"os"
	"testing"

	"github.com/davecgh/go-spew/spew"
)

func TestRead(t *testing.T) {
	note := NewFile()

	// open test file
	zip, err := os.Open("test.zip")
	if err != nil {
		t.Error(err)
	}
	defer zip.Close()

	// read file into note
	err = note.Read(zip)
	if err != nil {
		t.Error(err)
	}

	spew.Dump(note)
}
