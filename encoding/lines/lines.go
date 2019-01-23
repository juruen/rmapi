// Package lines provides primitives for encoding and decoding
// the .lines format which is a proprietary format created by
// Remarkable to store the data of a drawing made with the device.
//
// Axel Huebl has made a great job of understanding this binary format and
// has written an excellent blog post that helped a lot for writting this package.
// https://plasma.ninja/blog/devices/remarkable/binary/format/2017/12/26/reMarkable-lines-file-format.html
// As well, he has its own implementation of this decoder in C++ at this repository.
// https://github.com/ax3l/lines-are-beautiful
//
// As Ben Johnson says, "In the Go standard library, we use the term encoding
// and marshaling for two separate but related ideas. An encoder in Go is an object
// that applies structure to a stream of bytes while marshaling refers
// to applying structure to bounded, in-memory bytes."
// https://medium.com/go-walkthrough/go-walkthrough-encoding-package-bc5e912232d
//
// We will follow this convention and refer to marshaling for this encoder/decoder
// because we want to transform a .lines binary into a bounded in-memory representation
// of a .lines file.
//
// To try to be as idiomatic as possible, this package implements the two following interfaces
// of the default encoding package (https://golang.org/pkg/encoding/).
//  - BinaryMarshaler
//  - BinaryUnmarshaler
//
// The scope of this package is defined as just the encoding/decoding of the .lines format.
// It will only deal with bytes and not files (one must take care of unzipping the archive
// taken from the device, extracting and providing the content of .lines file as bytes).
//
// This package won't be used for retrieving metadata or attached PDF, ePub files.
package lines

// Header starting a .lines binary file. This can help recognizing a .lines file.
const (
	Header    = "reMarkable lines with selections and layers"
	HeaderLen = 43
)

// Width and Height of the device in pixels.
const (
	Width  int = 1404
	Height int = 1872
)

// BrushColor defines the 3 colors of the brush.
type BrushColor int

// Mapping of the three colors.
const (
	Black BrushColor = 0
	Grey  BrushColor = 1
	White BrushColor = 2
)

// BrushType respresents the type of brush.
//
// The different types of brush are explained here:
// https://blog.remarkable.com/how-to-find-your-perfect-writing-instrument-for-notetaking-on-remarkable-f53c8faeab77
type BrushType int

// Mappings for brush types.
const (
	BallPoint   BrushType = 2
	Marker      BrushType = 3
	Fineliner   BrushType = 4
	SharpPencil BrushType = 7
	TiltPencil  BrushType = 1
	Brush       BrushType = 0
	Highlighter BrushType = 5
	Eraser      BrushType = 6
	EraseArea   BrushType = 8
)

// BrushSize represents the base brush sizes.
type BrushSize float32

// 3 different brush sizes are noticed.
const (
	Small  BrushSize = 1.875
	Medium BrushSize = 2.0
	Large  BrushSize = 2.125
)

// A Notebook represents an entire note.
type Notebook struct {
	Pages []Page
}

// A Page contains information regarding a single page
// and is composed of layers.
type Page struct {
	Layers []Layer
}

// A Layer contains lines.
type Layer struct {
	Lines []Line
}

// A Line is composed of points.
type Line struct {
	BrushType  BrushType
	BrushColor BrushColor
	BrushSize  BrushSize
	Points     []Point
}

// A Point has coordinates.
type Point struct {
	X           float32
	Y           float32
	PenPressure float32
	XRotation   float32
	YRotation   float32
}

// New helps creating an empty note.
// By mashaling an empty Notebook and exporting it
// to the device, we should generate an empty note
// as if it were created using the device itself.
// TODO
func New(n string) *Notebook {
	return &Notebook{}
}
