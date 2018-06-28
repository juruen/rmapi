package shell

import (
	"errors"
	"fmt"

	"github.com/abiosoft/ishell"
	"github.com/juruen/rmapi/util"
)

// Requests upload doc to cloud
func putDoc(pC *ishell.Context, ctx *ShellCtxt, pSrcName string, pDstDir string) int {

	docName := util.DocPathToName(pSrcName)

	_, err := ctx.api.Filetree.NodeByPath(docName, ctx.node)
	if err == nil {

		pC.Err(errors.New(fmt.Sprint(pSrcName, " already exists")))

		return -1
	}

	pC.Printf("uploading: [%s]...", pSrcName)

	document, err := ctx.api.UploadDocument(pDstDir, pSrcName)

	if err != nil {
		pC.Err(errors.New(fmt.Sprint("Failed to upload file ", pSrcName, err.Error())))
		return -1
	}

	pC.Println(" complete")

	ctx.api.Filetree.AddDocument(*document)

	return 1
}

func printUsage(pC *ishell.Context) {
	pC.Println("Usage:\n")
	pC.Println("    [/]> put  <file-list>")
	pC.Println("    [/]> put  <file-list> <dst-dir>\n")
	pC.Println("<file-list> can be * (all files) or, 1 or more files separated by spaces\n")
}

func putCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "put",
		Help:      "copy local documents to cloud",
		Completer: createFsEntryCompleter(),
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("missing source files\n")

				printUsage(c)
				return
			}

			// Total number of arguments.
			argsLen := len(c.Args)

			// Check if last argument is a directory.
			

			// Extract the destination directory data if available.
			dstDir := ctx.node.Id()

			if len(c.Args) > 1 {
				node, err := ctx.api.Filetree.NodeByPath(c.Args[(argsLen-1)], ctx.node)

				if err != nil || node.IsFile() {
					// Make one?
					c.Err(errors.New("directory doesn't exist"))
					return
				}

				dstDir = node.Id()
			}

			putDoc(c, ctx, c.Args[0], dstDir)

			// ioutil.

			// srcName := c.Args[0]

			// docName := util.DocPathToName(srcName)

			// c.Printf("uploading: [%s]...", srcName)

			// document, err := ctx.api.UploadDocument(dstDir, srcName)

			// if err != nil {
			// 	c.Err(errors.New(fmt.Sprint("Failed to upload file", srcName, err.Error())))
			// 	return
			// }

			// c.Println(" complete")

			// ctx.api.Filetree.AddDocument(*document)

			// End of function
		},
	}
}
