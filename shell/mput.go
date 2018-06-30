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
				c.Err(errors.New("remote directory does not exist"))
				return
			}

			path, err := ctx.api.Filetree.NodeToPath(node)

			if err != nil || node.IsFile() {
				c.Err(errors.New("remote directory does not exist"))
				return
			}

			// Back up current remote location.
			currCtxPath := ctx.path
			currCtxNode := ctx.node
			// Change to requested directory.
			ctx.path = path
			ctx.node = node

			putFilesAndDirs(ctx, c, "./", 0)

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

// Print the required number of spaces
func treeFormat(pC *ishell.Context, num int) {

	if num != 0 {
		pC.Printf("│")
	} else {
		pC.Printf("├")
	}

	for i := 0; i < 2*num; i++ {
		pC.Printf(" ")
	}

	if num != 0 {
		pC.Printf("  └── ")
	} else {
		pC.Printf("── ")
	}
}

func putFilesAndDirs(pCtx *ShellCtxt, pC *ishell.Context, localDir string, depth int) bool {

	if localDir == "./" {
		pC.Println(".")
	}

	os.Chdir(localDir) // Change to the local source directory.

	wd, _ := os.Getwd()
	// pC.Println("DEBUG: changing to directory", wd)
	dirList, err := ioutil.ReadDir(wd)

	if err != nil {
		pC.Err(fmt.Errorf("could not read the directory: ", wd))
		return false
	}

	for _, d := range dirList {
		name := d.Name()

		switch mode := d.Mode(); {
		case mode.IsDir():

			// Is a directory. Create directory and make a recursive call.
			_, err := pCtx.api.Filetree.NodeByPath(name, pCtx.node)

			if err != nil {
				// Directory does not exist. Create directory.
				treeFormat(pC, depth)
				pC.Printf("creating directory [%s]...", name)
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
				treeFormat(pC, depth)
				pC.Printf("directory [%s] already exists\n", name)
			}

			// Error checking not required? Unless, someone deletes
			// or renames the directory meanwhile.

			node, _ := pCtx.api.Filetree.NodeByPath(name, pCtx.node)
			path, _ := pCtx.api.Filetree.NodeToPath(node)

			// Back up current remote location.
			currCtxPath := pCtx.path
			currCtxNode := pCtx.node

			pCtx.path = path
			pCtx.node = node

			putFilesAndDirs(pCtx, pC, name, depth+1)

			// Reset.
			pCtx.path = currCtxPath
			pCtx.node = currCtxNode

		case mode.IsRegular():

			// Is a file.
			if checkFileType(name) {
				// Is a pdf or epub file

				docName := util.DocPathToName(name)
				_, err := pCtx.api.Filetree.NodeByPath(docName, pCtx.node)

				if err == nil {
					// Document already exists.
					treeFormat(pC, depth)
					pC.Printf("document [%s] already exists\n", name)
				} else {
					// Document does not exist.
					treeFormat(pC, depth)
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

	if localDir != "./" {
		// pC.Println("DEBUG: exiting directory", wd)
		os.Chdir("..")
	}

	return true
}
