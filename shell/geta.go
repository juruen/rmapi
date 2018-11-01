package shell

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"archive/zip"
	"os"
	"io"
	"github.com/abiosoft/ishell"
	"path/filepath"
)


func getCmdA(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "geta",
		Help:      "copy remote file annotated to local",
		Completer: createEntryCompleter(ctx),
		Func: func(c *ishell.Context) {
			// Parse cmd args
			if len(c.Args) == 0 {
				c.Err(errors.New("missing source file"))
				return
			}
			srcName := c.Args[0]

			// Download document as zip
			node, err := ctx.api.Filetree.NodeByPath(srcName, ctx.node)
			if err != nil || node.IsDirectory() {
				c.Err(errors.New("file doesn't exist"))
				return
			}

			c.Println(fmt.Sprintf("downlading: [%s]...", srcName))

			zipFile := fmt.Sprintf("%s.zip", node.Name())
			err = ctx.api.FetchDocument(node.Document.ID, zipFile)
			if err != nil {
				c.Err(errors.New(fmt.Sprintf("Failed to download file %s with %s", srcName, err.Error())))
				return
			}
			
			// Unzip document
			tmpFolder := node.Document.ID
			_, err = unzip(zipFile, tmpFolder)
			if err != nil {
				c.Err(err)
				os.Remove(zipFile)
				return
			}

			// Convert lines file
			linesFile := fmt.Sprintf("%s/%s.lines", tmpFolder, node.Document.ID)
			svgFiles := fmt.Sprintf("%s/%s", node.Name(), node.Name())
			os.MkdirAll(node.Name(), 0755)
			rM2svg := os.Getenv("GOPATH") + "/src/github.com/peerdavid/rmapi/tools/rM2svg"
			_, err = exec.Command(rM2svg, "-i", linesFile, "-o", svgFiles).CombinedOutput()
			if err != nil {
				c.Err(err)
			}

			// Cleanup
			os.Remove(zipFile)
			os.RemoveAll(tmpFolder)
			c.Println("OK")
		},
	}
}


// From https://golangcode.com/unzip-files-in-go/
func unzip(src string, dest string) ([]string, error) {

    var filenames []string

    r, err := zip.OpenReader(src)
    if err != nil {
        return filenames, err
    }
    defer r.Close()

    for _, f := range r.File {

        rc, err := f.Open()
        if err != nil {
            return filenames, err
        }
        defer rc.Close()

        // Store filename/path for returning and using later on
        fpath := filepath.Join(dest, f.Name)

        // Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
        if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
            return filenames, fmt.Errorf("%s: illegal file path", fpath)
        }

        filenames = append(filenames, fpath)

        if f.FileInfo().IsDir() {

            // Make Folder
            os.MkdirAll(fpath, os.ModePerm)

        } else {

            // Make File
            if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
                return filenames, err
            }

            outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
            if err != nil {
                return filenames, err
            }

            _, err = io.Copy(outFile, rc)

            // Close the file without defer to close before next iteration of loop
            outFile.Close()

            if err != nil {
                return filenames, err
            }

        }
    }
    return filenames, nil
}
