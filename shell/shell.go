package shell

import (
	"fmt"
	"os"

	"github.com/abiosoft/ishell"
	"github.com/juruen/rmapi/api"
)

type ShellCtxt struct {
	node     *api.Node
	fileTree *api.FileTreeCtx
	httpCtx  *api.HttpClientCtx
	path     string
}

func (ctx *ShellCtxt) prompt() string {
	return fmt.Sprintf("[%s]>", ctx.path)
}

func RunShell(httpCtx *api.HttpClientCtx, fileTreeCtx *api.FileTreeCtx) {
	shell := ishell.New()
	ctx := &ShellCtxt{fileTreeCtx.Root(), fileTreeCtx, httpCtx, "/"}

	shell.Println("ReMarkable Cloud API Shell")
	shell.SetPrompt(ctx.prompt())

	shell.AddCmd(lsCmd(ctx))
	shell.AddCmd(pwdCmd(ctx))
	shell.AddCmd(cdCmd(ctx))
	shell.AddCmd(getCmd(ctx))
	shell.AddCmd(mgetCmd(ctx))
	shell.AddCmd(mkdirCmd(ctx))
	shell.AddCmd(rmCmd(ctx))
	shell.AddCmd(mvCmd(ctx))

	if len(os.Args) > 1 {
		shell.Process(os.Args[1:]...)
	} else {
		shell.Run()
	}
}
