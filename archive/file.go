package archive

import (
	"github.com/juruen/rmapi/encoding/rm"
)

// Set the default pagedata template to Blank
const defaultPagadata string = "Blank"

// Zip represents an entire Remarkable archive file.
type Zip struct {
	Content Content
	Pages   []Page
	Pdf     []byte
	Epub    []byte
	UUID    string
}

// NewZip creates a File with sane defaults.
func NewZip() *Zip {
	content := Content{
		ExtraMetadata: ExtraMetadata{
			LastBrushColor:           "Black",
			LastBrushThicknessScale:  "2",
			LastColor:                "Black",
			LastEraserThicknessScale: "2",
			LastEraserTool:           "Eraser",
			LastPen:                  "Ballpoint",
			LastPenColor:             "Black",
			LastPenThicknessScale:    "2",
			LastPencil:               "SharpPencil",
			LastPencilColor:          "Black",
			LastPencilThicknessScale: "2",
			LastTool:                 "SharpPencil",
			ThicknessScale:           "2",
		},
		FileType:       "",
		FontName:       "",
		LastOpenedPage: 0,
		LineHeight:     -1,
		Margins:        100,
		Orientation:    "portrait",
		PageCount:      0,
		Pages:          []string{},
		TextScale:      1,
		Transform: Transform{
			M11: 1,
			M12: 0,
			M13: 0,
			M21: 0,
			M22: 1,
			M23: 0,
			M31: 0,
			M32: 0,
			M33: 1,
		},
	}

	return &Zip{
		Content: content,
	}
}

// A Page represents a note page.
type Page struct {
	// Data is the rm binary encoded file representing the drawn content
	Data *rm.Rm
	// Metadata is a json file containing information about layers
	Metadata Metadata
	// Thumbnail is a small image of the overall page
	Thumbnail []byte
	// Pagedata contains the name of the selected background template
	Pagedata string
}

// Metadata represents the structure of a .metadata json file associated to a page.
type Metadata struct {
	Layers []Layer `json:"layers"`
}

// Layers is a struct contained into a Metadata struct.
type Layer struct {
	Name string `json:"name"`
}

// Content represents the structure of a .content json file.
type Content struct {
	ExtraMetadata ExtraMetadata `json:"extraMetadata"`

	// FileType is "pdf", "epub" or empty for a simple note
	FileType       string `json:"fileType"`
	FontName       string `json:"fontName"`
	LastOpenedPage int    `json:"lastOpenedPage"`
	LineHeight     int    `json:"lineHeight"`
	Margins        int    `json:"margins"`
	// Orientation can take "portrait" or "landscape".
	Orientation string `json:"orientation"`
	PageCount   int    `json:"pageCount"`
	// Pages is a list of page IDs
	Pages     []string `json:"pages"`
	TextScale int      `json:"textScale"`

	Transform Transform `json:"transform"`
}

// ExtraMetadata is a struct contained into a Content struct.
type ExtraMetadata struct {
	LastBrushColor           string `json:"LastBrushColor"`
	LastBrushThicknessScale  string `json:"LastBrushThicknessScale"`
	LastColor                string `json:"LastColor"`
	LastEraserThicknessScale string `json:"LastEraserThicknessScale"`
	LastEraserTool           string `json:"LastEraserTool"`
	LastPen                  string `json:"LastPen"`
	LastPenColor             string `json:"LastPenColor"`
	LastPenThicknessScale    string `json:"LastPenThicknessScale"`
	LastPencil               string `json:"LastPencil"`
	LastPencilColor          string `json:"LastPencilColor"`
	LastPencilThicknessScale string `json:"LastPencilThicknessScale"`
	LastTool                 string `json:"LastTool"`
	ThicknessScale           string `json:"ThicknessScale"`
}

// Transform is a struct contained into a Content struct.
type Transform struct {
	M11 int `json:"m11"`
	M12 int `json:"m12"`
	M13 int `json:"m13"`
	M21 int `json:"m21"`
	M22 int `json:"m22"`
	M23 int `json:"m23"`
	M31 int `json:"m31"`
	M32 int `json:"m32"`
	M33 int `json:"m33"`
}
