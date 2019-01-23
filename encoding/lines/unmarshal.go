package lines

import (
	"bytes"
)

// UnmarshalBinary implements encoding.MarshalBinary for
// transforming bytes into a notebook
// TODO
func (n Notebook) UnmarshalBinary(data []byte) error {
	return nil
}

type linesReader struct {
	r  *bytes.Reader
	nb Notebook
}

func newLinesReader(data []byte) *linesReader {
	lr := linesReader{
		bytes.NewReader(data),
		Notebook{},
	}
	return lr
}

// TODO
func (lr linesReader) readHeader() error {
	return nil
}

// TODO
func (lr linesReader) readNbPages() (uint32, error) {
	return 0, nil
}

// TODO
func (lr linesReader) readNbLayers() (uint32, error) {
	return 0, nil
}

// TODO
func (lr linesReader) readNbLines() (uint32, error) {
	return 0, nil
}

// TODO
func (lr linesReader) readLine() (Line, error) {
	return nil, nil
}

// TODO
func (lr linesReader) readNbPoints() (uint32, error) {
	return 0, nil
}

// TODO
func (lr linesReader) readPoint() (Point, error) {
	return nil, nil
}
