package shell

import (
	"errors"
	"fmt"
	"time"
	"os/exec"
	"strings"
	"archive/zip"
	"os"
	"io"
	"github.com/abiosoft/ishell"
	"path/filepath"
	"github.com/juruen/rmapi/model"
)


func getaCmd(ctx *ShellCtxt) *ishell.Cmd {
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

			// Download document as zip
			node, err := ctx.api.Filetree.NodeByPath(srcName, ctx.node)
			if err != nil || node.IsDirectory() {
				c.Err(errors.New("file doesn't exist"))
				return
			}

			err = getAnnotatedDocument(ctx, node, "")
			if err != nil {
				c.Err(err)
				return
			}

			
			c.Println("Ok")
		},
	}
}


func getAnnotatedDocument(ctx *ShellCtxt, node *model.Node, path string) error {
	zipFile := fmt.Sprintf(".%s.zip", node.Name())
	err := ctx.api.FetchDocument(node.Document.ID, zipFile)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to download file %s with %s", node.Name(), err.Error()))
	}

	// Set output name and output file name
	output := node.Name()
	if(path != ""){
		output = fmt.Sprintf("%s/%s", path, node.Name())
	}
	outputFileName := fmt.Sprintf("%s.pdf", output)

	// Parse last modified time of document on api
	modifiedClientTime, err := time.Parse(time.RFC3339Nano, node.Document.ModifiedClient)
	if err != nil {
		// If we could not parse the time correctly, we still continue 
		// with the execution such that the pdf is downloaded...
		fmt.Println(err)
		fmt.Println("(Warning) Could not parse modified time. Overwrite existing file.")
		modifiedClientTime = time.Now().Local()
	}

	// If document has not changed since last update skip pdf convertion
	outputFile, err := os.Stat(outputFileName)
	if !os.IsNotExist(err) {
		outputFileModTime := outputFile.ModTime()
		if(outputFileModTime.Equal(modifiedClientTime)){
			fmt.Print("File has not changed since last download. ")
			os.Remove(zipFile)
			return nil
		}
	}

	// Unzip document
	tmpFolder := fmt.Sprintf(".%s", node.Document.ID)
	_, err = unzip(zipFile, tmpFolder)
	if err != nil {
		os.Remove(zipFile)
		return err
	}

	// Convert to pdf
	exportPdf := os.Getenv("GOPATH") + "/src/github.com/juruen/rmapi/tools/exportAnnotatedPdf"
	rM2svg := os.Getenv("GOPATH") + "/src/github.com/juruen/rmapi/tools/rM2svg"
	_, err = exec.Command(
		"/bin/bash", 
		exportPdf, 
		tmpFolder,
		node.Document.ID, 
		output, 
		rM2svg).CombinedOutput()

	if err != nil {
	   fmt.Println(err)
	}

	if err != nil {
		os.Remove(zipFile)
		os.RemoveAll(tmpFolder)
		return err
	}

	// Set creation time
	err = os.Chtimes(outputFileName, modifiedClientTime, modifiedClientTime)
	if err != nil {
		fmt.Println("(Warning) Could not set modified time of pdf file.")
	}

	// Cleanup
	os.Remove(zipFile)
	os.RemoveAll(tmpFolder)
	return nil
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
