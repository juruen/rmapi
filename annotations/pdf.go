package annotations

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/juruen/rmapi/encoding/rm"

	"github.com/juruen/rmapi/archive"
	"github.com/juruen/rmapi/log"
	//"github.com/unidoc/unipdf/v3/annotator"
	"github.com/unidoc/unipdf/v3/contentstream"
	"github.com/unidoc/unipdf/v3/contentstream/draw"
	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/creator"
	pdf "github.com/unidoc/unipdf/v3/model"
)

const (
	DeviceWidth   = float64(rm.Width)
	DeviceHeight  = float64(rm.Height)
	PathSkip      = 3
	GShighlighter = "GShiglighter"
	GSnormal      = "GS"
)

var rmPageSize = creator.PageSize{445, 594}

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
	AnnotationsOnly bool //export the annotations without the background/pdf
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

	if zip.Content.FileType == "epub" {
		return errors.New("only pdf and notebooks supported")
	}

	if err = p.initBackgroundPages(zip.Payload); err != nil {
		return err
	}

	if len(zip.Pages) == 0 {
		return errors.New("the document has no pages")
	}

	c := creator.New()
	if p.template {
		// use the standard page size
		c.SetPageSize(rmPageSize)
	}

	if p.pdfReader != nil && p.options.AllPages {
		outlines := p.pdfReader.GetOutlineTree()
		c.SetOutlineTree(outlines)
	}

	for i, pageAnnotations := range zip.Pages {
		hasContent := pageAnnotations.Data != nil

		// do not add a page when there are no annotations
		if !p.options.AllPages && !hasContent {
			continue
		}

		// TODO: check if page is horizontal and rotate

		page, err := p.addBackgroundPage(c, i+1)

		if err != nil {
			return err
		}

		ratio := c.Height() / c.Width()

		var scale float64
		if ratio < (float64(DeviceHeight) / float64(DeviceWidth)) {
			scale = c.Width() / DeviceWidth
		} else {
			scale = c.Height() / DeviceHeight
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

		contentCreator := contentstream.NewContentCreator()
		contentCreator.Add_q()

		// Setting transparency is a total mess and require a different Dictionary for each alpha value
		// PS: there is no documentation about this.
		var GSname core.PdfObjectName = GSnormal

		GS := core.MakeDict()
		GS.Set("CA", core.MakeFloat(float64(StrokeSettings[HighlighterRender].Opacity)))
		page.AddExtGState(GShighlighter, GS)

		GS = core.MakeDict()
		GS.Set("CA", core.MakeFloat(float64(StrokeSettings[FinelineRender].Opacity)))
		page.AddExtGState(GSnormal, GS)

		var GScount uint32 = 0

		for _, layer := range pageAnnotations.Data.Layers {
			for _, line := range layer.Lines {
				if len(line.Points) < 1 {
					continue
				}

				if StrokeMap[line.BrushType] == EraserRender || StrokeMap[line.BrushType] == EraseAreaRender {
					continue
				}

				ss, ok := StrokeSettings[StrokeMap[line.BrushType]]
				if !ok {
					ss = StrokeSettings[FinelineRender]
				}

				ss.Ratio = float32(scale)
				ss.Line = line
				var lastwidth float32 = 0
				var opacity float32 = 1
				var colour float32 = 0
				var points []PointRender
				for j := 0; j < len(line.Points); j++ {
					ss.Point = line.Points[j]
					ss.LastWidth = lastwidth
					if j%ss.Length == 0 {
						lastwidth = ss.GetWidth()
						opacity = ss.GetOpacity()
						colour = ss.GetColour()
					}

					points = append(points, PointRender{
						X:       float64(ss.Point.X) * float64(ss.Ratio),
						Y:       c.Height() - float64(ss.Point.Y)*float64(ss.Ratio),
						Width:   float64(lastwidth),
						Opacity: float64(opacity),
						Colour:  float64(colour),
						Render:  ss.Render,
					})

					ss.LastWidth = lastwidth
				}

				for j := PathSkip; j < len(points); j++ {
					// Did tried Bezier: useless.
					//Draw Path
					path := draw.NewPath()
					path = path.AppendPoint(draw.NewPoint(points[j-PathSkip].X, points[j-PathSkip].Y))
					path = path.AppendPoint(draw.NewPoint(points[j].X, points[j].Y))

					// Style

					// Set colour
					// TODO: Cool features: set colour
					// -string in layer
					// -type of pen

					contentCreator.Add_RG(points[j-1].Colour, points[j-1].Colour, points[j-1].Colour)

					// Set Cap
					var Cap string = "Round cap."
					if points[j-1].Render == HighlighterRender {
						Cap = "Butt cap" // Projecting square cap
					}
					contentCreator.Add_J(Cap)

					// Set join
					var Join string = "Round  join"
					if points[j-1].Render == HighlighterRender {
						Cap = "Miter join" // Bevel  join
					}
					contentCreator.Add_j(Join)

					//Set Tranparency
					if points[j-1].Render == HighlighterRender {
						contentCreator.Add_gs(GShighlighter)
					} else if points[j-1].Render == TiltPencilRender {
						GSname = core.PdfObjectName(fmt.Sprintf("GS%d", GScount))
						for page.HasExtGState(GSname) {
							GScount++
							GSname = core.PdfObjectName(fmt.Sprintf("GS%d", GScount))
						}
						GS = core.MakeDict()
						GS.Set("CA", core.MakeFloat(points[j-1].Opacity))
						page.AddExtGState(GSname, GS)
						contentCreator.Add_gs(GSname)
					} else {
						contentCreator.Add_gs(GSnormal)
					}

					// Set width
					contentCreator.Add_w(points[j-1].Width)

					draw.DrawPathWithCreator(path, contentCreator)

					// Be able to set a new style for next line
					if points[j-1].Render != HighlighterRender && j%ss.Length == 0 {
						contentCreator.Add_S()
					}

					// annotator Version: have some serious bug with width > 0.1
					// Should be the right way to go, but waiting that unipdf fix the bugs

					// lineDef := annotator.LineAnnotationDef{X1: points[i-PathSkip].X,
					//	Y1:        points[i-PathSkip].Y,
					//	X2:        points[i].X,
					//	Y2:        points[i].Y,
					//	LineColor: pdf.NewPdfColorDeviceRGB(points[i].Colour, points[i].Colour, points[i].Colour),
					//	Opacity:   points[i].Opacity,
					//	LineWidth: points[i].Width,
					// }
					// ann, err := annotator.CreateLineAnnotation(lineDef)
					// if err != nil {
					// 	return err
					// }
					// page.AddAnnotation(ann)

				}

				contentCreator.Add_S()
			}
		}

		contentCreator.Add_Q()
		drawingOperations := contentCreator.Operations().String()
		pageContentStreams, err := page.GetAllContentStreams()
		//hack: wrap the page content in a context to prevent transformation matrix misalignment
		wrapper := []string{"q", pageContentStreams, "Q", drawingOperations}
		page.SetContentStreams(wrapper, core.NewFlateEncoder())
	}

	return c.WriteToFile(p.outputFilePath)
}

