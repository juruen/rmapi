package shell

import (
	"fmt"
	"path"
	"strings"

	"github.com/peerdavid/rmapi/log"
	"github.com/peerdavid/rmapi/model"
)

func prefixToNodeDir(ctx *ShellCtxt, s []string) (*model.Node, string) {
	node := ctx.node
	isPrefix := len(s) > 0 && s[len(s)-1] != ""

	log.Trace.Println("isPrefix", isPrefix)

	if !isPrefix {
		return node, ""
	}

	prefix := unescapeSpaces(s[len(s)-1])

	log.Trace.Println("prefix", prefix)

	node, err := ctx.api.Filetree.NodeByPath(prefix, ctx.node)

	// Prefix matches an entry
	if err == nil {
		if node.IsDirectory() {
			if strings.HasSuffix(prefix, "/") {
				return node, prefix
			} else {
				return node.Parent, path.Dir(prefix)
			}
		}
	}

	base := path.Base(prefix)

	log.Trace.Println("base", base)

	if base == prefix {
		return ctx.node, ""
	}

	dir := path.Dir(prefix)

	log.Trace.Println("dir", dir)

	node, err = ctx.api.Filetree.NodeByPath(dir, ctx.node)

	if err != nil {
		return nil, ""
	}

	return node, dir
}

type nodeCheckFn func(*model.Node) bool

func createDirCompleter(ctx *ShellCtxt) func([]string) []string {
	return createCompleter(ctx, func(n *model.Node) bool { return n.IsDirectory() })
}

func createFileCompleter(ctx *ShellCtxt) func([]string) []string {
	return createCompleter(ctx, func(n *model.Node) bool { return n.IsFile() })
}

func createEntryCompleter(ctx *ShellCtxt) func([]string) []string {
	return createCompleter(ctx, func(n *model.Node) bool { return true })
}

func createCompleter(ctx *ShellCtxt, check nodeCheckFn) func([]string) []string {
	return func(s []string) []string {
		options := make([]string, 0)

		log.Trace.Println("completer:", s, len(s))

		node, dir := prefixToNodeDir(ctx, s)

		if node == nil {
			return options
		}

		for _, n := range node.Children {
			if !check(n) {
				continue
			}

			var entry string
			if n.IsDirectory() {
				entry = fmt.Sprintf("%s/", n.Name())
			} else {
				entry = fmt.Sprintf("%s", n.Name())
			}

			if dir != "" {
				if !strings.HasSuffix(dir, "/") {
					dir += "/"
				}
				entry = fmt.Sprintf("%s%s", dir, entry)
			}

			entry = escapeSpaces(entry)

			options = append(options, entry)
		}

		log.Trace.Println("options", options)

		return options
	}
}
