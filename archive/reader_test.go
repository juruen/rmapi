package archive

import (
	"os"
	"testing"
)

func TestRead(t *testing.T) {
	zip := NewZip()

	// open test file
	file, err := os.Open("test.zip")
	if err != nil {
		t.Error(err)
	}
	defer file.Close()

	// read file into note
	err = zip.Read(file)
	if err != nil {
		t.Error(err)
	}
}
