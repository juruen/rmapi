package shell

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

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
	pC.Println("\nUsage:\n")
	pC.Println("    [/]> put <file-list>")
	pC.Println("    [/]> put <file-list> -d <dst-dir>\n")
	pC.Println("<file-list> can be * (all files) or, 1 or more files separated by spaces\n")
}

func putCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "put",
		Help:      "copy local documents to cloud",
		Completer: createFsEntryCompleter(),
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("too few arguments for command 'put'\n")

				printUsage(c)
				return
			}

			// Total number of arguments.
			argsLen := len(c.Args)

			// Bool to flag if last argument is a dir (default -> true)
			userDir := false
			dIndex := -1

			// Check if -d option is present.
			for index, str := range c.Args {
				if str[0] == '-' {
					if str == "-d" {
						userDir = true
						dIndex = index
						break
					} else {
						c.Err(errors.New(fmt.Sprint("invalid option ", str, " for command 'put'")))

						printUsage(c)
						return
					}
				}
			}

			if dIndex == 0 {
				c.Err(errors.New("missing source files"))
				return
			}

			dstDir := ctx.node.Id()

			if userDir {
				// Number of arguments for option -d must be 1
				if argsLen > dIndex+2 {
					c.Err(errors.New("too many arguments for option -d"))
					return
				}

				node, err := ctx.api.Filetree.NodeByPath(c.Args[(argsLen-1)], ctx.node)

				if err != nil || node.IsFile() {
					c.Err(errors.New("directory doesn't exist"))
					return
				}

				dstDir = node.Id()
			}

			// Last index of doc list.
			listMaxIndex := argsLen - 1
			if dIndex != -1 {
				listMaxIndex = dIndex - 1
			}

			// is * (all files) ?.
			if listMaxIndex == 0 && c.Args[0] == "*" {
				// Upload all files
				fileList, err := ioutil.ReadDir("./")

				if err != nil {
					c.Err(errors.New("could not read current directory"))
					return
				}

				for _, f := range fileList {
					fName := f.Name()
					if strings.Contains(fName, ".pdf") || strings.Contains(fName, ".epub") {
						// If the file has a '.pdf' or a '.epub' extension, then upload
						putDoc(c, ctx, fName, dstDir)
					}
				}

			} else {
				// Upload the listed files one by one.
				for i := 0; i <= listMaxIndex; i++ {
					putDoc(c, ctx, c.Args[i], dstDir)
				}
			}

		},
	}
}
