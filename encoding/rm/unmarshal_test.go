package rm

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func testUnmarshalBinary(t *testing.T, fn string, ver Version) *Rm {
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

	return rm
}

func TestUnmarshalBinaryV5(t *testing.T) {
	rm := testUnmarshalBinary(t, "test_v5.rm", V5)
	for _, layer := range rm.Layers {
		for _, line := range layer.Lines {
			if line.BrushSize != 2.0 {
				t.Error("Incorrectly parsing BrushSize")
			}
		}
	}
}

func TestUnmarshalBinaryV3(t *testing.T) {
	testUnmarshalBinary(t, "test_v3.rm", V3)
}
