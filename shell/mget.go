package shell

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/abiosoft/ishell"
	"github.com/juruen/rmapi/filetree"
	"github.com/juruen/rmapi/model"
)

func mgetCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "mget",
		Help:      "recursively copy remote directory to local",
		Completer: createDirCompleter(ctx),
		Func: func(c *ishell.Context) {
			flagSet := flag.NewFlagSet("mget", flag.ContinueOnError)
			incremental := flagSet.Bool("i", false, "incremental")
			outputDir := flagSet.String("o", "./", "output folder")
			_ = flagSet.Bool("d", false, "remove deleted/moved")

			if err := flagSet.Parse(c.Args); err != nil {
				if err != flag.ErrHelp {
					c.Err(err)
				}
				return
			}
			argRest := flagSet.Args()

			if len(argRest) == 0 {
				c.Err(errors.New(("missing source dir")))
				return
			}

			srcName := argRest[0]

			node, err := ctx.api.Filetree.NodeByPath(srcName, ctx.node)

			if err != nil || node.IsFile() {
				c.Err(errors.New("directory doesn't exist"))
				return
			}

			visitor := filetree.FileTreeVistor{
				func(currentNode *model.Node, currentPath []string) bool {
					idxDir := 0
					if srcName == "." && len(currentPath) > 0 {
						idxDir = 1
					}

					fileName := currentNode.Name() + ".zip"

					dst := path.Join(*outputDir, filetree.BuildPath(currentPath[idxDir:], fileName))
					dir := path.Dir(dst)

					os.MkdirAll(dir, 0766)

					if currentNode.IsDirectory() {
						return filetree.ContinueVisiting
					}

					lastModified, err := currentNode.LastModified()
					if err != nil {
						fmt.Printf("%v for %s\n", err, dst)
						lastModified = time.Now()
					}

					if *incremental {
						stat, err := os.Stat(dst)
						if err == nil {
							localMod := stat.ModTime()

							if !lastModified.After(localMod) {
								return filetree.ContinueVisiting
							}
						}
					}

					c.Printf("downloading [%s]...", dst)

					err = ctx.api.FetchDocument(currentNode.Document.ID, dst)

					if err == nil {
						c.Println(" OK")

						err = os.Chtimes(dst, lastModified, lastModified)
						if err != nil {
							c.Err(fmt.Errorf("cant set lastModified for %s", dst))
						}
						return filetree.ContinueVisiting
					}

					c.Err(fmt.Errorf("Failed to download file %s", currentNode.Name()))

					return filetree.ContinueVisiting
				},
			}

			filetree.WalkTree(node, visitor)
		},
	}
}
