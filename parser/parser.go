package parser

import (
	"regexp"
	"strings"

	"github.com/arm1n18/nimble/protocol"
)

func ParseCommand(s string) (string, []string) {
	r := regexp.MustCompile(`"((?:\\.|[^"\\])*)"|(\S+)`)
	cmd := r.FindAllString(s, -1)[0]
	args := r.FindAllString(s, -1)[1:]

	for i, a := range args {
		if len(a) >= 2 && a[0] == '"' && a[len(a)-1] == '"' {
			a = a[1 : len(a)-1]
		}
		args[i] = strings.ReplaceAll(a, "\\\"", "\"")
	}

	return strings.ToUpper(cmd), args
}

func IsPatternCmd(cmd string) bool {
	var ok bool

	switch {
	case strings.HasSuffix(cmd, "*"):
		i := strings.Index(cmd, "*")
		if len(cmd[i:]) > 1 {
			protocol.ErrorMessage("Invalid pattern command")
			return false
		}
		ok = true
	case strings.HasSuffix(cmd, "?"):
		i := strings.Index(cmd, "?")
		for _, l := range cmd[i:] {
			if l != '?' {
				protocol.ErrorMessage("Invalid pattern command")
				return false
			}
		}

		return true
	default:
		protocol.ErrorMessage("Invalid pattern command")
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

func IsKeyAllowed(keys ...string) bool {
	for _, key := range keys {
		key = strings.Trim(key, "")
		if key == "" || key == "*" || key == " " {
			return false
		}
	}

	return true
}
