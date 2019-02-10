package archive

import "testing"

func TestOpenReader(t *testing.T) {
	r, err := OpenReader("test.zip")
	if err != nil {
		t.Error(err)
	}
	t.Log(r)
}
