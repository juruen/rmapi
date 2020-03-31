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
	//psize := creator.PageSize{DeviceWidth * PPI, DeviceHeight * PPI}
	//c.SetPageSize(psize)
	fmt.Printf("canvas width height %f %f %f \n", c.Width(), c.Height(), c.Width()/c.Height())
	psize := creator.PageSizeA4
	c.SetPageSize(psize)

	fmt.Printf("canvas width height %f %f %f \n", c.Width(), c.Height(), c.Width()/c.Height())
	fmt.Printf("canvas %f \n", creator.PageSizeA4[0]/creator.PageSizeLegal[0])
	fmt.Printf("canvas %f \n", creator.PageSizeA4[1]/creator.PageSizeLegal[1])
	fmt.Printf("page A4 %v \n", creator.PageSizeA4)
	fmt.Printf("page Legal %v \n", creator.PageSizeLegal)
	ratioX := float64(c.Width() / DeviceWidth)
	ratioY := float64(c.Height() / DeviceHeight)

	line := c.NewLine(0.0, 0.0, c.Width(), c.Height())
	line.SetLineWidth(0.6)
	black := creator.ColorRGBFromHex("#000000")
	line.SetColor(black)
	c.Draw(line)

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

		if !hasContent {
			continue
		}

		var maxx float32
		var maxy float32
		for _, layer := range pageAnnotations.Data.Layers {
			for _, line := range layer.Lines {
				if len(line.Points) < 1 {
					continue
				}
				var isHighlighter bool

				if line.BrushType == rm.HighlighterV5 {
					isHighlighter = true
				}

				for i := 1; i < len(line.Points); i++ {
					s := line.Points[i-1]
					x1 := float64(s.X) * ratioX
					y1 := float64(s.Y) * ratioY

					if s.X > maxx {
						maxx = s.X
					}
					if s.Y > maxy {
						maxy = s.Y
					}
					s = line.Points[i]
					x2 := float64(s.X) * ratioX
					y2 := float64(s.Y) * ratioY

					if s.X > maxx {
						maxx = s.X
					}
					if s.Y > maxy {
						maxy = s.Y
					}

					if isHighlighter {
						lineDef := annotator.LineAnnotationDef{X1: x1, Y1: y1, X2: x2, Y2: y2}
						lineDef.LineColor = pdf.NewPdfColorDeviceRGB(1.0, 0.0, 0.0)
						lineDef.Opacity = 0.50
						lineDef.LineWidth = 6.0
						ann, err := annotator.CreateLineAnnotation(lineDef)
						if err != nil {
							return err
						}
						page.AddAnnotation(ann)
					} else {
						line := c.NewLine(x1, y1, x2, y2)
						line.SetLineWidth(0.6)
						black := creator.ColorRGBFromHex("#000000")
						line.SetColor(black)
						c.Draw(line)
					}
				}
			}
		}
		fmt.Printf("maxX: %f maxY: %f\n", maxx, maxy)
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
	//bbox := pdf.PdfRectangle{Llx: 0, Lly: 0, Urx: c.Width(), Ury: c.Height()}
	var page *pdf.PdfPage
	if p.template == false && !p.options.AnnotationsOnly {
		var err error
		page, err = p.pdfReader.GetPage(pageNum)
		if err != nil {
			return nil, err
		}

		bbox, err := page.GetMediaBox()
		if err != nil {
			log.Error.Println(err)
		}

		bbox.Urx = c.Width()
		bbox.Ury = c.Height()
		bbox.Llx = 0
		bbox.Lly = 0
		//page.MediaBox = &bbox

		if err != nil {
			return nil, err
		}

		if err = c.AddPage(page); err != nil {
			return nil, err
		}
	} else {
		page = c.NewPage()
		c.SetPageMargins(0.0, 0.0, 0.0, 0.0)
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
