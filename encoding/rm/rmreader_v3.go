package rm

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const HeaderV3    = "reMarkable .lines file, version=3          "
type rmReaderV3 struct {
	bytes.Reader
}
func (r *rmReaderV3) readNumber() (uint32, error) {
	var nb uint32
	if err := binary.Read(r, binary.LittleEndian, &nb); err != nil {
		return 0, fmt.Errorf("Wrong number read")
	}
	return nb, nil
}

func (r *rmReaderV3) readLine() (Line, error) {
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

func (r *rmReaderV3) readPoint() (Point, error) {
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
