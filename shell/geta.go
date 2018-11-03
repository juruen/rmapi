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
	"github.com/peerdavid/rmapi/model"
)


func getCmdA(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "geta",
		Help:      "copy remote file to local with annotations. Optional [svg, pdf]",
		Completer: createEntryCompleter(ctx),
		Func: func(c *ishell.Context) {
			// Parse cmd args
			if len(c.Args) == 0 {
				c.Err(errors.New("missing source file"))
				return
			}
			srcName := c.Args[0]

			fileType := "pdf"
			if len(c.Args) == 2 {
				fileType = c.Args[1]
			}

			// Download document as zip
			node, err := ctx.api.Filetree.NodeByPath(srcName, ctx.node)
			if err != nil || node.IsDirectory() {
				c.Err(errors.New("file doesn't exist"))
				return
			}
			c.Println(fmt.Sprintf("downlading: [%s]...", srcName))

			zipFile := fmt.Sprintf(".%s.zip", node.Name())
			err = ctx.api.FetchDocument(node.Document.ID, zipFile)
			if err != nil {
				c.Err(errors.New(fmt.Sprintf("Failed to download file %s with %s", srcName, err.Error())))
				return
			}
			
			// Unzip document
			tmpFolder := fmt.Sprintf("%s", node.Document.ID)
			docFiles, err := unzip(zipFile, tmpFolder)
			if err != nil {
				c.Err(err)
				os.Remove(zipFile)
				return
			}
			
			// Convert to annotated pdf
			pdfFile := fmt.Sprintf("%s/%s.pdf", tmpFolder, node.Document.ID)
			if contains(docFiles, pdfFile){
				c.Println(fmt.Sprintf("creating annoated pdf: [%s]...", srcName))

				if(fileType == "svg"){
					c.Err(errors.New("svg export not supported for annotated pdf."))
					os.Remove(zipFile)
					os.RemoveAll(tmpFolder)
					return
				}

				// Convert lines file to svg foreground
				svgFiles := fmt.Sprintf("%s/foreground", node.Document.ID)
				err = linesToSvg(tmpFolder, node, svgFiles)
				if err != nil {
					c.Err(err)
					os.Remove(zipFile)
					os.RemoveAll(tmpFolder)
					return
				}

				// Convert to pdf
				c.Println(fmt.Sprintf("creating annotated pdf: [%s]...", srcName))
				exportPdf := os.Getenv("GOPATH") + "/src/github.com/peerdavid/rmapi/tools/exportAnnotatedPdf"
				_, err = exec.Command("/bin/sh", exportPdf, tmpFolder, node.Document.ID, node.Name()).CombinedOutput()
				if err != nil {
					c.Err(err)
					os.Remove(zipFile)
					os.RemoveAll(tmpFolder)
					return
				}

			} else {
				c.Println(fmt.Sprintf("creating notebook: [%s]...", srcName))

				svgTmpFolder := tmpFolder
				if(fileType == "svg"){
					svgTmpFolder = node.Name()
					os.MkdirAll(svgTmpFolder, 0755)
				}

				svgFiles := fmt.Sprintf("%s/%s", svgTmpFolder, node.Name())
				err = linesToSvg(tmpFolder, node, svgFiles)
				if err != nil {
					c.Err(err)
					os.Remove(zipFile)
					os.RemoveAll(tmpFolder)
					return
				}

				if(fileType == "pdf"){
					svgToPdf := os.Getenv("GOPATH") + "/src/github.com/peerdavid/rmapi/tools/svgToPdf"
					_, err = exec.Command("/bin/sh", svgToPdf, svgFiles, node.Name()).CombinedOutput()
					if err != nil {
						c.Err(err)
						os.Remove(zipFile)
						os.RemoveAll(tmpFolder)
						return
					}
				}
			}

			// Cleanup
			os.Remove(zipFile)
			os.RemoveAll(tmpFolder)
			c.Println("Ok")
		},
	}
}



func linesToSvg(tmpFolder string, node *model.Node, outName string) error{
	linesFile := fmt.Sprintf("%s/%s", tmpFolder, node.Document.ID)
	lines2Svg := os.Getenv("GOPATH") + "/src/github.com/peerdavid/rmapi/tools/lines2svg"
	_, err := exec.Command(lines2Svg, linesFile, outName).CombinedOutput()
	return err
}


func contains(a []string, e string) bool {
    for _, v := range a {
        if v == e {
            return true
        }
    }
    return false
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
