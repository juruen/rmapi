package shell

import (
	"errors"
	"fmt"
	"github.com/juruen/rmapi/annotations"
	"github.com/juruen/rmapi/util"
	"io/ioutil"
	"os"
	"path"

	"github.com/abiosoft/ishell"
)

func getACmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "geta",
		Help:      "copy remote file to local and generate a PDF with its annotations",
		Completer: createEntryCompleter(ctx),
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Err(errors.New("missing source file"))
				return
			}

			srcName := c.Args[0]

			node, err := ctx.api.Filetree.NodeByPath(srcName, ctx.node)

			if err != nil || node.IsDirectory() {
				c.Err(errors.New("file doesn't exist"))
				return
			}

			c.Println(fmt.Sprintf("downlading: [%s]...", srcName))

			zipName := fmt.Sprintf("%s.zip", node.Name())
			err = ctx.api.FetchDocument(node.Document.ID, zipName)

			if err != nil {
				c.Err(errors.New(fmt.Sprintf("Failed to download file %s with %s", srcName, err.Error())))
				return
			}

			tmp, err := ioutil.TempDir("", "rmapizip")
			if err != nil {
				c.Err(errors.New(fmt.Sprintf("Failed to create tmp dir  with %s", err.Error())))
				return
			}

			defer os.RemoveAll(tmp)

			err = util.Unzip(zipName, tmp)
			if err != nil {
				c.Err(errors.New(fmt.Sprintf("Failed to unzip %s with %s", srcName, err.Error())))
				return
			}

			unzipDir := path.Join(tmp, node.Id())
			pdfName := fmt.Sprintf("%s-annotations.pdf", node.Name())

			generator := annotations.CreatePdfGenerator(unzipDir, pdfName)
			err = generator.Generate()

			if err != nil {
				c.Err(errors.New(fmt.Sprintf("Failed to generate annotations for %s with %s", srcName, err.Error())))
				return
			}

			c.Printf("Annotations generated in: %s\n", pdfName)
		},
	}
}
