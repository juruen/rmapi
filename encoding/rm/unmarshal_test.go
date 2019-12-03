package rm

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestUnmarshalBinaryV5(t *testing.T) {
	b, err := ioutil.ReadFile("test_v5.rm")
	if err != nil {
		t.Error("can't open test_v5.rm file")
	}

	rm := New()
	err = rm.UnmarshalBinary(b)
	if err != nil {
		t.Error(err)
	}

	t.Log(rm)

	fmt.Println("unmarshaling complete")
}

func TestUnmarshalBinaryV3(t *testing.T) {
	b, err := ioutil.ReadFile("test_v3.rm")
	if err != nil {
		t.Error("can't open test_v3.rm file")
	}

	rm := New()
	err = rm.UnmarshalBinary(b)
	if err != nil {
		t.Error(err)
	}

	t.Log(rm)

	fmt.Println("unmarshaling complete")
}
