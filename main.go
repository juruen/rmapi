package main

import (
	"github.com/juruen/rmapi/api"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/shell"
)

func main() {
	log.InitLog()

	ctx := api.CreateApiCtx(api.AuthHttpCtx())

	if ctx.Filetree == nil {
		log.Error.Fatal("failed to build documents tree")
	}

	shell.RunShell(ctx)
}
