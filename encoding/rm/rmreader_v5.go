
package rm

import (
	"bytes"
	"encoding/binary"
	"fmt"
)
const HeaderV5    = "reMarkable .lines file, version=5          "
type rmReaderV5 struct {
	bytes.Reader
}
const (
	BallPointV2   BrushType = 15
	MarkerV2      BrushType = 16
	FinelinerV2   BrushType = 17
	SharpPencilV2 BrushType = 13
	TiltPencilV2  BrushType = 14
	BrushV2       BrushType = 12
	HighlighterV2 BrushType = 18
)

func remapBrushType(brushType BrushType) BrushType {
    switch (brushType) {
        case BallPointV2:
            return BallPoint
        case MarkerV2:
            return Marker
        case FinelinerV2:
            return Fineliner
        case SharpPencilV2:
            return SharpPencil
        case TiltPencilV2:
            return TiltPencil
        case BrushV2:
            return Brush
        case HighlighterV2:
            return Highlighter
        default:
            return brushType

    }
}
func (r *rmReaderV5) readNumber() (uint32, error) {

	var nb uint32
	if err := binary.Read(r, binary.LittleEndian, &nb); err != nil {
		return 0, fmt.Errorf("Wrong number read")
	}
	return nb, nil
}

func (r *rmReaderV5) readLine() (Line, error) {
	var line Line

    var brushType BrushType
	if err := binary.Read(r, binary.LittleEndian, &brushType); err != nil {
		return line, fmt.Errorf("Failed to read line")
	}
    line.BrushType = remapBrushType(brushType)

	if err := binary.Read(r, binary.LittleEndian, &line.BrushColor); err != nil {
		return line, fmt.Errorf("Failed to read line")
	}

	if err := binary.Read(r, binary.LittleEndian, &line.Padding); err != nil {
		return line, fmt.Errorf("Failed to read line")
	}

	if err := binary.Read(r, binary.LittleEndian, &line.Unknown); err != nil {
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

func (r *rmReaderV5) readPoint() (Point, error) {
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
