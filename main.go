package main

import (
	"os"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/juruen/rmapi/api"
	"github.com/juruen/rmapi/fusefs"
	"github.com/juruen/rmapi/log"
	"github.com/juruen/rmapi/shell"
)

const AUTH_RETRIES = 3

func mount_fuse(ctx *api.ApiCtx) {
	mountPoint := os.Args[2]

	root := fusefs.NewFuseFsRoot(ctx)
	conn := nodefs.NewFileSystemConnector(root, nil)
	server, err := fuse.NewServer(conn.RawFS(), mountPoint, &fuse.MountOptions{})
	if err != nil {
		log.Error.Printf("Mount fail: %v\n", err)
		os.Exit(1)
	}

	log.Info.Println("rM fs mounted on", mountPoint)

	server.Serve()
}

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

	log.Info.Println(os.Args)
	if len(os.Args) == 3 && os.Args[1] == "--fuse-mount" {
		mount_fuse(ctx)
	} else {
		run_shell(ctx)
	}

}
