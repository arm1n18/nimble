package config

import (
	"flag"
	"log"
	"strings"

	"github.com/arm1n18/nimble/protocol"
)

type User struct {
	name     string
	password string
}

type Mode string

const (
	ReadOnly  Mode = "read-only"
	ReadWrite Mode = "read-write"
)

type Config struct {
	Host  string
	Port  int
	mode  Mode
	users []User
}

func CreateConfig() *Config {
	c := Config{}

	c.Host = *flag.String("host", "localhost", "Server host")
	c.Port = *flag.Int("port", 8085, "Server port")
	m := flag.String("mode", string(ReadWrite), "Server mode")
	usersFlag := flag.String("user", "nimble:default", "Server users")

	flag.Parse()

	md := Mode(*m)
	switch md {
	case ReadOnly, ReadWrite:
		c.mode = Mode(*m)
	default:
		log.Fatalf("invalid mode: %s (use %s or %s)", *m, ReadOnly, ReadWrite)
	}

	if *usersFlag != "" {
		users := strings.Split(*usersFlag, ",")

		for _, user := range users {
			parts := strings.Split(user, ":")
			c.users = append(c.users, User{
				name:     parts[0],
				password: strings.Join(parts[1:], ""),
			})
		}
	} else {
		c.users = append(c.users, User{
			name:     "nimble",
			password: "default",
		})
	}

	return &c
}

func (c *Config) getUser(name string) *User {
	for _, user := range c.users {
		if user.name == name {
			return &user
		}
	}

	return nil
}

func (c *Config) SecureConnection() bool {
	return len(c.users) > 0
}

func (c *Config) Authenticate(name, password string) bool {
	user := c.getUser(name)
	if user == nil {
		return false
	}

	return user.password == password
}

func (c *Config) SetMode(mode string) (string, bool) {
	var result string

	switch mode {
	case "READONLY":
		c.mode = ReadOnly
		result = protocol.OkMessage("Cache mode changed to READONLY")
	case "READWRITE":
		c.mode = ReadWrite
		result = protocol.OkMessage("Cache mode changed to READWRITE")
	default:
		result = protocol.ErrorMessage("Unknown mode: %s", mode)
		return result, false
	}

	return result, true
}

func (c *Config) GetMode() Mode {
	return c.mode
}

func (c *Config) GetUsers() []string {
	s := make([]string, 0, len(c.users))

	for _, v := range c.users {
		s = append(s, v.name)
	}

	return s
}

func (c *Config) ReadOnlyMiddleware(fn func() string) string {
	if c.GetMode() == ReadOnly {
		return protocol.ErrReadOnly.Error()
	}
	return fn()
}
