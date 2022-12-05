package shell

import (
	"sort"
	"strings"
	"time"

	"github.com/abiosoft/ishell"
	"github.com/juruen/rmapi/model"
	flag "github.com/ogier/pflag"
)

func sortNodes(in []*model.Node, options LsOptions) []*model.Node {
	sort.SliceStable(in, func(i, j int) bool {
		if options.DirFirst {
			if in[i].IsDirectory() && in[j].IsFile() {
				return true
			}
			if in[j].IsDirectory() && in[i].IsFile() {
				return false
			}
		}

		var left string
		var right string
		if options.ByTime {
			left = in[i].Document.ModifiedClient
			right = in[j].Document.ModifiedClient
		} else {
			left = strings.ToLower(in[i].Name())
			right = strings.ToLower(in[j].Name())
		}

		if options.Reverse {
			return left > right
		}
		return left < right
	})
	return in
}

func displayNode(c *ishell.Context, e *model.Node, d LsOptions) {
	if !d.Compact {
		eType := "d"
		if e.IsFile() {
			eType = "f"
		}
		c.Printf("[%s]\t%s\n", eType, e.Name())
		return
	}

	isFolder := ""
	if e.IsDirectory() {
		isFolder = "/"
	}
	if d.Long {
		t, _ := e.LastModified()
		c.Printf("%s %s%s\n", t.Local().Format(time.RFC3339), e.Name(), isFolder)
		return
	}
	c.Printf("%s%s\n", e.Name(), isFolder)
}

type LsOptions struct {
	Long     bool
	Compact  bool
	Reverse  bool
	DirFirst bool
	ByTime   bool
}

func lsCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "ls",
		Help:      "list directory",
		Completer: createEntryCompleter(ctx),
		Func: func(c *ishell.Context) {
			flagSet := flag.NewFlagSet("ls", flag.ContinueOnError)
			d := LsOptions{}
			flagSet.BoolVarP(&d.Compact, "compact", "c", false, "compact format")
			flagSet.BoolVarP(&d.Long, "long", "l", false, "long format")
			flagSet.BoolVarP(&d.Reverse, "reverse", "r", false, "reverse sort")
			flagSet.BoolVarP(&d.DirFirst, "group-directories", "d", false, "group directories")
			flagSet.BoolVarP(&d.ByTime, "time", "t", false, "sort by time")
			if err := flagSet.Parse(c.Args); err != nil {
				if err != flag.ErrHelp {
					c.Err(err)
				}
				return
			}
			argRest := flagSet.Args()

			var nodes []*model.Node
			if len(argRest) < 1 {
				nodes = ctx.node.Nodes()
			} else {
				var err error
				target := argRest[0]
				nodes, err = ctx.api.Filetree().NodesByPath(target, ctx.node, true)

				if err != nil {
					c.Err(err)
					return
				}
			}

			sorted := sortNodes(nodes, d)

			for _, e := range sorted {
				displayNode(c, e, d)
			}
		},
	}
}