func (p *PdfGenerator) initBackgroundPages(pdfArr []byte) error {
	if len(pdfArr) > 0 {
		pdfReader, err := pdf.NewPdfReader(bytes.NewReader(pdfArr))
		if err != nil {
			return err
		}

		encrypted, err := pdfReader.IsEncrypted()
		if err != nil {
			return nil
		}
		if encrypted {
			valid, err := pdfReader.Decrypt([]byte(""))
			if err != nil {
				return err
			}
			if !valid {
				return fmt.Errorf("cannot decrypt")
			}

		}

		p.pdfReader = pdfReader
		p.template = false
		return nil
	}

	p.template = true
	return nil
}

func (p *PdfGenerator) addBackgroundPage(c *creator.Creator, pageNum int) (*pdf.PdfPage, error) {
	var page *pdf.PdfPage

	if !p.template && !p.options.AnnotationsOnly {
		tmpPage, err := p.pdfReader.GetPage(pageNum)
		if err != nil {
			return nil, err
		}
		mbox, err := tmpPage.GetMediaBox()
		if err != nil {
			return nil, err
		}

		// TODO: adjust the page if cropped
		pageHeight := mbox.Ury - mbox.Lly
		pageWidth := mbox.Urx - mbox.Llx
		// use the pdf's page size
		c.SetPageSize(creator.PageSize{pageWidth, pageHeight})
		c.AddPage(tmpPage)
		page = tmpPage
	} else {
		page = c.NewPage()
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
