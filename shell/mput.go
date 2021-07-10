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

			treeFormatStr := "├"

			// Back up current remote location.
			currCtxPath := ctx.path
			currCtxNode := ctx.node
			// Change to requested directory.
			ctx.path = path
			ctx.node = node

			c.Println()
			putFilesAndDirs(ctx, c, "./", 0, &treeFormatStr)
			c.Println()

			// Reset.
			ctx.path = currCtxPath
			ctx.node = currCtxNode
		},
	}
}

// Print the required spaces and characters for tree formatting.
//
// Input -> [*ishell.Context]
//			[int]				tree depth (0 ... N-1)
//			[int]				Current item index in directory
//			[int]				Current directory list length
//			[*string]			Book keeping for tree formatting
func treeFormat(pC *ishell.Context, num int, lIndex int, lSize int, tFS *string) {

	tFStr := ""

	for i := 0; i <= num; i++ {
		if i == num {
			if lIndex == lSize-1 {
				tFStr += "└"
				pC.Printf("└── ") // Last item in current directory.
			} else if lSize > 1 {
				tFStr += "├"
				pC.Printf("├── ")
			}
		} else {
			prevStr := string([]rune(*tFS)[i])
			if prevStr == "│" || prevStr == "├" {
				tFStr += "│"
				pC.Printf("│")
			} else {
				tFStr += " "
				pC.Printf(" ")
			}

			pC.Printf("   ")
		}
	}

	*tFS = tFStr
}

func putFilesAndDirs(pCtx *ShellCtxt, pC *ishell.Context, localDir string, depth int, tFS *string) bool {

	if depth == 0 {
		pC.Println(pCtx.path)
	}

	os.Chdir(localDir) // Change to the local source directory.

	wd, _ := os.Getwd()
	dirList, err := ioutil.ReadDir(wd)

	if err != nil {
		pC.Err(fmt.Errorf("could not read the directory: %s", wd))
		return false
	}

	lSize := len(dirList)
	for index, d := range dirList {

		name := d.Name()

		if !pCtx.useHiddenFiles && strings.HasPrefix(d.Name(), ".") {
			continue
		}

		switch mode := d.Mode(); {
		case mode.IsDir():

			// Is a directory. Create directory and make a recursive call.
			_, err := pCtx.api.Filetree.NodeByPath(name, pCtx.node)

			if err != nil {
				// Directory does not exist. Create directory.
				treeFormat(pC, depth, index, lSize, tFS)
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
				treeFormat(pC, depth, index, lSize, tFS)
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

			putFilesAndDirs(pCtx, pC, name, depth+1, tFS)

			// Reset.
			pCtx.path = currCtxPath
			pCtx.node = currCtxNode

		case mode.IsRegular():

			docName, ext := util.DocPathToName(name)

			if !util.IsFileTypeSupported(ext) {
				continue
			}

			_, err := pCtx.api.Filetree.NodeByPath(docName, pCtx.node)

			if err == nil {
				// Document already exists.
				treeFormat(pC, depth, index, lSize, tFS)
				pC.Printf("document [%s] already exists\n", name)
			} else {
				// Document does not exist.
				treeFormat(pC, depth, index, lSize, tFS)
				pC.Printf("uploading: [%s]...", name)
			}
			doc, err := pCtx.api.UploadDocument(pCtx.node, name)

			if err != nil {
				pC.Err(fmt.Errorf("failed to upload file %s", name))
			} else {
				// Document uploaded successfully.
				pC.Println(" complete")
				pCtx.api.Filetree.AddDocument(*doc)
			}

		}
	}

	if localDir != "./" {
		os.Chdir("..")
	}

	return true
}
