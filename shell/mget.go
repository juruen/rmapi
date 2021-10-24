package shell

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
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
			outputDir := flagSet.String("o", ".", "output folder")
			removeDeleted := flagSet.Bool("d", false, "remove deleted/moved")

			if err := flagSet.Parse(c.Args); err != nil {
				if err != flag.ErrHelp {
					c.Err(err)
				}
				return
			}

			target := path.Clean(*outputDir)
			if *removeDeleted && target == "." {
				c.Err(fmt.Errorf("set a folder explictly with the -o flag when removing deleted (and not .)"))
				return
			}

			argRest := flagSet.Args()
			if len(argRest) == 0 {
				c.Err(errors.New(("missing source dir")))
				return
			}
			srcName := argRest[0]

			node, err := ctx.api.Filetree().NodeByPath(srcName, ctx.node)

			if err != nil || node.IsFile() {
				c.Err(errors.New("directory doesn't exist"))
				return
			}

			fileMap := make(map[string]struct{})
			fileMap[target] = struct{}{}

			visitor := filetree.FileTreeVistor{
				func(currentNode *model.Node, currentPath []string) bool {
					idxDir := 0
					if srcName == "." && len(currentPath) > 0 {
						idxDir = 1
					}

					fileName := currentNode.Name() + ".zip"

					dst := path.Join(target, filetree.BuildPath(currentPath[idxDir:], fileName))
					fileMap[dst] = struct{}{}

					dir := path.Dir(dst)
					fileMap[dir] = struct{}{}

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

			if *removeDeleted {
				filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						c.Err(fmt.Errorf("can't read %s %v", path, err))
						return nil
					}
					//just to be sure
					if path == target {
						return nil
					}
					if _, ok := fileMap[path]; !ok {
						var err error
						if info.IsDir() {
							c.Println("Removing folder ", path)
							err = os.RemoveAll(path)
							if err != nil {
								c.Err(err)
							}
							return filepath.SkipDir
						}

						c.Println("Removing ", path)
						err = os.Remove(path)
						if err != nil {
							c.Err(err)
						}
					}
					return nil
				})
			}
		},
	}
}
