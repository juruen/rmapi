package shell

import (
	"fmt"

	"github.com/abiosoft/ishell"
	"github.com/juruen/rmapi/util"
)

func putCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "put",
		Help: "copy a local document to cloud",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("missing source file")
				return
			}

			srcName := c.Args[0]

			docName := util.PdfPathToName(srcName)

			_, err := ctx.fileTree.NodeByPath(docName, ctx.node)
			if err == nil {
				c.Println("entry already exists")
				return
			}

			c.Println(fmt.Sprintf("uploading: [%s]...", srcName))

			document, err := ctx.httpCtx.UploadDocument(ctx.node.Id(), srcName)

			if err != nil {
				c.Println("Failed to upload file: %s", err.Error())
				return
			}

			c.Println("OK")

			ctx.fileTree.AddDocument(*document)
		},
	}
}
