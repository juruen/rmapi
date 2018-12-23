package shell

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/abiosoft/ishell"
	"github.com/juruen/rmapi/filetree"
	"github.com/juruen/rmapi/model"
)

func mgetaCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name:      "mgeta",
		Help:      "recursively copy remote directory to local with annotations",
		Completer: createDirCompleter(ctx),
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Err(errors.New(("missing source dir")))
				return
			}

			srcName := c.Args[0]

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
					
					dst := "./" + filetree.BuildPath(currentPath[idxDir:], currentNode.Name())
					dir := path.Dir(dst)
					os.MkdirAll(dir, 0766)

					if currentNode.IsDirectory() {
						return filetree.ContinueVisiting
					}

					c.Printf("downloading [%s]...", dst)
					
					err = getAnnotatedDocument(ctx, currentNode, fmt.Sprintf("%s", dir))

					if err == nil {
						c.Println(" OK")
						return filetree.ContinueVisiting
					}

					c.Err(errors.New(fmt.Sprintf("Failed to downlaod file %s", currentNode.Name())))

					return filetree.ContinueVisiting
				},
			}

			filetree.WalkTree(node, visitor)
		},
	}
}
