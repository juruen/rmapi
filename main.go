package main

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/juruen/rmapi/api"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/shell"
)

func init_log() {
	var trace io.Writer
	if os.Getenv("RMAPI_TRACE") == "1" {
		trace = os.Stdout
	} else {
		trace = ioutil.Discard
	}

	log.Init(trace, os.Stdout, os.Stdout, os.Stderr)
}

func main() {
	init_log()

	ctx := api.AuthHttpCtx()

	fileTree := ctx.DocumentsFileTree()

	if fileTree == nil {
		log.Error.Fatal("failed to build documents tree")
	}

	shell.RunShell(ctx, fileTree)
}
