package shell

import (
	"os"
	"path"

	"github.com/abiosoft/ishell"
	"github.com/juruen/rmapi/api"
)

func mgetCmd(ctx *ShellCtxt) *ishell.Cmd {
	return &ishell.Cmd{
		Name: "mget",
		Help: "recursively copy remote directory to local",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("missing source dir")
				return
			}

			srcName := c.Args[0]

			node, err := ctx.fileTree.NodeByPath(srcName, ctx.node)

			if err != nil || node.IsFile() {
				c.Println("directory doesn't exist")
				return
			}

			visitor := api.FileTreeVistor{
				func(currentNode *api.Node, currentPath []string) bool {
					idxDir := 0
					if srcName == "." && len(currentPath) > 0 {
						idxDir = 1
					}

					dst := "./" + api.BuildPath(currentPath[idxDir:], currentNode.Name())
					dir := path.Dir(dst)

					os.MkdirAll(dir, 0766)

					if currentNode.IsDirectory() {
						return api.ContinueVisiting
					}

					c.Printf("downloading [%s]...", dst)

					err = ctx.httpCtx.FetchDocument(currentNode.Document.ID, dst)

					if err == nil {
						c.Println(" OK")
						return api.ContinueVisiting
					}

					c.Println("Failed to downlaod file: %s", err)

					return api.ContinueVisiting
				},
			}

			api.WalkTree(node, visitor)
		},
	}
}
