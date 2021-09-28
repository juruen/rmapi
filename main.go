package main

import (
	"flag"
	"os"

	"github.com/juruen/rmapi/api"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/shell"
)

const AUTH_RETRIES = 3

func run_shell(ctx api.ApiCtx, args []string) {
	err := shell.RunShell(ctx, args)

	if err != nil {
		log.Error.Println("Error: ", err)

		os.Exit(1)
	}
}

func main() {
	// log.InitLog()
	ni := flag.Bool("ni", false, "not interactive")
	flag.Parse()
	rstArgs := flag.Args()

	var ctx api.ApiCtx
	var err error
	for i := 0; i < AUTH_RETRIES; i++ {
		ctx, err = api.CreateApiCtx(api.AuthHttpCtx(i > 0, *ni))

		if err != nil {
			log.Trace.Println(err)
		} else {
			break
		}
	}

	if err != nil {
		log.Error.Fatal("failed to build documents tree, last error: ", err)
	}

	run_shell(ctx, rstArgs)
}
