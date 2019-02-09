package archive

import "testing"

func TestOpen(t *testing.T) {
	r, err := Open("test.zip")
	if err != nil {
		t.Error(err)
	}
	t.Log(r)
}
