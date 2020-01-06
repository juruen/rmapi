package rm

import (
	"bytes"
	"fmt"
)

// UnmarshalBinary implements encoding.UnmarshalBinary for
// transforming bytes into a Rm page
func (rm *Rm) UnmarshalBinary(data []byte) error {
	lr, err := newRmReader(data)
	if  err != nil {
		return err
	}

	nbLayers, err := lr.readNumber()
	if err != nil {
		return err
	}

	rm.Layers = make([]Layer, nbLayers)
	for i := uint32(0); i < nbLayers; i++ {
		nbLines, err := lr.readNumber()
		if err != nil {
			return err
		}

		rm.Layers[i].Lines = make([]Line, nbLines)
		for j := uint32(0); j < nbLines; j++ {
			line, err := lr.readLine()
			if err != nil {
				return err
			}
			rm.Layers[i].Lines[j] = line
		}
	}

	return nil
}

type rmReader interface {
    readNumber() (uint32, error) 
    readLine() (Line, error) 
    readPoint() (Point, error) 
}

func newRmReader(data []byte) (rmReader, error) {
	br := bytes.NewReader(data)
	buf := make([]byte, HeaderLen)

	n, err := br.Read(buf)
	if err != nil {
		return nil, err
	}

	if n != HeaderLen {
		return nil, fmt.Errorf("Wrong header")
	}
    switch string(buf) {
        case HeaderV5:
            return &rmReaderV5{*br},nil
        case HeaderV3:
            return &rmReaderV3{*br},nil
        default:
            return nil, fmt.Errorf("Unknown header")
    }
}


