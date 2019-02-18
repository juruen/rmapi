package archive

import "testing"

func TestOpenReader(t *testing.T) {
	r, err := OpenReader("test_reader.zip")
	if err != nil {
		t.Error(err)
	}
	t.Log(r)
}
