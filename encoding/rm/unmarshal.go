package rm

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// UnmarshalBinary implements encoding.UnmarshalBinary for
// transforming bytes into a Rm page
func (rm *Rm) UnmarshalBinary(data []byte) error {
	lr := newRmReader(data)
	if err := lr.checkHeader(); err != nil {
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

type rmReader struct {
	bytes.Reader
}

func newRmReader(data []byte) rmReader {
	br := bytes.NewReader(data)
	return rmReader{*br}
}

func (r *rmReader) checkHeader() error {
	buf := make([]byte, HeaderLen)

	n, err := r.Read(buf)
	if err != nil {
		return err
	}

	if n != HeaderLen || string(buf) != Header {
		return fmt.Errorf("Wrong header")
	}
	return nil
}

func (r *rmReader) readNumber() (uint32, error) {
	var nb uint32
	if err := binary.Read(r, binary.LittleEndian, &nb); err != nil {
		return 0, fmt.Errorf("Wrong number read")
	}
	return nb, nil
}

func (r *rmReader) readLine() (Line, error) {
	var line Line

	if err := binary.Read(r, binary.LittleEndian, &line.BrushType); err != nil {
		return line, fmt.Errorf("Failed to read line")
	}

	if err := binary.Read(r, binary.LittleEndian, &line.BrushColor); err != nil {
		return line, fmt.Errorf("Failed to read line")
	}

	if err := binary.Read(r, binary.LittleEndian, &line.Padding); err != nil {
		return line, fmt.Errorf("Failed to read line")
	}

	if err := binary.Read(r, binary.LittleEndian, &line.BrushSize); err != nil {
		return line, fmt.Errorf("Failed to read line")
	}

	nbPoints, err := r.readNumber()
	if err != nil {
		return line, err
	}

	if nbPoints == 0 {
		return line, nil
	}

	line.Points = make([]Point, nbPoints)

	for i := uint32(0); i < nbPoints; i++ {
		p, err := r.readPoint()
		if err != nil {
			return line, err
		}

		line.Points[i] = p
	}

	return line, nil
}

func (r *rmReader) readPoint() (Point, error) {
	var point Point

	if err := binary.Read(r, binary.LittleEndian, &point.X); err != nil {
		return point, fmt.Errorf("Failed to read point")
	}
	if err := binary.Read(r, binary.LittleEndian, &point.Y); err != nil {
		return point, fmt.Errorf("Failed to read point")
	}
	if err := binary.Read(r, binary.LittleEndian, &point.Speed); err != nil {
		return point, fmt.Errorf("Failed to read point")
	}
	if err := binary.Read(r, binary.LittleEndian, &point.Direction); err != nil {
		return point, fmt.Errorf("Failed to read point")
	}
	if err := binary.Read(r, binary.LittleEndian, &point.Width); err != nil {
		return point, fmt.Errorf("Failed to read point")
	}
	if err := binary.Read(r, binary.LittleEndian, &point.Pressure); err != nil {
		return point, fmt.Errorf("Failed to read point")
	}

	return point, nil
}
