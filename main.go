package main

import (
	"io/ioutil"
	"os"
)

func main() {
	Init(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	ctx := authHttpCtx()

	ctx.listDocuments()
}
