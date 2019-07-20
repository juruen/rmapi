package shell

import (
	"fmt"
	"os"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/peerdavid/rmapi/api"
	"github.com/peerdavid/rmapi/model"
	"github.com/peerdavid/rmapi/log"
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

	// Add geta and mgeta if all necessary tools are installed
	missingTools := verifyGetaCmdTools()
	hasAllToolsForGetaInstalled := len(missingTools) <= 0
	if(hasAllToolsForGetaInstalled){
		shell.AddCmd(getaCmd(ctx))
		shell.AddCmd(mgetaCmd(ctx))
	} else {
		log.Warning.Println(fmt.Sprintf("Commands geta and mgeta are disabled" + 
			" because the following tools are not installed: \n %v", 
			strings.Join(missingTools, "\n ")))
	}

	setCustomCompleter(shell)

	if len(os.Args) > 1 {
		return shell.Process(os.Args[1:]...)
	} else {
		shell.Run()

		return nil
	}
}
