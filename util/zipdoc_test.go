package util

import (
	"fmt"
	"testing"
)

func TestZipFile(t *testing.T) {
	d, err := CreateZipDocument("1234", "zipdoc_test.pdf")
	fmt.Println(d)
	if err != nil {
		t.Error(err)
	}

}
