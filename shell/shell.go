package shell

import (
	"fmt"
	"os"

	"github.com/abiosoft/ishell"
	"github.com/juruen/rmapi/api"
	"github.com/juruen/rmapi/model"
)

type ShellCtxt struct {
	node *model.Node
	api  *api.ApiCtx
	path string
}

func (ctx *ShellCtxt) prompt() string {
	return fmt.Sprintf("[%s]>", ctx.path)
}

func setCustomCompleter(shell *ishell.Shell) {
	cmdCompleter := make(cmdToCompleter)
	for _, cmd := range shell.Cmds() {
		cmdCompleter[cmd.Name] = cmd.Completer
	}

	completer := shellPathCompleter{cmdCompleter}
	shell.CustomCompleter(completer)
}

func RunShell(apiCtx *api.ApiCtx) error {
	shell := ishell.New()
	ctx := &ShellCtxt{apiCtx.Filetree.Root(), apiCtx, apiCtx.Filetree.Root().Name()}

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
	shell.AddCmd(putCmd(ctx))
	shell.AddCmd(mputCmd(ctx))
	shell.AddCmd(versionCmd(ctx))
	shell.AddCmd(statCmd(ctx))

	setCustomCompleter(shell)

	if len(os.Args) > 1 {
		return shell.Process(os.Args[1:]...)
	} else {
		shell.Run()

		return nil
	}
}
