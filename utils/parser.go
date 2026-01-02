package utils

import (
	"cache/logger"
	"strings"
)

func ParseCommand(input string) (string, []string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", []string{}
	}
	cmdName := parts[0]
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
			logger.Error("Invalid pattern command")
			return false
		}
		ok = true
	case strings.HasSuffix(cmd, "?"):
		i := strings.Index(cmd, "?")
		for _, l := range cmd[i:] {
			if l != '?' {
				logger.Error("Invalid pattern command")
				return false
			}
		}

		return true
	default:
		logger.Error("Invalid pattern command")
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
