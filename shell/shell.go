package shell

import (
	"fmt"

	"github.com/abiosoft/ishell"
	"github.com/juruen/rmapi/api"
)

type ShellCtxt struct {
	node     *api.Node
	fileTree *api.FileTreeCtx
	httpCtx  *api.HttpClientCtx
	path     string
}

func lsCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "ls",
		Help: "list current directory",
		Func: func(c *ishell.Context) {
			for _, e := range ctx.node.Children {
				eType := "d"
				if e.IsFile() {
					eType = "f"
				}
				c.Printf("[%s]\t%s\n", eType, e.Name())
			}
		},
	}
}

func pwdCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "pwd",
		Help: "print current directory",
		Func: func(c *ishell.Context) {
			c.Println(ctx.path)
		},
	}
}

func cdCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "cd",
		Help: "change directory",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				return
			}

			target := c.Args[0]

			node, err := ctx.fileTree.NodeByPath(target, ctx.node)

			if err != nil || node.IsFile() {
				c.Println("directory doesn't exist")
				return
			}

			path, err := ctx.fileTree.NodeToPath(node)

			if err != nil || node.IsFile() {
				c.Println("directory doesn't exist")
				return
			}

			ctx.path = path
			ctx.node = node

			c.Println()
			c.SetPrompt(ctx.prompt())
		},
	}
}

func getCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "get",
		Help: "copy remote file to local",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("missing source file")
				return
			}

			srcName := c.Args[0]
			found := false
			var srcId string

			for _, e := range ctx.node.Children {
				if e.Name() == srcName && e.IsFile() {
					found = true
					srcId = e.Document.ID
					break
				}
			}

			sourceNode, _ := ctx.node.Children[srcId]
			if !found {
				c.Println("remote file doesn't exist")
				return
			}

			c.Println(fmt.Sprintf("downlading: [%s]...", srcName))

			err := ctx.httpCtx.FetchDocument(sourceNode.Document.ID, fmt.Sprintf("%s.zip", srcName))

			if err == nil {
				c.Println("OK")
				return
			}

			c.Println("Failed to downlaod file: %s", err.Error())
		},
	}
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

	shell.Run()
}
