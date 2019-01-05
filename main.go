package main

import (
	"os"

	"github.com/juruen/rmapi/api"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/shell"
)

const AUTH_RETRIES = 3

func run_shell(ctx *api.ApiCtx) {
	err := shell.RunShell(ctx)

	if err != nil {
		log.Error.Println("Error: ", err)
		os.Exit(1)
	}
}

func main() {
	log.InitLog()

	var ctx *api.ApiCtx
	for i := 0; i < AUTH_RETRIES; i++ {
		ctx = api.CreateApiCtx(api.AuthHttpCtx())

		if ctx.Filetree == nil && i < AUTH_RETRIES {
			log.Error.Println("retrying...")
		}
	}

	if ctx.Filetree == nil {
		log.Error.Fatal("failed to build documents tree")
	}

	run_shell(ctx)
}
