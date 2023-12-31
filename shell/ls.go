package shell

import (
	"errors"
	"sort"
	"strings"

	"github.com/abiosoft/ishell"
)

func lsCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "ls",
		Help:      "list directory",
		Completer: createEntryCompleter(ctx),
		Func: func(c *ishell.Context) {
			node := ctx.node

			if len(c.Args) == 1 {
				target := c.Args[0]

				argNode, err := ctx.api.Filetree().NodeByPath(target, ctx.node)

				if err != nil || node.IsFile() {
					c.Err(errors.New("directory doesn't exist"))
					return
				}

				node = argNode
			}

			children := node.Children

			// create an array of the keys of the children map
			keys := make([]string, 0, len(children))
			for k := range children {
				keys = append(keys, k)
			}

			// sort the keys by the name of the node case insensitively
			sort.SliceStable(keys, func(i, j int) bool {
				name1 := strings.ToLower(children[keys[i]].Name())
				name2 := strings.ToLower(children[keys[j]].Name())
				return name1 < name2
			})

			for _, key := range keys {
				e := children[key]

				eType := "d"
				if e.IsFile() {
					eType = "f"
				}
				c.Printf("[%s]\t%s\n", eType, e.Name())
			}
		},
	}
}
