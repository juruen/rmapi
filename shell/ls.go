package shell

import 	"github.com/abiosoft/ishell"

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
