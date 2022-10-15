package main

import (
	"flag"
	"os"

	"github.com/juruen/rmapi/api"
	"github.com/juruen/rmapi/config"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/shell"
)

const AUTH_RETRIES = 3

func main() {
	ni := flag.Bool("ni", false, "not interactive")
	flag.Parse()
	otherFlags := flag.Args()

	if len(otherFlags) > 0 {
		switch otherFlags[0] {
		case "logout":
			configFile := config.ConfigPath()
			err := os.Remove(configFile)
			if err != nil {
				log.Error.Fatalln(err)
			}
			return
		}
	}

	var ctx api.ApiCtx
	var err error
	var userInfo *api.UserInfo

	for i := 0; i < AUTH_RETRIES; i++ {
		authCtx := api.AuthHttpCtx(i > 0, *ni)

		userInfo, err = api.ParseToken(authCtx.Tokens.UserToken)
		if err != nil {
			log.Trace.Println(err)
			continue
		}

		ctx, err = api.CreateApiCtx(authCtx, userInfo.SyncVersion)
		if err != nil {
			log.Trace.Println(err)
		} else {
			break
		}
	}

	if err != nil {
		log.Error.Fatal("failed to build documents tree, last error: ", err)
	}

	err = shell.RunShell(ctx, userInfo, otherFlags)

	if err != nil {
		log.Error.Println("Error: ", err)

		os.Exit(1)
	}
}
