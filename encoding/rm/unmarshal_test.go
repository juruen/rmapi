package rm

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func testUnmarshalBinary(t *testing.T, fn string, ver Version) {
	b, err := ioutil.ReadFile(fn)
	if err != nil {
		t.Errorf("can't open %s file", fn)
	}

	rm := New()
	err = rm.UnmarshalBinary(b)
	if err != nil {
		t.Error(err)
	}

	if rm.Version != ver {
		t.Error("wrong version parsed")
	}

	t.Log(rm)

	fmt.Println("unmarshaling complete")
}

func TestUnmarshalBinaryV5(t *testing.T) {
	testUnmarshalBinary(t, "test_v5.rm", V5)
}

func TestUnmarshalBinaryV3(t *testing.T) {
	testUnmarshalBinary(t, "test_v3.rm", V3)
}
