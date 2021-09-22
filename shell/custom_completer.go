package shell

import (
	"strings"
)

type cmdToCompleter map[string]func([]string) []string

type shellPathCompleter struct {
	cmdCompleter cmdToCompleter
}

func (ic shellPathCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	if len(ic.cmdCompleter) == 0 {
		return nil, len(line)
	}

	words := parseArguments(string(line))

	var cWords []string
	prefix := ""
	if len(words) > 0 && line[pos-1] != ' ' {
		prefix = words[len(words)-1]
		cWords = ic.getWords(words[:len(words)-1], prefix)
	} else {
		cWords = ic.getWords(words, prefix)
	}

	var suggestions [][]rune
	for _, w := range cWords {
		if strings.HasPrefix(w, prefix) {
			suggestions = append(suggestions, []rune(strings.TrimPrefix(w, prefix)))
		}
	}
	if len(suggestions) == 1 && prefix != "" && string(suggestions[0]) == "" {
		suggestions = [][]rune{[]rune(" ")}
	}
	return suggestions, len(prefix)
}

func (ic shellPathCompleter) getWords(w []string, prefix string) []string {
	if len(w) == 0 {
		return make([]string, 0)
	}

	completer, ok := ic.cmdCompleter[w[0]]
	if !ok {
		return make([]string, 0)
	}

	args := make([]string, len(w))
	for i, v := range w[1:] {
		args[i] = v
	}
	args = append(args, prefix)

	return completer(args)
}
