package main

import (
	"io/ioutil"
	"os"

	"github.com/juruen/rmapi/api"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/shell"
)

func main() {
	log.Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	ctx := api.AuthHttpCtx()

	fileTree := ctx.DocumentsFileTree()

	if fileTree == nil {
		log.Error.Fatal("failed to build documents tree")
	}

	shell.RunShell(ctx, fileTree)
}
