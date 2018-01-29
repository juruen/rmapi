package main

import (
	"fmt"
	"strings"

	"github.com/abiosoft/ishell"
)

type ShellCtxt struct {
	node *Node
	path []string
}

func lsCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "ls",
		Help: "list current directory",
		Func: func(c *ishell.Context) {
			for _, e := range ctx.node.Children {
				eType := "d"
				if e.isFile() {
					eType = "f"
				}
				c.Printf("[%s]\t%s\n", eType, e.name())
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
				if e.name() == target && e.isDirectory() {
					ctx.node = e
					ctx.path = append(ctx.path, e.name())
					c.Println("")
					c.SetPrompt(ctx.prompt())
					return
				}
			}

			c.Println("dir doesn't exist")
		},
	}
}

func (ctx *ShellCtxt) pathString() string {
	return fmt.Sprintf("/%s", strings.Join(ctx.path, "/"))
}

func (ctx *ShellCtxt) prompt() string {
	return fmt.Sprintf("[%s]>", ctx.pathString())
}

func runShell(fileTreeCtx *FileTreeCtx) {
	shell := ishell.New()
	ctx := &ShellCtxt{fileTreeCtx.root, make([]string, 0)}

	shell.Println("ReMarkable Cloud API Shell")
	shell.SetPrompt(ctx.prompt())

	shell.AddCmd(lsCmd(ctx))
	shell.AddCmd(pwdCmd(ctx))
	shell.AddCmd(cdCmd(ctx))

	shell.Run()
}
