package shell

import (
	"errors"
	"flag"
	"fmt"
	"github.com/abiosoft/ishell"
	"github.com/juruen/rmapi/annotations"
)

func getACmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "geta",
		Help:      "copy remote file to local and generate a PDF with its annotations",
		Completer: createEntryCompleter(ctx),
		Func: func(c *ishell.Context) {

			flagSet := flag.NewFlagSet("geta", flag.ContinueOnError)
			addPageNumbers := flagSet.Bool("p", false, "add page numbers")
			allPages := flagSet.Bool("a", false, "all pages")
			annotationsOnly := flagSet.Bool("n", false, "annotations only")
			if err := flagSet.Parse(c.Args); err != nil {
				if err != flag.ErrHelp {
					c.Err(err)
				}
				return
			}
			argRest := flagSet.Args()
			if len(argRest) == 0 {
				c.Err(errors.New("missing source file"))
				return
			}

			srcName := argRest[0]

			node, err := ctx.api.Filetree.NodeByPath(srcName, ctx.node)

			if err != nil || node.IsDirectory() {
				c.Err(errors.New("file doesn't exist"))
				return
			}

			c.Println(fmt.Sprintf("downloading: [%s]...", srcName))

			zipName := fmt.Sprintf("%s.zip", node.Name())
			err = ctx.api.FetchDocument(node.Document.ID, zipName)

			if err != nil {
				c.Err(errors.New(fmt.Sprintf("Failed to download file %s with %s", srcName, err.Error())))
				return
			}

			pdfName := fmt.Sprintf("%s-annotations.pdf", node.Name())
			options := annotations.PdfGeneratorOptions{AddPageNumbers: *addPageNumbers, AllPages: *allPages, AnnotationsOnly: *annotationsOnly}
			generator := annotations.CreatePdfGenerator(zipName, pdfName, options)
			err = generator.Generate()

			if err != nil {
				c.Err(errors.New(fmt.Sprintf("Failed to generate annotations for %s with %s", srcName, err.Error())))
				return
			}

			c.Printf("Annotations generated in: %s\n", pdfName)
		},
	}
}
