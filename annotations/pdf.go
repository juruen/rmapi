package annotations

import (
	"github.com/jung-kurt/gofpdf"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

const (
	RmX = 1404
	RmY = 1872
	A4X = 210
	A4Y = 297
	ratioA4X = float32(A4X) / float32(RmX)
	ratioA4Y = float32(A4Y) / float32(RmY)
	linesExtension = ".rm"
)

type PdfGenerator struct {
	inputDirPath string
	outputFilePath string
}

type page struct {
	layers []Layer
}

func CreatePdfGenerator(inputDirPath, outputFilePath string) PdfGenerator {
	return PdfGenerator{inputDirPath: inputDirPath, outputFilePath: outputFilePath}
}

func (p PdfGenerator) Generate() error {
	files, err := ioutil.ReadDir(p.inputDirPath)
	if err != nil {
		return err
	}

	pages := make([]page, 0)
	for _, f := range p.rmFiles(files) {
		reader, err := CreateRmReader(path.Join(p.inputDirPath, f))
		if err != nil {
			return err
		}

		layers, err := reader.Parse()
		if err != nil {
			return err
		}

		pages = append(pages, page{layers})
	}

	if err = p.generatePages(pages); err != nil {
		return err
	}

	return nil
}

func (p PdfGenerator) rmFiles(files []os.FileInfo) []string {
	rmFiles := make([]string, 0)
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), linesExtension) {
			continue
		}

		pageNumStr := strings.TrimSuffix(f.Name(), linesExtension)
		_, err := strconv.Atoi(pageNumStr)

		if err != nil {
			continue
		}

		rmFiles = append(rmFiles, f.Name())
	}

	sort.Slice(rmFiles, lineFileSorter(rmFiles))

	return rmFiles
}

func lineFileSorter(rmFiles []string) func(i int, j int) bool {
	return func(i int, j int) bool {
		a, _ := strconv.Atoi(strings.TrimSuffix(rmFiles[i], linesExtension))
		b, _ := strconv.Atoi(strings.TrimSuffix(rmFiles[j], linesExtension))

		return a < b
	}
}


func (p PdfGenerator) generatePages(pages []page) error {
	pdf := gofpdf.New("P", "mm", "A4", "")

	for _, page := range pages {
		pdf.AddPage()

		for _, layer := range page.layers {
			for _, stroke := range layer.Strokes {

				if len(stroke.Segments) < 1 {
					continue
				}

				for i := 1; i < len(stroke.Segments); i++ {
					s := stroke.Segments[i-1].toA4()
					x1 := s.X
					y1 := s.Y

					s = stroke.Segments[i].toA4()
					x2 := s.X
					y2 := s.Y

					pdf.Line(float64(x1), float64(y1), float64(x2), float64(y2))
				}
			}
		}
	}

	return pdf.OutputFileAndClose(p.outputFilePath)
}

func (p PdfGenerator) setStyle(pdf *gofpdf.Fpdf, stroke Stroke) {
	pdf.SetDrawColor(0, 0,0)
	pdf.SetLineWidth(0)
	pdf.SetLineCapStyle("round")
}

func (s Segment) toA4() Segment {
	s.X *= ratioA4X
	s.Y *= ratioA4Y

	return s
}