package annotations

import (
	"github.com/juruen/rmapi/encoding/rm"
	"math"
)

// Just a translation of https://github.com/peerdavid/remapy/blob/master/model/render.py
// Use some ideas for golang code of https://github.com/rorycl/rm2pdf/blob/master/rmpdf/stroke.go

type StrokeSetting struct {
	Render    BrushRenderType
	Ratio     float32
	LastWidth float32
	Opacity   float32
	Length    int
	Line      rm.Line
	Point     rm.Point
}

func (p StrokeSetting) GetWidth() float32 {
	var width float32 = 0

	switch p.Render {
	case BallPointRender:
		width = (0.5 + p.Point.Pressure) + (1 * p.Point.Width) - 0.5*(p.Point.Speed/50)

	case MarkerRender:
		width = 0.9*((1*p.Point.Width)-0.4*p.Point.Direction) + (0.1 * p.LastWidth)

	case FinelineRender:
		width = (float32(math.Pow(0.5*float64(p.Line.BrushSize), 10))) * 3

	case SharpPencilRender:
		width = float32(math.Pow(float64(p.Line.BrushSize), 2))

	case TiltPencilRender:
		width = float32(math.Min(
			float64(0.5*((((0.8*float32(p.Line.BrushSize))+(0.5*p.Point.Pressure))*(1*p.Point.Width))-(0.25*float32(math.Pow(float64(p.Point.Direction), 1.8))))),
			float64(p.Line.BrushSize)*10))

	case BrushRender:
		width = 0.7 * (((1 + (1.4 * p.Point.Pressure)) * (1 * p.Point.Width)) - (0.5 * p.Point.Direction) - (0.5 * p.Point.Speed / 50))

	case HighlighterRender:
		width = 30

	case EraserRender:
		width = float32(p.Line.BrushSize) * 2

	case EraseAreaRender:
		width = float32(p.Line.BrushSize)

	case CalligraphyRender:
		width = 0.5*(((1+p.Point.Pressure)*(1*p.Point.Width))-0.3*p.Point.Direction) + (0.2 * p.LastWidth)

	}

	return width * p.Ratio
}

func (p StrokeSetting) GetOpacity() float32 {
	var opacity float32 = p.Opacity
	if p.Render == TiltPencilRender {
		opacity = float32(math.Max(0, math.Min(1,
			math.Max(0.05, math.Min(0.7, float64(p.Point.Pressure))))))
	}
	return opacity

}

func (p StrokeSetting) GetColour() float32 {
	var colour float32 = 0
	switch p.Line.BrushColor {
	case rm.Black:
		colour = 0
	case rm.Grey:
		colour = 0.5
	default:
		colour = 1
	}
	return colour
}

type BrushRenderType uint32

const (
	BallPointRender   BrushRenderType = 1
	MarkerRender      BrushRenderType = 2
	FinelineRender    BrushRenderType = 3
	SharpPencilRender BrushRenderType = 4
	TiltPencilRender  BrushRenderType = 5
	BrushRender       BrushRenderType = 6
	HighlighterRender BrushRenderType = 7
	EraserRender      BrushRenderType = 8
	EraseAreaRender   BrushRenderType = 9
	CalligraphyRender BrushRenderType = 10
)

// Map of pen numbers in a reMarkable binary .rm file
var StrokeMap = map[rm.BrushType]BrushRenderType{

	rm.BallPoint:   BallPointRender,
	rm.Marker:      MarkerRender,
	rm.Fineliner:   FinelineRender,
	rm.SharpPencil: SharpPencilRender,
	rm.TiltPencil:  TiltPencilRender,
	rm.Brush:       BrushRender,
	rm.Highlighter: HighlighterRender,
	rm.Eraser:      EraserRender,
	rm.EraseArea:   EraseAreaRender,

	// v5 brings new brush type IDs
	rm.BallPointV5:   BallPointRender,
	rm.MarkerV5:      MarkerRender,
	rm.FinelinerV5:   FinelineRender,
	rm.SharpPencilV5: SharpPencilRender,
	rm.TiltPencilV5:  TiltPencilRender,
	rm.BrushV5:       BrushRender,
	rm.HighlighterV5: HighlighterRender,
	rm.CalligraphyV5: CalligraphyRender,
}

var StrokeSettings = map[BrushRenderType]StrokeSetting{
	BallPointRender: {
		Render:  BallPointRender,
		Opacity: 1,
		Length:  5,
	},
	MarkerRender: {
		Render:  MarkerRender,
		Opacity: 1,
		Length:  3,
	},
	FinelineRender: {
		Render:  FinelineRender,
		Opacity: 1,
		Length:  1000,
	},
	SharpPencilRender: {
		Render:  SharpPencilRender,
		Opacity: 0.7,
		Length:  1000,
	},
	TiltPencilRender: {
		Render:  TiltPencilRender,
		Opacity: 1,
		Length:  2,
	},
	BrushRender: {
		Render:  BrushRender,
		Opacity: 1,
		Length:  2,
	},
	HighlighterRender: {
		Render:  HighlighterRender,
		Opacity: 0.2,
		Length:  2,
	},
	EraserRender: {
		Render:  EraserRender,
		Opacity: 1,
		Length:  10,
	},
	EraseAreaRender: {
		Render:  EraserRender,
		Opacity: 0,
		Length:  10,
	},
	CalligraphyRender: {
		Render:  CalligraphyRender,
		Opacity: 1,
		Length:  2,
	},
}

type PointRender struct {
	X       float64
	Y       float64
	Width   float64
	Opacity float64
	Colour  float64
	Render  BrushRenderType
}
