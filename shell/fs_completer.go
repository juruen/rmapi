package shell

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/peerdavid/rmapi/log"
)

func prefixToDir(s []string) string {
	isPrefix := len(s) > 0 && s[len(s)-1] != ""

	log.Trace.Println("isPrefix", isPrefix)

	if !isPrefix {
		return "./"
	}

	prefix := unescapeSpaces(s[len(s)-1])

	log.Trace.Println("prefix", prefix)

	fstat, err := os.Stat(prefix)

	// Prefix matches an entry
	if err == nil {
		if fstat.IsDir() {
			if strings.HasSuffix(prefix, "/") {
				return prefix
			} else {
				return path.Dir(prefix)
			}
		}
	}

	base := path.Base(prefix)

	log.Trace.Println("base", base)

	if base == prefix {
		return ""
	}

	dir := path.Dir(prefix)

	log.Trace.Println("dir", dir)

	fstat, err = os.Stat(dir)

	if err != nil || !fstat.IsDir() {
		return ""
	}

	return dir
}

type fileCheckFn func(os.FileInfo) bool

func createFsDirCompleter(ctx *ShellCtxt) func([]string) []string {
	return createFsCompleter(func(e os.FileInfo) bool { return e.IsDir() })
}

func createFsFileCompleter(ctx *ShellCtxt) func([]string) []string {
	return createFsCompleter(func(e os.FileInfo) bool { return !e.IsDir() })
}

func createFsEntryCompleter() func([]string) []string {
	return createFsCompleter(func(e os.FileInfo) bool { return true })
}

func createFsCompleter(check fileCheckFn) func([]string) []string {
	return func(s []string) []string {
		options := make([]string, 0)

		log.Trace.Println("completer:", len(s))

		dir := prefixToDir(s)

		if dir == "" {
			return options
		}

		entries, err := ioutil.ReadDir(dir)

		if err != nil {
			return options
		}

		for _, n := range entries {
			if !check(n) {
				continue
			}

			var entry string
			if info, err := os.Stat(dir + "/" + n.Name()); err == nil && info.IsDir() {
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

			if !n.IsDir() && !strings.HasSuffix(entry, ".pdf") && !strings.HasSuffix(entry, ".epub") {
				continue
			}

			options = append(options, entry)
		}

		log.Trace.Println("options", options)

		return options
	}
}
