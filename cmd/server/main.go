package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	cmd "github.com/arm1n18/nimble/commands"
	"github.com/arm1n18/nimble/config"
	"github.com/arm1n18/nimble/logger"
	"github.com/arm1n18/nimble/parser"
	"github.com/arm1n18/nimble/protocol"
	"github.com/arm1n18/nimble/storage"
)

type Session struct {
	Authorized bool
}

type Request struct {
	Message string `json:"Message"`
	Body    string `json:"Body"`
}

func registerServer(conf *config.Config) {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", conf.Host, conf.Port))
	if err != nil {
		protocol.FatalError("failed to register server: %v\n", err)
	}

	logger.ServerInfo(*conf)

	defer l.Close()

	s := storage.CreateCache(conf.MaxHistory)

	go s.BGGC(5 * time.Second)

	for {
		conn, _ := l.Accept()
		go handleConnection(s, conn, conf)
	}
}

func handleConnection(s *storage.Cache, conn net.Conn, conf *config.Config) {
	defer conn.Close()

	session := &Session{Authorized: false}
	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		resp := handleCommand(s, session, conf, line)
		if resp == "Client disconnected" {
			return
		} else {
			fmt.Fprintln(conn, resp)
		}
	}
}

func handleCommand(s *storage.Cache, session *Session, conf *config.Config, line string) string {
	cmdName, cmdArgs := parser.ParseCommand(line)

	if cmdName == "AUTH" {
		if session.Authorized {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !conf.SecureConnection() {
			// add error
			return protocol.ErrInvalidSyntax.Error()
		}

		if len(cmdArgs) != 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		user, password := cmdArgs[0], cmdArgs[1]

		if !conf.Authenticate(user, password) {
			return protocol.ErrorMessage("Invalid username or password")
		}

		session.Authorized = true
		return "OK"
	}

	if conf.SecureConnection() && !session.Authorized {
		return "Authentication required."
	}

	if cmdName == "PING" {
		return "PONG"
	}

	switch cmdName {
	case "SET":
		if len(cmdArgs) < 2 || len(cmdArgs) > 3 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if len(cmdArgs) == 2 {
			cmdArgs = append(cmdArgs, "-1")
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			resp := cmd.SET(s, cmdArgs[0], cmdArgs[1], cmdArgs[2])
			if resp.Success {
				s.AddToHistory(cmdArgs[0], line)
			}

			return resp.Output
		})
	case "MSET":
		if len(cmdArgs) < 2 || len(cmdArgs)%2 != 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs...) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			resp := cmd.MSET(s, cmdArgs...)
			if resp.Success {
				s.AddToHistory(cmdArgs[0], line)
			}

			return resp.Output
		})
	case "GET":
		if len(cmdArgs) != 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.GET(s, cmdArgs[0])
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "MGET":
		if len(cmdArgs) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs...) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.MGET(s, cmdArgs...)
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "KEYS":
		if len(cmdArgs) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		resp := cmd.KEYS(s, cmdArgs...)
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "HSET":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			resp := cmd.HSET(s, cmdArgs[0], cmdArgs[1:]...)
			if resp.Success {
				s.AddToHistory(cmdArgs[0], line)
			}

			return resp.Output
		})
	case "HGET":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.HGET(s, cmdArgs[0], cmdArgs[1:]...)
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "HDEL":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.HDEL(s, cmdArgs[0], cmdArgs[1:]...)
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "HLEN":
		if len(cmdArgs) > 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.HLEN(s, cmdArgs[0])
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "HKEYS":
		if len(cmdArgs) > 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.HKEYS(s, cmdArgs[0])
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "HVALUES":
		if len(cmdArgs) > 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.HVALUES(s, cmdArgs[0])
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "DEL":
		if len(cmdArgs) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs...) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			resp := cmd.DEL(s, cmdArgs...)
			if resp.Success {
				s.AddToHistory(cmdArgs[0], line)
			}

			return resp.Output
		})
	case "COPY":
		if len(cmdArgs) != 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs...) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			resp := cmd.COPY(s, cmdArgs[0], cmdArgs[1])
			if resp.Success {
				s.AddToHistory(cmdArgs[0], line)
			}

			return resp.Output
		})
	case "RENAME":
		if len(cmdArgs) != 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs...) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			resp := cmd.RENAME(s, cmdArgs[0], cmdArgs[1])
			if resp.Success {
				s.AddToHistory(cmdArgs[0], line)
			}

			return resp.Output
		})
	case "ESET":
		if len(cmdArgs) != 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			resp := cmd.ESET(s, cmdArgs[0])
			if resp.Success {
				s.AddToHistory(cmdArgs[0], line)
			}

			return resp.Output
		})
	case "LSET":
		if len(cmdArgs) < 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			resp := cmd.LSET(s, cmdArgs[0], cmdArgs[1:]...)
			if resp.Success {
				s.AddToHistory(cmdArgs[0], line)
			}

			return resp.Output
		})
	case "LGET":
		if len(cmdArgs) < 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.LGET(s, cmdArgs[0], cmdArgs[1:]...)
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "LCLEAR":
		if len(cmdArgs) > 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.LCLEAR(s, cmdArgs[0])
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "LLEN":
		if len(cmdArgs) > 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		resp := cmd.LLEN(s, cmdArgs[0])
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "SPUSH":
		if len(cmdArgs) < 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			resp := cmd.SPUSH(s, cmdArgs[0], cmdArgs[1:]...)
			if resp.Success {
				s.AddToHistory(cmdArgs[0], line)
			}

			return resp.Output
		})
	case "EPUSH":
		if len(cmdArgs) < 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			resp := cmd.EPUSH(s, cmdArgs[0], cmdArgs[1:]...)
			if resp.Success {
				s.AddToHistory(cmdArgs[0], line)
			}

			return resp.Output
		})
	case "SPOP":
		if len(cmdArgs) > 2 {
			return protocol.ErrInvalidSyntax.Error()
		}
		if len(cmdArgs) == 1 {
			cmdArgs = append(cmdArgs, "")
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			resp := cmd.SPOP(s, cmdArgs[0], cmdArgs[1])
			if resp.Success {
				s.AddToHistory(cmdArgs[0], line)
			}

			return resp.Output
		})
	case "EPOP":
		if len(cmdArgs) > 2 {
			return protocol.ErrInvalidSyntax.Error()
		}
		if len(cmdArgs) == 1 {
			cmdArgs = append(cmdArgs, "")
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			resp := cmd.EPOP(s, cmdArgs[0], cmdArgs[1])
			if resp.Success {
				s.AddToHistory(cmdArgs[0], line)
			}

			return resp.Output
		})
	case "SRANGE":
		if len(cmdArgs) != 3 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.SRANGE(s, cmdArgs[0], cmdArgs[1:])
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "SADD":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			resp := cmd.SADD(s, cmdArgs[0], cmdArgs[1:]...)
			if resp.Success {
				s.AddToHistory(cmdArgs[0], line)
			}

			return resp.Output
		})
	case "SREM":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			resp := cmd.SREM(s, cmdArgs[0], cmdArgs[1:]...)
			if resp.Success {
				s.AddToHistory(cmdArgs[0], line)
			}

			return resp.Output
		})
	case "SLEN":
		if len(cmdArgs) > 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.SLEN(s, cmdArgs[0])
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "SMEMBERS":
		if len(cmdArgs) > 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.SMEMBERS(s, cmdArgs[0])
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "ZADD":
		if len(cmdArgs) != 3 {
			return protocol.ErrorMessage("Ivalid syntax")
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.ZADD(s, cmdArgs[0], cmdArgs[1], cmdArgs[2])
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "ZREM":
		if len(cmdArgs) != 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.ZREM(s, cmdArgs[0], cmdArgs[1])
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "ZRANGE":
		if len(cmdArgs) != 3 {
			return protocol.ErrorMessage("Ivalid syntax")
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.ZRANGE(s, cmdArgs[0], cmdArgs[1], cmdArgs[2])
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "SCORE":
		if len(cmdArgs) != 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.SCORE(s, cmdArgs[0], cmdArgs[1])
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "LSCORE":
		if len(cmdArgs) < 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.LSCORE(s, cmdArgs[0], cmdArgs[1:]...)
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "TTK":
		if len(cmdArgs) != 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			resp := cmd.TTK(s, cmdArgs[0], cmdArgs[1])
			if resp.Success {
				s.AddToHistory(cmdArgs[0], line)
			}

			return resp.Output
		})
	case "TTL":
		if len(cmdArgs) != 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.TTL(s, cmdArgs[0])
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "LIST":
		resp := cmd.LIST(s)
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "LISTLEN":
		resp := cmd.LISTLEN(s)
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "INCR":
		if len(cmdArgs) > 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			resp := cmd.INCR(s, cmdArgs[0])
			if resp.Success {
				s.AddToHistory(cmdArgs[0], line)
			}

			return resp.Output
		})
	case "DECR":
		if len(cmdArgs) > 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			resp := cmd.DECR(s, cmdArgs[0])
			if resp.Success {
				s.AddToHistory(cmdArgs[0], line)
			}

			return resp.Output
		})
	case "INCRBY":
		if len(cmdArgs) > 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			resp := cmd.INCRBY(s, cmdArgs[0], cmdArgs[1])
			if resp.Success {
				s.AddToHistory(cmdArgs[0], line)
			}

			return resp.Output
		})
	case "DECRBY":
		if len(cmdArgs) > 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			resp := cmd.DECRBY(s, cmdArgs[0], cmdArgs[1])
			if resp.Success {
				s.AddToHistory(cmdArgs[0], line)
			}

			return resp.Output
		})
	case "MUL":
		if len(cmdArgs) > 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			resp := cmd.MUL(s, cmdArgs[0], cmdArgs[1])
			if resp.Success {
				s.AddToHistory(cmdArgs[0], line)
			}

			return resp.Output
		})
	case "DIV":
		if len(cmdArgs) > 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			resp := cmd.DIV(s, cmdArgs[0], cmdArgs[1])
			if resp.Success {
				s.AddToHistory(cmdArgs[0], line)
			}

			return resp.Output
		})
	case "EXISTS":
		if len(cmdArgs) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs...) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.EXISTS(s, cmdArgs...)
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "LEXISTS":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs...) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.LEXISTS(s, cmdArgs...)
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "HCONTAINS":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.HCONTAINS(s, cmdArgs[0], cmdArgs[1:]...)
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "LHCONTAINS":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.LHCONTAINS(s, cmdArgs[0], cmdArgs[1:]...)
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "CONTAINS":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.CONTAINS(s, cmdArgs[0], cmdArgs[1:])
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "LCONTAINS":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.LCONTAINS(s, cmdArgs[0], cmdArgs[1:])
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "SCONTAINS":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.SCONTAINS(s, cmdArgs[0], cmdArgs[1:]...)
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "LSCONTAINS":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.LSCONTAINS(s, cmdArgs[0], cmdArgs[1:]...)
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "INDEXOF":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.INDEXOF(s, cmdArgs[0], cmdArgs[1])
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "TYPE":
		if len(cmdArgs) > 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		resp := cmd.TYPE(s, cmdArgs[0])
		if resp.Success {
			s.AddToHistory(cmdArgs[0], line)
		}

		return resp.Output
	case "MODE":
		if len(cmdArgs) == 0 {
			return fmt.Sprint(conf.GetMode())

		} else if len(cmdArgs) > 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		str, _ := conf.SetMode(strings.ToUpper(cmdArgs[0]))
		return str
	case "EXIT":
		return protocol.OkMessage("Client disconnected")
	default:
		return "Unknown command (type \\HELP)"
	}
}

func main() {
	conf := config.CreateConfig()
	registerServer(conf)
}
