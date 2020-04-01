package annotations

import (
	"bytes"
	"fmt"
	"github.com/juruen/rmapi/archive"
	"github.com/juruen/rmapi/encoding/rm"
	"github.com/juruen/rmapi/log"
	annotator "github.com/unidoc/unipdf/v3/annotator"
	"github.com/unidoc/unipdf/v3/creator"
	pdf "github.com/unidoc/unipdf/v3/model"
	"os"
)

const (
	PPI          = 226
	DeviceHeight = 1872
	DeviceWidth  = 1404
)

type PdfGenerator struct {
	zipName        string
	outputFilePath string
	options        PdfGeneratorOptions
	pdfReader      *pdf.PdfReader
	template       bool
}

type PdfGeneratorOptions struct {
	AddPageNumbers  bool
	AllPages        bool
	AnnotationsOnly bool
}

func CreatePdfGenerator(zipName, outputFilePath string, options PdfGeneratorOptions) *PdfGenerator {
	return &PdfGenerator{zipName: zipName, outputFilePath: outputFilePath, options: options}
}

func normalized(p1, p2 rm.Point, ratioX, ratioY float64) (float64, float64, float64, float64) {
	return float64(p1.X) * ratioX, float64(p1.Y) * ratioY, float64(p2.X) * ratioX, float64(p2.Y) * ratioY
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
	c.SetPageSize(creator.PageSizeA4)

	ratioX := c.Width() / DeviceWidth
	//ratioY := c.Height() / DeviceHeight

	for i, pageAnnotations := range zip.Pages {
		hasContent := pageAnnotations.Data != nil

		// do not add a page when there are no annotaions
		if !p.options.AllPages && !hasContent {
			continue
		}

		page, err := p.addBackgroundPage(c, i+1)
		if err != nil {
			return err
		}
		if page == nil {
			log.Error.Fatal("page is null")
		}

		if err != nil {
			return err
		}
		if !hasContent {
			continue
		}

		for _, layer := range pageAnnotations.Data.Layers {
			for _, line := range layer.Lines {
				if len(line.Points) < 1 {
					continue
				}

				if line.BrushType == rm.HighlighterV5 {
					last := len(line.Points) - 1
					x1, y1, x2, y2 := normalized(line.Points[0], line.Points[last], ratioX, ratioX)
					// make horizontal lines only
					y2 = y1
					//todo: y cooridnates are reversed
					lineDef := annotator.LineAnnotationDef{X1: x1 - 1, Y1: c.Height() - y1, X2: x2, Y2: c.Height() - y2}
					lineDef.LineColor = pdf.NewPdfColorDeviceRGB(1.0, 1.0, 0.0) //yellow
					lineDef.Opacity = 0.5
					lineDef.LineWidth = 5.0
					ann, err := annotator.CreateLineAnnotation(lineDef)
					if err != nil {
						return err
					}
					page.AddAnnotation(ann)
				} else {
					for i := 1; i < len(line.Points); i++ {
						x1, y1, x2, y2 := normalized(line.Points[i-1], line.Points[i], ratioX, ratioX)
						line := c.NewLine(x1, y1, x2, y2)
						line.SetLineWidth(0.6)
						black := creator.ColorRGBFromHex("#000000")
						line.SetColor(black)
						c.Draw(line)
					}
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

func (p *PdfGenerator) addBackgroundPage(c *creator.Creator, pageNum int) (*pdf.PdfPage, error) {
	page := c.NewPage()

	if p.template == false && !p.options.AnnotationsOnly {
		page, err := p.pdfReader.GetPage(pageNum)
		if err != nil {
			return nil, err
		}
		block, err := creator.NewBlockFromPage(page)
		if err != nil {
			return nil, err
		}
		//convert: Letter->A4
		block.SetPos(0.0, 0.0)
		block.ScaleToWidth(c.Width())

		err = c.Draw(block)
		if err != nil {
			return nil, err
		}
	}

	if p.options.AddPageNumbers {
		c.DrawFooter(func(block *creator.Block, args creator.FooterFunctionArgs) {
			p := c.NewParagraph(fmt.Sprintf("%d", args.PageNum))
			p.SetFontSize(8)
			w := block.Width() - 20
			h := block.Height() - 10
			p.SetPos(w, h)
			block.Draw(p)
		})
	}
	return page, nil
}
