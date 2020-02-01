package annotations

import (
	"bytes"
	"os"

	"github.com/juruen/rmapi/archive"
	"github.com/unidoc/unipdf/v3/creator"
	pdf "github.com/unidoc/unipdf/v3/model"
)

const (
	ratioA4X = float32(0.443)
	ratioA4Y = float32(0.443)
)

type PdfGenerator struct {
	zipName        string
	outputFilePath string
	options        PdfGeneratorOptions
	pdfReader      *pdf.PdfReader
	template       bool
}

type PdfGeneratorOptions struct {
	AddPageNumbers bool
}

func CreatePdfGenerator(zipName, outputFilePath string, options PdfGeneratorOptions) *PdfGenerator {
	return &PdfGenerator{zipName: zipName, outputFilePath: outputFilePath, options: options}
}

func (p *PdfGenerator) Generate() error {
	file, err := os.Open(p.zipName)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	zip := archive.NewZip()

	fi, err := file.Stat()
	if err != nil {
		return err
	}

	err = zip.Read(file, fi.Size())
	if err != nil {
		return err
	}

	if err = p.initBackgroundPages(zip.Pdf); err != nil {
		return err
	}

	c := creator.New()
	for i, pageAnnotations := range zip.Pages {
		if err := p.addBackgroundPage(c, i+1); err != nil {
			return err
		}

		for _, layer := range pageAnnotations.Data.Layers {
			for _, line := range layer.Lines {
				if len(line.Points) < 1 {
					continue
				}

				for i := 1; i < len(line.Points); i++ {
					s := line.Points[i-1]
					x1 := s.X * ratioA4X
					y1 := s.Y * ratioA4Y

					s = line.Points[i]
					x2 := s.X * ratioA4X
					y2 := s.Y * ratioA4Y

					line := c.NewLine(float64(x1), float64(y1), float64(x2), float64(y2))
					line.SetLineWidth(0.6)
					black := creator.ColorRGBFromHex("#000000")
					line.SetColor(black)
					c.Draw(line)
				}
			}
		}
	}

	return c.WriteToFile(p.outputFilePath)
}

func (p *PdfGenerator) initBackgroundPages(pdfArr []byte) error {
	if len(pdfArr) > 0 {
		pdfReader, err := pdf.NewPdfReader(bytes.NewReader(pdfArr))
		if err != nil {
			return err
		}

		p.pdfReader = pdfReader
		p.template = false
		return nil
	}

	p.template = true
	return nil
}

func (p *PdfGenerator) addBackgroundPage(c *creator.Creator, pageNum int) error {
	if p.template == false {
		page, err := p.pdfReader.GetPage(pageNum)
		if err != nil {
			return err
		}

		if err = c.AddPage(page); err != nil {
			return err
		}
		return nil
	}

	c.NewPage()
	c.SetPageSize(creator.PageSizeA4)
	return nil
}
