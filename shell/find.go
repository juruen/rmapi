package shell

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/abiosoft/ishell"
	"github.com/juruen/rmapi/filetree"
	"github.com/juruen/rmapi/model"
	flag "github.com/ogier/pflag"
)

func findCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "find",
		Help:      "find files recursively, usage: find dir [regexp]",
		Completer: createDirCompleter(ctx),
		Func: func(c *ishell.Context) {
			flagSet := flag.NewFlagSet("ls", flag.ContinueOnError)
			var compact bool
			flagSet.BoolVarP(&compact, "compact", "c", false, "compact format")
			if err := flagSet.Parse(c.Args); err != nil {
				if err != flag.ErrHelp {
					c.Err(err)
				}
				return
			}
			argRest := flagSet.Args()
			var start, pattern string
			switch len(argRest) {
			case 2:
				pattern = argRest[1]
				fallthrough
			case 1:
				start = argRest[0]
			case 0:
				start = ctx.path
			default:
				c.Err(errors.New("missing arguments; usage find [dir] [regexp]"))
				return
			}

			startNode, err := ctx.api.Filetree().NodeByPath(start, ctx.node)

			if err != nil {
				c.Err(errors.New("start directory doesn't exist"))
				return
			}

			var matchRegexp *regexp.Regexp
			if pattern != "" {
				matchRegexp, err = regexp.Compile(pattern)
				if err != nil {
					c.Err(errors.New("failed to compile regexp"))
					return
				}
			}

			filetree.WalkTree(startNode, filetree.FileTreeVistor{
				Visit: func(node *model.Node, path []string) bool {
					entryName := formatEntry(compact, path, node)

					if matchRegexp == nil {
						c.Println(entryName)
						return false
					}

					if !matchRegexp.Match([]byte(entryName)) {
						return false
					}

					c.Println(entryName)

					return false
				},
			})
		},
	}
}
func formatEntry(compact bool, path []string, node *model.Node) string {
	fullpath := filepath.Join(strings.Join(path, "/"), node.Name())
	if compact {
		if node.IsDirectory() {
			return fullpath + "/"
		}

		return fullpath
	}
	var entryType string
	if node.IsDirectory() {
		entryType = "[d] "
	} else {
		entryType = "[f] "
	}
	return entryType + fullpath
}
