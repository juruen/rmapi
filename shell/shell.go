package shell

import (
	"fmt"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/juruen/rmapi/api"
)

type ShellCtxt struct {
	node *api.Node
	path []string
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
			c.Println(ctx.pathString())
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

			if target == ".." {
				if ctx.node.Parent != nil {
					ctx.node = ctx.node.Parent
					ctx.path = ctx.path[:len(ctx.path)-1]
					c.Println("")
					c.SetPrompt(ctx.prompt())
					return
				}
			}

			for _, e := range ctx.node.Children {
				if e.Name() == target && e.IsDirectory() {
					ctx.node = e
					ctx.path = append(ctx.path, e.Name())
					c.Println("")
					c.SetPrompt(ctx.prompt())
					return
				}
			}

			c.Println("dir doesn't exist")
		},
	}
}

func getCmd(httpCtx *api.HttpClientCtx, ctx *ShellCtxt) *ishell.Cmd {
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

			err := httpCtx.FetchDocument(sourceNode.Document.ID, fmt.Sprintf("%s.zip", srcName))

			if err == nil {
				c.Println("OK")
				return
			}

			c.Println("Failed to downlaod file: %s", err.Error())
		},
	}
}

func (ctx *ShellCtxt) pathString() string {
	return fmt.Sprintf("/%s", strings.Join(ctx.path, "/"))
}

func (ctx *ShellCtxt) prompt() string {
	return fmt.Sprintf("[%s]>", ctx.pathString())
}

func RunShell(httpCtx *api.HttpClientCtx, fileTreeCtx *api.FileTreeCtx) {
	shell := ishell.New()
	ctx := &ShellCtxt{fileTreeCtx.Root(), make([]string, 0)}

	shell.Println("ReMarkable Cloud API Shell")
	shell.SetPrompt(ctx.prompt())

	shell.AddCmd(lsCmd(ctx))
	shell.AddCmd(pwdCmd(ctx))
	shell.AddCmd(cdCmd(ctx))
	shell.AddCmd(getCmd(httpCtx, ctx))

	shell.Run()
}
