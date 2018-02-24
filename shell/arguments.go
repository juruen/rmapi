package shell

import "strings"

func parseArguments(line string) []string {
	words := [][]rune{}

	words = append(words, make([]rune, 0))
	slashes := 0
	args := 0
	for _, v := range line {
		if v == ' ' {
			// Consume blank spaces between args
			if len(words[args]) == 0 {
				continue
			}

			if slashes > 0 && (slashes%2) == 1 {
				// Found escaped space within argument
				words[args] = append(words[args], v)
			} else {
				// End of argument
				words = append(words, make([]rune, 0))
				args = args + 1
			}

			slashes = 0
			continue
		}

		if v == '\\' {
			slashes = slashes + 1
		} else {
			slashes = 0
		}

		words[args] = append(words[args], v)
	}

	result := make([]string, 0)
	for _, w := range words {
		if len(w) == 0 {
			continue
		}

		result = append(result, string(w))
	}

	return result
}

func escapeSpaces(s string) string {
	return strings.Replace(s, " ", "\\ ", -1)
}

func unescapeSpaces(s string) string {
	return strings.Replace(s, "\\ ", " ", -1)
}
