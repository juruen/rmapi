package shell

import "github.com/abiosoft/ishell"

func pwdCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "pwd",
		Help: "print current directory",
		Func: func(c *ishell.Context) {
			c.Println(ctx.path)
		},
	}
}
