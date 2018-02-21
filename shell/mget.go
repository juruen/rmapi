package shell

import (
	"os"
	"path"

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
			if len(c.Args) == 0 {
				c.Println("missing source dir")
				return
			}

			srcName := c.Args[0]

			node, err := ctx.api.Filetree.NodeByPath(srcName, ctx.node)

			if err != nil || node.IsFile() {
				c.Println("directory doesn't exist")
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

					err = ctx.api.FetchDocument(currentNode.Document.ID, dst)

					if err == nil {
						c.Println(" OK")
						return filetree.ContinueVisiting
					}

					c.Println("Failed to downlaod file: %s", err)

					return filetree.ContinueVisiting
				},
			}

			filetree.WalkTree(node, visitor)
		},
	}
}
