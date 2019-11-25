package rm

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestUnmarshalBinary(t *testing.T) {
	b, err := ioutil.ReadFile("test_v5.rm")
	if err != nil {
		t.Error("can't open test.rm file")
	}

	rm := New()
	err = rm.UnmarshalBinary(b)
	if err != nil {
		t.Error(err)
	}

	t.Log(rm)

	fmt.Println("unmarshaling complete")
	// Output:
	// unmarshaling complete
}
