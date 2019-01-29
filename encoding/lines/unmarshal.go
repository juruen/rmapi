package lines

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// UnmarshalBinary implements encoding.MarshalBinary for
// transforming bytes into a notebook
func (n Notebook) UnmarshalBinary(data []byte) error {
	lr := newLinesReader(data)
	if err := lr.checkHeader(); err != nil {
		return err
	}

	nbPages, err := lr.readNumber()
	if err != nil {
		return err
	}

	n.Pages = make([]Page, nbPages)
	for i := uint32(0); i < nbPages; i++ {
		nbLayers, err := lr.readNumber()
		if err != nil {
			return err
		}

		n.Pages[i].Layers = make([]Layer, nbLayers)
		for j := uint32(0); j < nbLayers; j++ {
			nbLines, err := lr.readNumber()
			if err != nil {
				return err
			}

			n.Pages[i].Layers[j].Lines = make([]Line, nbLines)
			for k := uint32(0); j < nbLines; j++ {
				line, err := lr.readLine()
				if err != nil {
					return err
				}
				n.Pages[i].Layers[j].Lines[k] = line
			}
		}
	}

	return nil
}

type linesReader struct {
	bytes.Reader
}

func newLinesReader(data []byte) linesReader {
	br := bytes.NewReader(data)
	return linesReader{*br}
}

func (lr *linesReader) checkHeader() error {
	buf := make([]byte, HeaderLen)

	n, err := lr.Read(buf)
	if err != nil {
		return err
	}

	if n != HeaderLen || string(buf) != Header {
		return fmt.Errorf("Wrong header")
	}
	return nil
}

func (lr *linesReader) readNumber() (uint32, error) {
	var nb uint32
	if err := binary.Read(lr, binary.LittleEndian, &nb); err != nil {
		return 0, fmt.Errorf("Wrong number read")
	}
	return nb, nil
}

func (lr *linesReader) readLine() (Line, error) {
	var line Line

	if err := binary.Read(lr, binary.LittleEndian, &line.BrushType); err != nil {
		return line, fmt.Errorf("Failed to read line")
	}

	if err := binary.Read(lr, binary.LittleEndian, &line.BrushColor); err != nil {
		return line, fmt.Errorf("Failed to read line")
	}

	if err := binary.Read(lr, binary.LittleEndian, &line.Padding); err != nil {
		return line, fmt.Errorf("Failed to read line")
	}

	if err := binary.Read(lr, binary.LittleEndian, &line.BrushSize); err != nil {
		return line, fmt.Errorf("Failed to read line")
	}

	nbPoints, err := lr.readNumber()
	if err != nil {
		return line, err
	}

	if nbPoints == 0 {
		return line, nil
	}

	line.Points = make([]Point, nbPoints)

	for i := uint32(0); i < nbPoints; i++ {
		p, err := lr.readPoint()
		if err != nil {
			return line, err
		}

		line.Points[i] = p
	}

	return line, nil
}

func (lr *linesReader) readPoint() (Point, error) {
	var point Point

	if err := binary.Read(lr, binary.LittleEndian, &point.X); err != nil {
		return point, fmt.Errorf("Failed to read point")
	}
	if err := binary.Read(lr, binary.LittleEndian, &point.Y); err != nil {
		return point, fmt.Errorf("Failed to read point")
	}
	if err := binary.Read(lr, binary.LittleEndian, &point.PenPressure); err != nil {
		return point, fmt.Errorf("Failed to read point")
	}
	if err := binary.Read(lr, binary.LittleEndian, &point.XRotation); err != nil {
		return point, fmt.Errorf("Failed to read point")
	}
	if err := binary.Read(lr, binary.LittleEndian, &point.YRotation); err != nil {
		return point, fmt.Errorf("Failed to read point")
	}

	return point, nil
}
