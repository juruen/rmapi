package annotations

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const (
	rmHeader = "reMarkable .lines file, version=3          "
	rmHeaderLen = 43
)

type rmReader struct {
	file io.Reader
	path string
}

type Layer struct {
	Strokes []Stroke
}

type Stroke struct {
	Pen uint32
	Color uint32
	Unknown uint32
	Width float32
	Segments []Segment
}

type Segment struct {
	X float32
	Y float32
	Pressure float32
	Tilt float32
	Unknown1 float32
	Unknown2 float32
}

func CreateRmReader(path string) (rmReader, error) {
	file, err := os.Open(path)
	if err != nil {
		return rmReader{}, err
	}

	return rmReader{file: file, path: path}, nil
}

func (rm rmReader) Parse() ([]Layer, error) {
	if err := rm.checkHeader(); err != nil {
		return nil, err
	}

	numLayers, err := rm.numLayers()

	if err != nil {
		return nil, err
	}

	layers := make([]Layer, numLayers)

	for i := uint32(0); i < numLayers; i++ {
		numStrokes, err := rm.numStrokes()
		if err != nil {
			return nil, err
		}

		strokes := make([]Stroke, numStrokes)
		for j := uint32(0); j < numStrokes; j++ {
			stroke, err := rm.parseStroke()
			if err != nil {
				return nil, err
			}

			strokes[j] = stroke
		}

		layers[i].Strokes = strokes
	}

	return layers, nil
}

func (rm rmReader) checkHeader() error {
	headerLen := rmHeaderLen
	buf := make([]byte, rmHeaderLen)

	n, err := rm.file.Read(buf)

	if err != nil {
		return err
	}

	if n != headerLen {
		return fmt.Errorf("%s is not a valid rm file, invalid header length", rm.path)
	}

	fileId := string(buf[:rmHeaderLen])

	if fileId != rmHeader {
		return fmt.Errorf("%s is not a valid rm file, wrong header %s", rm.path, fileId)
	}

	return nil
}

func (rm rmReader) numLayers() (uint32, error) {
	var numLayers uint32

	if err := binary.Read(rm.file, binary.LittleEndian, &numLayers); err != nil {
		return 0, fmt.Errorf("%s is not a valid rm file, wrong number of layers", rm.path)
	}

	return numLayers, nil
}

func (rm rmReader) numStrokes() (uint32, error) {
	var numStrokes uint32

	if err := binary.Read(rm.file, binary.LittleEndian, &numStrokes); err != nil {
		return 0, fmt.Errorf("%s is not a valid rm file, wrong number of strokes", rm.path)
	}

	return numStrokes, nil
}

func (rm rmReader) parseStroke() (Stroke, error) {
	var stroke Stroke

	if err := binary.Read(rm.file, binary.LittleEndian, &stroke.Pen); err != nil {
		return stroke, fmt.Errorf("%s: failed to parse stroke", rm.path)
	}

	if err := binary.Read(rm.file, binary.LittleEndian, &stroke.Color); err != nil {
		return stroke, fmt.Errorf("%s: failed to parse stroke", rm.path)
	}

	if err := binary.Read(rm.file, binary.LittleEndian, &stroke.Unknown); err != nil {
		return stroke, fmt.Errorf("%s: failed to parse stroke", rm.path)
	}

	if err := binary.Read(rm.file, binary.LittleEndian, &stroke.Width); err != nil {
		return stroke, fmt.Errorf("%s: failed to parse stroke", rm.path)
	}

	var numSegments uint32
	if err := binary.Read(rm.file, binary.LittleEndian, &numSegments); err != nil {
		return stroke, fmt.Errorf("%s: failed to parse stroke", rm.path)
	}

	if numSegments == 0 {
		return stroke, nil
	}

	stroke.Segments = make([]Segment, numSegments)

	for i := uint32(0); i < numSegments; i++ {
		s, err := rm.parseSegment()

		if err != nil {
			return stroke, fmt.Errorf("%s: failed to parse segment", rm.path)
		}

		stroke.Segments[i] = s
	}

	return stroke, nil
}


func (rm rmReader) parseSegment() (Segment, error) {
	var segment Segment

	if err := binary.Read(rm.file, binary.LittleEndian, &segment.X); err != nil {
		return segment, fmt.Errorf("%s: failed to parse segment", rm.path)
	}

	if err := binary.Read(rm.file, binary.LittleEndian, &segment.Y); err != nil {
		return segment, fmt.Errorf("%s: failed to parse segment", rm.path)
	}

	if err := binary.Read(rm.file, binary.LittleEndian, &segment.Pressure); err != nil {
		return segment, fmt.Errorf("%s: failed to parse segment", rm.path)
	}

	if err := binary.Read(rm.file, binary.LittleEndian, &segment.Tilt); err != nil {
		return segment, fmt.Errorf("%s: failed to parse segment", rm.path)
	}

	if err := binary.Read(rm.file, binary.LittleEndian, &segment.Unknown1); err != nil {
		return segment, fmt.Errorf("%s: failed to parse segment", rm.path)
	}

	if err := binary.Read(rm.file, binary.LittleEndian, &segment.Unknown2); err != nil {
		return segment, fmt.Errorf("%s: failed to parse segment", rm.path)
	}

	return segment, nil
}