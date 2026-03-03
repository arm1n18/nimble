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

	s := storage.CreateCache()

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

		response := handleCommand(s, session, conf, line)
		if response == "Client disconnected" {
			return
		} else {
			fmt.Fprintln(conn, response)
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
			return cmd.SET(s, cmdArgs[0], cmdArgs[1], cmdArgs[2])
		})
	case "MSET":
		if len(cmdArgs) < 2 || len(cmdArgs)%2 != 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs...) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			return cmd.MSET(s, cmdArgs...)
		})
	case "GET":
		if len(cmdArgs) != 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.GET(s, cmdArgs[0])
	case "MGET":
		if len(cmdArgs) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs...) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.MGET(s, cmdArgs...)
	case "KEYS":
		if len(cmdArgs) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		return cmd.KEYS(s, cmdArgs...)
	case "HSET":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			return cmd.HSET(s, cmdArgs[0], cmdArgs[1:]...)
		})
	case "HGET":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.HGET(s, cmdArgs[0], cmdArgs[1:]...)
	case "HDEL":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.HDEL(s, cmdArgs[0], cmdArgs[1:]...)
	case "HLEN":
		if len(cmdArgs) > 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.HLEN(s, cmdArgs[0])
	case "HKEYS":
		if len(cmdArgs) > 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.HKEYS(s, cmdArgs[0])
	case "HVALUES":
		if len(cmdArgs) > 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.HVALUES(s, cmdArgs[0])
	case "DEL":
		if len(cmdArgs) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs...) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			return cmd.DEL(s, cmdArgs...)
		})
	case "COPY":
		if len(cmdArgs) != 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs...) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			return cmd.COPY(s, cmdArgs[0], cmdArgs[1])
		})
	case "RENAME":
		if len(cmdArgs) != 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs...) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			return cmd.RENAME(s, cmdArgs[0], cmdArgs[1])
		})
	case "ESET":
		if len(cmdArgs) != 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			return cmd.ESET(s, cmdArgs[0])
		})
	case "LSET":
		if len(cmdArgs) < 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			return cmd.LSET(s, cmdArgs[0], cmdArgs[1:]...)
		})
	case "LGET":
		if len(cmdArgs) < 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.LGET(s, cmdArgs[0], cmdArgs[1:]...)
	case "LCLEAR":
		if len(cmdArgs) > 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.LCLEAR(s, cmdArgs[0])
	case "LLEN":
		if len(cmdArgs) > 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		return cmd.LLEN(s, cmdArgs[0])
	case "SPUSH":
		if len(cmdArgs) < 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			return cmd.SPUSH(s, cmdArgs[0], cmdArgs[1:]...)
		})
	case "EPUSH":
		if len(cmdArgs) < 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			return cmd.EPUSH(s, cmdArgs[0], cmdArgs[1:]...)
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
			return cmd.SPOP(s, cmdArgs[0], cmdArgs[1])
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
			return cmd.EPOP(s, cmdArgs[0], cmdArgs[1])
		})
	case "SRANGE":
		if len(cmdArgs) != 3 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.SRANGE(s, cmdArgs[0], cmdArgs[1:])
	case "SADD":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			return cmd.SADD(s, cmdArgs[0], cmdArgs[1:]...)
		})
	case "SREM":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			return cmd.SREM(s, cmdArgs[0], cmdArgs[1:]...)
		})
	case "SLEN":
		if len(cmdArgs) > 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.SLEN(s, cmdArgs[0])
	case "SMEMBERS":
		if len(cmdArgs) > 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.SMEMBERS(s, cmdArgs[0])
	case "ZADD":
		if len(cmdArgs) != 3 {
			return protocol.ErrorMessage("Ivalid syntax")
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.ZADD(s, cmdArgs[0], cmdArgs[1], cmdArgs[2])
	case "ZREM":
		if len(cmdArgs) != 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.ZREM(s, cmdArgs[0], cmdArgs[1])
	case "ZRANGE":
		if len(cmdArgs) != 3 {
			return protocol.ErrorMessage("Ivalid syntax")
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.ZRANGE(s, cmdArgs[0], cmdArgs[1], cmdArgs[2])
	case "SCORE":
		if len(cmdArgs) != 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.SCORE(s, cmdArgs[0], cmdArgs[1])
	case "LSCORE":
		if len(cmdArgs) < 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.LSCORE(s, cmdArgs[0], cmdArgs[1:]...)
	case "TTK":
		if len(cmdArgs) != 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			return cmd.TTK(s, cmdArgs[0], cmdArgs[1])
		})
	case "TTL":
		if len(cmdArgs) != 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.TTL(s, cmdArgs[0])
	case "LIST":
		return cmd.LIST(s)
	case "LISTLEN":
		return cmd.LISTLEN(s)
	case "INCR":
		if len(cmdArgs) > 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			return cmd.INCR(s, cmdArgs[0])
		})
	case "DECR":
		if len(cmdArgs) > 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			return cmd.DECR(s, cmdArgs[0])
		})
	case "INCRBY":
		if len(cmdArgs) > 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			return cmd.INCRBY(s, cmdArgs[0], cmdArgs[1])
		})
	case "DECRBY":
		if len(cmdArgs) > 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			return cmd.DECRBY(s, cmdArgs[0], cmdArgs[1])
		})
	case "MUL":
		if len(cmdArgs) > 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			return cmd.MUL(s, cmdArgs[0], cmdArgs[1])
		})
	case "DIV":
		if len(cmdArgs) > 2 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return conf.ReadOnlyMiddleware(func() string {
			return cmd.DIV(s, cmdArgs[0], cmdArgs[1])
		})
	case "EXISTS":
		if len(cmdArgs) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs...) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.EXISTS(s, cmdArgs...)
	case "LEXISTS":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs...) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.LEXISTS(s, cmdArgs...)
	case "HCONTAINS":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.HCONTAINS(s, cmdArgs[0], cmdArgs[1:]...)
	case "LHCONTAINS":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.LHCONTAINS(s, cmdArgs[0], cmdArgs[1:]...)
	case "CONTAINS":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.CONTAINS(s, cmdArgs[0], cmdArgs[1:])
	case "LCONTAINS":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.LCONTAINS(s, cmdArgs[0], cmdArgs[1:])
	case "SCONTAINS":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.SCONTAINS(s, cmdArgs[0], cmdArgs[1:]...)
	case "LSCONTAINS":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.LSCONTAINS(s, cmdArgs[0], cmdArgs[1:]...)
	case "INDEXOF":
		if len(cmdArgs[1:]) == 0 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.INDEXOF(s, cmdArgs[0], cmdArgs[1])
	case "TYPE":
		if len(cmdArgs) > 1 {
			return protocol.ErrInvalidSyntax.Error()
		}

		if !parser.IsKeyAllowed(cmdArgs[0]) {
			return protocol.ErrWrongKey.Error()
		}

		return cmd.TYPE(s, cmdArgs[0])
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
