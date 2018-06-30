package shell

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/juruen/rmapi/util"
)

func mputCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "mput",
		Help:      "recursively copy local files to remote directory",
		Completer: createFsEntryCompleter(),
		Func: func(c *ishell.Context) {

			argsLen := len(c.Args)

			if argsLen == 0 {
				c.Err(errors.New(("missing destination dir")))
				return
			}

			if argsLen > 1 {
				c.Err(errors.New(("too many arguments for command mput")))
				return
			}

			// Past this point, the number of arguments is 1.

			node, err := ctx.api.Filetree.NodeByPath(c.Args[0], ctx.node)

			if err != nil || node.IsFile() {
				c.Err(errors.New("directory doesn't exist"))
				return
			}

			path, err := ctx.api.Filetree.NodeToPath(node)

			if err != nil || node.IsFile() {
				c.Err(errors.New("directory doesn't exist"))
				return
			}

			// dstDir := node.Id()

			// back up
			currCtxPath := ctx.path
			currCtxNode := ctx.node

			// Change to requested directory.
			ctx.path = path
			ctx.node = node

			putFilesAndDirs(ctx, c, "./")

			// Reset.
			ctx.path = currCtxPath
			ctx.node = currCtxNode
		},
	}
}

// Checks whether the file has a pdf or epub extension.
// Input -> Valid file name.
// Returns -> true if the file is a pdf or epub
//		   -> false otherwise
func checkFileType(fName string) bool {
	return (strings.Contains(fName, ".pdf") ||
		strings.Contains(fName, ".epub"))
}

func putFilesAndDirs(pCtx *ShellCtxt, pC *ishell.Context, localDir string) bool {

	os.Chdir(localDir) // Change to the local source directory

	wd, _ := os.Getwd()
	dirList, err := ioutil.ReadDir(wd)

	if err != nil {
		pC.Err(fmt.Errorf("could not read the directory: ", wd))
		return false
	}

	// Directory has been read.

	for _, d := range dirList {
		name := d.Name()

		switch mode := d.Mode(); {
		case mode.IsDir():

			// Is a directory. Create directory and make a recursive call.
			_, err := pCtx.api.Filetree.NodeByPath(name, pCtx.node)

			if err != nil {
				// Directory does not exist. Create directory.
				pC.Printf("creating directory [%s] ...", name)
				doc, err := pCtx.api.CreateDir(pCtx.node.Id(), name)

				if err != nil {
					pC.Err(errors.New(fmt.Sprint("failed to create directory", err)))
					continue
				} else {
					pC.Println(" complete")
					pCtx.api.Filetree.AddDocument(doc) // Add dir to file tree.
				}
			} else {
				// Directory already exists.
				pC.Printf("directory [%s] already exists\n", name)
			}

		case mode.IsRegular():

			// Is a file.
			if checkFileType(name) {
				// Is a pdf or epub file

				docName := util.DocPathToName(name)
				_, err := pCtx.api.Filetree.NodeByPath(docName, pCtx.node)

				if err == nil {
					// Document already exists.

					pC.Printf("document [%s] already exists\n", name)
				} else {
					// Document does not exist.

					pC.Printf("uploading: [%s]...", name)
					doc, err := pCtx.api.UploadDocument(pCtx.node.Id(), name)

					if err != nil {
						pC.Err(fmt.Errorf("Failed to upload file ", name))
					} else {
						// Document uploaded successfully.
						pC.Println(" complete")
						pCtx.api.Filetree.AddDocument(*doc)
					}
				}
			}
		}
	}

	return true
}
