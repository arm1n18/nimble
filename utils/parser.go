package utils

import (
	"nimble/formatter"
	"strings"
)

func ParseCommand(input string) (string, []string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", []string{}
	}
	cmdName := strings.ToUpper(parts[0])
	cmdArgs := parts[1:]

	return cmdName, cmdArgs
}

// If ends with * or ? then it's a pattern command
func IsPatternCmd(cmd string) bool {
	var ok bool

	switch {
	case strings.HasSuffix(cmd, "*"):
		i := strings.Index(cmd, "*")
		if len(cmd[i:]) > 1 {
			formatter.ErrorMessage("Invalid pattern command")
			return false
		}
		ok = true
	case strings.HasSuffix(cmd, "?"):
		i := strings.Index(cmd, "?")
		for _, l := range cmd[i:] {
			if l != '?' {
				formatter.ErrorMessage("Invalid pattern command")
				return false
			}
		}

		return true
	default:
		formatter.ErrorMessage("Invalid pattern command")
		ok = false
	}

	return ok
}

func GetPatternSymbol(pattern string) (string, bool) {
	switch {
	case strings.HasSuffix(pattern, "*"):
		return "*", true
	case strings.HasSuffix(pattern, "?"):
		return "?", true
	default:
		return "", false
	}
}

func removeQuotes(s *[]string, st, n int) bool {
	for i := st; i < len(*s); i += n {
		if len((*s)[i]) >= 2 {

			if (*s)[i][0] == '\'' || (*s)[i][len((*s)[i])-1] == '\'' {
				formatter.ErrInvalidSyntax.Error()
				return false
			}

			if (*s)[i][0] != '"' && (*s)[i][len((*s)[i])-1] != '"' {
				// (*s)[i] = (*s)[i][1 : len((*s)[i])-1]
				continue
			} else if (*s)[i][0] != '"' || (*s)[i][len((*s)[i])-1] != '"' {
				formatter.ErrInvalidSyntax.Error()
				return false
			} else {
				(*s)[i] = (*s)[i][1 : len((*s)[i])-1]
			}
		}
	}

	return true
}
