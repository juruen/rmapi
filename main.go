package main

import (
	"io/ioutil"
	"os"
)

func main() {
	Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	ctx := authHttpCtx()

	fileTree := ctx.documentsFileTree()

	if fileTree == nil {
		Error.Fatal("failed to build documents tree")
	}

	runShell(fileTree)
}
