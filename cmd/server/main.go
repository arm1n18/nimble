package main

import (
	"bufio"
	"fmt"
	"net"
	cmd "nimble/commands"
	"nimble/formatter"
	"nimble/secure"
	"nimble/storage"
	"nimble/utils"
	"regexp"
	"strings"
	"time"
)

type Request struct {
	Message string `json:"Message"`
	Body    string `json:"Body"`
}

func registerServer() {
	l, err := net.Listen("tcp", ":8085")
	if err != nil {
		formatter.FatalError("failed to register server: %v\n", err)
	}

	defer l.Close()

	c := storage.CreateCache()
	go c.StartBgCleanup(5 * time.Second)

	for {
		conn, _ := l.Accept()
		go handleConnection(c, conn)
	}
}

func handleConnection(c *storage.Cache, conn net.Conn) {
	defer conn.Close()

	session := &secure.Session{Authorized: false}
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		response := handleCommand(c, session, scanner.Text())
		fmt.Fprintln(conn, response)
	}
}

func handleCommand(c *storage.Cache, s *secure.Session, str string) string {
	cmdName, cmdArgs := utils.ParseCommand(str)

	if cmdName == "AUTH" {
		if len(cmdArgs) != 1 {
			return "Password is empty"
		}

		if !secure.Authenticate(cmdArgs[0]) {
			return "Invalid password"
		}

		s.Authorized = true
		return "OK"
	}

	if secure.SecureConnection(c) && !s.Authorized {
		return "Authentication required."
	}

	if cmdName == "PING" {
		return "PONG"
	}

	switch cmdName {
	case "SET":
		if len(cmdArgs) != 2 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return secure.ReadOnlyMiddleware(c, func() string {
			return cmd.SET(c, cmdArgs[0], cmdArgs[1])
		})
	case "MSET":
		if len(cmdArgs) < 2 || len(cmdArgs)%2 != 0 {
			return formatter.ErrInvalidSyntax.Error()
		}

		res := regexp.MustCompile(`"[^"]*"|\S+`)
		return secure.ReadOnlyMiddleware(c, func() string {
			return cmd.MSET(c, res.FindAllString(strings.Join(cmdArgs, " "), -1)...)
		})
	case "GET":
		if len(cmdArgs) != 1 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.GET(c, cmdArgs[0])
	case "MGET":
		if len(cmdArgs) == 0 {
			return formatter.ErrInvalidSyntax.Error()
		}

		cmd.MGET(c, cmdArgs...)
	case "KEYS":
		if len(cmdArgs) == 0 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.KEYS(c, cmdArgs...)
	case "HSET":
		if len(cmdArgs[1:]) == 0 {
			return formatter.ErrInvalidSyntax.Error()
		}

		res := regexp.MustCompile(`"[^"]*"|\S+`)
		matches := res.FindAllString(strings.Join(cmdArgs, " "), -1)
		return secure.ReadOnlyMiddleware(c, func() string {
			return cmd.HSET(c, cmdArgs[0], matches[1:]...)
		})
	case "SADD":
		if len(cmdArgs[1:]) == 0 {
			return formatter.ErrInvalidSyntax.Error()
		}

		res := regexp.MustCompile(`"[^"]*"|\S+`)
		matches := res.FindAllString(strings.Join(cmdArgs, " "), -1)
		return secure.ReadOnlyMiddleware(c, func() string {
			return cmd.SADD(c, cmdArgs[0], matches[1:]...)
		})
	case "SREM":
		if len(cmdArgs[1:]) == 0 {
			return formatter.ErrInvalidSyntax.Error()
		}

		res := regexp.MustCompile(`"[^"]*"|\S+`)
		matches := res.FindAllString(strings.Join(cmdArgs, " "), -1)
		return secure.ReadOnlyMiddleware(c, func() string {
			return cmd.SREM(c, cmdArgs[0], matches[1:]...)
		})
	case "SLEN":
		if len(cmdArgs) > 1 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.SLEN(c, cmdArgs[0])
	case "ZADD":
		if len(cmdArgs) != 3 {
			return formatter.ErrorMessage("Ivalid syntax")
		}

		return cmd.ZADD(c, cmdArgs[0], cmdArgs[1], cmdArgs[2])
	case "ZRANGEBYSCORE":
		if len(cmdArgs) != 3 {
			return formatter.ErrorMessage("Ivalid syntax")
		}

		return cmd.ZRANGEBYSCORE(c, cmdArgs[0], cmdArgs[1], cmdArgs[2])
	case "ZREM":
		if len(cmdArgs) != 2 {
			return formatter.ErrInvalidSyntax.Error()
		}
		return cmd.ZREM(c, cmdArgs[0], cmdArgs[1])
	case "SCORE":
		if len(cmdArgs) != 2 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.SCORE(c, cmdArgs[0], cmdArgs[1])
	case "LSCORE":
		if len(cmdArgs) < 2 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.LSCORE(c, cmdArgs[0], cmdArgs[1:])
	case "SMEMBERS":
		if len(cmdArgs) > 1 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.SMEMBERS(c, cmdArgs[0])
	case "HGET":
		if len(cmdArgs[1:]) == 0 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.HGET(c, cmdArgs[0], cmdArgs[1:]...)
	case "HDEL":
		if len(cmdArgs[1:]) == 0 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.HDEL(c, cmdArgs[0], cmdArgs[1:]...)
	case "HLEN":
		if len(cmdArgs) > 1 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.HLEN(c, cmdArgs[0])
	case "HKEYS":
		if len(cmdArgs) > 1 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.HKEYS(c, cmdArgs[0])
	case "HVALUES":
		if len(cmdArgs) > 1 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.HVALUES(c, cmdArgs[0])
	case "DEL":
		if len(cmdArgs) == 0 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return secure.ReadOnlyMiddleware(c, func() string {
			return cmd.DEL(c, cmdArgs...)
		})
	case "COPY":
		if len(cmdArgs) != 2 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return secure.ReadOnlyMiddleware(c, func() string {
			return cmd.COPY(c, cmdArgs[0], cmdArgs[1])
		})
	case "RENAME":
		if len(cmdArgs) != 2 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return secure.ReadOnlyMiddleware(c, func() string {
			return cmd.RENAME(c, cmdArgs[0], cmdArgs[1])
		})
	case "ESET":
		if len(cmdArgs) != 1 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return secure.ReadOnlyMiddleware(c, func() string {
			return cmd.ESET(c, cmdArgs[0])
		})
	case "LSET":
		if len(cmdArgs) < 2 {
			return formatter.ErrInvalidSyntax.Error()
		}

		res := regexp.MustCompile(`"[^"]*"|\S+`)
		matches := res.FindAllString(strings.Join(cmdArgs, " "), -1)
		return secure.ReadOnlyMiddleware(c, func() string {
			return cmd.LSET(c, cmdArgs[0], matches[1:]...)
		})
	case "LGET":
		if len(cmdArgs) < 1 {
			return formatter.ErrInvalidSyntax.Error()
		}

		res := regexp.MustCompile(`"[^"]*"|\S+`)
		matches := res.FindAllString(strings.Join(cmdArgs, " "), -1)
		cmd.LGET(c, cmdArgs[0], matches[1:]...)
	case "LCLEAR":
		if len(cmdArgs) > 1 {
			return formatter.ErrInvalidSyntax.Error()
		}
		cmd.LCLEAR(c, cmdArgs[0])
	case "LLEN":
		if len(cmdArgs) > 1 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.LLEN(c, cmdArgs[0])
	case "SPUSH":
		if len(cmdArgs) < 2 {
			return formatter.ErrInvalidSyntax.Error()
		}

		res := regexp.MustCompile(`"[^"]*"|\S+`)
		matches := res.FindAllString(strings.Join(cmdArgs, " "), -1)
		return secure.ReadOnlyMiddleware(c, func() string {
			return cmd.SPUSH(c, cmdArgs[0], matches[1:]...)
		})
	case "EPUSH":
		if len(cmdArgs) < 2 {
			return formatter.ErrInvalidSyntax.Error()
		}

		res := regexp.MustCompile(`"[^"]*"|\S+`)
		matches := res.FindAllString(strings.Join(cmdArgs, " "), -1)
		return secure.ReadOnlyMiddleware(c, func() string {
			return cmd.SPUSH(c, cmdArgs[0], matches[1:]...)
		})
	case "SPOP":
		if len(cmdArgs) > 2 {
			return formatter.ErrInvalidSyntax.Error()
		}
		if len(cmdArgs) == 1 {
			cmdArgs = append(cmdArgs, "")
		}

		return secure.ReadOnlyMiddleware(c, func() string {
			return cmd.SPOP(c, cmdArgs[0], cmdArgs[1])
		})
	case "EPOP":
		if len(cmdArgs) > 2 {
			return formatter.ErrInvalidSyntax.Error()
		}
		if len(cmdArgs) == 1 {
			cmdArgs = append(cmdArgs, "")
		}

		return secure.ReadOnlyMiddleware(c, func() string {
			return cmd.EPOP(c, cmdArgs[0], cmdArgs[1])
		})
	case "SRANGE":
		if len(cmdArgs) != 3 {
			return formatter.ErrInvalidSyntax.Error()
		}
		cmd.SRANGE(c, cmdArgs[0], cmdArgs[1:])
	case "TTK":
		if len(cmdArgs) != 2 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return secure.ReadOnlyMiddleware(c, func() string {
			return cmd.TTK(c, cmdArgs[0], cmdArgs[1])
		})
	case "TTL":
		if len(cmdArgs) != 1 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.TTL(c, cmdArgs[0])
	case "COMPARE":
		if len(cmdArgs) > 2 {
			return formatter.ErrInvalidSyntax.Error()
		}
		cmd.COMPARE(c, cmdArgs)
	case "LIST":
		return cmd.LIST(c)
	case "LISTLEN":
		return cmd.LISTLEN(c)
	case "INCR":
		if len(cmdArgs) > 2 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return secure.ReadOnlyMiddleware(c, func() string {
			return cmd.INCR(c, cmdArgs[0])
		})
	case "DECR":
		if len(cmdArgs) > 2 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return secure.ReadOnlyMiddleware(c, func() string {
			return cmd.DECR(c, cmdArgs[0])
		})
	case "INCRBY":
		if len(cmdArgs) > 2 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return secure.ReadOnlyMiddleware(c, func() string {
			return cmd.INCRBY(c, cmdArgs[0], cmdArgs[1])
		})
	case "DECRBY":
		if len(cmdArgs) > 2 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return secure.ReadOnlyMiddleware(c, func() string {
			return cmd.DECRBY(c, cmdArgs[0], cmdArgs[1])
		})
	case "MULL":
		if len(cmdArgs) > 2 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return secure.ReadOnlyMiddleware(c, func() string {
			return cmd.MULL(c, cmdArgs[0], cmdArgs[1])
		})
	case "DIV":
		if len(cmdArgs) > 2 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return secure.ReadOnlyMiddleware(c, func() string {
			return cmd.DIV(c, cmdArgs[0], cmdArgs[1])
		})
	case "EXISTS":
		if len(cmdArgs) == 0 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.EXISTS(c, cmdArgs...)
	case "LEXISTS":
		if len(cmdArgs[1:]) == 0 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.LEXISTS(c, cmdArgs...)
	case "HCONTAINS":
		if len(cmdArgs[1:]) == 0 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.HCONTAINS(c, cmdArgs[0], cmdArgs[1:]...)
	case "LHCONTAINS":
		if len(cmdArgs[1:]) == 0 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.LHCONTAINS(c, cmdArgs[0], cmdArgs[1:]...)
	case "CONTAINS":
		if len(cmdArgs[1:]) == 0 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.CONTAINS(c, cmdArgs[0], cmdArgs[1:])
	case "LCONTAINS":
		if len(cmdArgs[1:]) == 0 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.LCONTAINS(c, cmdArgs[0], cmdArgs[1:])
	case "SCONTAINS":
		if len(cmdArgs[1:]) == 0 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.SCONTAINS(c, cmdArgs[0], cmdArgs[1:]...)
	case "LSCONTAINS":
		if len(cmdArgs[1:]) == 0 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.LSCONTAINS(c, cmdArgs[0], cmdArgs[1:]...)
	case "INDEXOF":
		if len(cmdArgs[1:]) == 0 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.INDEXOF(c, cmdArgs[0], cmdArgs[1])
	case "TYPE":
		if len(cmdArgs) > 1 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return cmd.TYPE(c, cmdArgs[0])
	case "MODE":
		if len(cmdArgs) == 0 {
			return fmt.Sprint(c.GetMode())

		} else if len(cmdArgs) > 1 {
			return formatter.ErrInvalidSyntax.Error()
		}

		return secure.ChangeMode(c, strings.ToUpper(cmdArgs[0]))
	default:
		// fmt.Println("\033[33;1mUnknown command (type \\HELP)\033[0m")
		return "Unknown command (type \\HELP)"
	}

	return ""
}

func main() {
	registerServer()
}
