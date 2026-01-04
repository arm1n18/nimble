package main

import (
	"bufio"
	cmd "cache/commands"
	"cache/logger"
	"cache/secure"
	"cache/storage"
	"cache/utils"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

var (
	commands = []struct {
		cmd    string
		desc   string
		format string
	}{
		{"SET", "SAVE DATA TO CACHE", "SET [K1] V1 [K2] V2 ..."},
		{"NUMSET", "SAVE DATA TO CACHE AS NUMBER", "NUMSET [K1] N1 [K2] N2 ..."},
		{"GET", "GET DATA FROM CACHE", "GET [K1] [K2] ..."},
		{"KEYS", "GET KEYS FROM CACHE BY PATTERN", "KEYS [PATTERN]"},
		{"HSET", "STORE HASHSET IN THE CACHE", "HSET H_NAME [K1] V1 [K2] V2 ..."},
		{"HGET", "GET HASHSET DATA FROM THE CACHE BY KEY", "HGET H_NAME [K1] [K2] ..."},
		{"HKEYS", "GET HASHSET DATA KEYS", "HKEYS H_NAME"},
		{"HVALUES", "GET HASHSET DATA VALUES", "HVALUES H_NAME"},
		{"HDEL", "DELETE HASHSET DATA FROM THE CACHE BY KEY", "HDEL H_NAME [K1] [K2] ..."},
		{"HLEN", "GET LENGTH OF THE HASHSET", "HLEN H_NAME"},
		{"DEL", "REMOVE ANY TYPE OF DATA FROM THE CACHE", "DEL [K1] [K2] ..."},
		{"COPY", "COPY DATA FROM ONE STRUCTURE TO ANOTHER", "COPY [F_K] [T_K]"},
		{"LSET", "SET VALUE AT INDEX IN THE ARRAY", "LSET ARR_NAME [I1] [V1] [I2] [V2] ..."},
		{"LGET", "GET VALUE AT INDEX IN THE ARRAY", "LGET ARR_NAME [I1] [I2] ..."},
		{"LCLEAR", "CLEAR THE ARRAY", "LCLEAR ARR_NAME"},
		{"LLEN", "GET LENGTH OF THE ARRAY", "LLEN ARR_NAME"},
		{"SPUSH", "PUSH TO THE START OF THE ARRAY", "SPUSH ARR_NAME V1 V2 ..."},
		{"EPUSH", "PUSH TO THE END OF THE ARRAY", "EPUSH ARR_NAME V1 V2 ..."},
		{"SRANGE", "GET LIST OF AN ARRAY IN A GIVEN RANGE", "SRANGE ARR_NAME [F_INDEX] [T_INDEX]"},
		{"TTK", "DATA LIFESPAN BEFORE DELETING", "TTK [K1] TIME"},
		{"TTL", "TIME LEFT BEFORE DATA IS DELETED", "TTL [K1] TIME"},
		{"COMPARE", "CHECK IF THE MEMORY SIZES OF THE DATA ARE EQUAL", "COMPARE [K1] [K2]"},
		{"LIST", "SHOW ALL THE KEYS", "LIST"},
		{"INCR", "INCREASE THE NUMBER BY 1", "INCR [K1]"},
		{"DECR", "DECREASE THE NUMBER BY 1", "DECR [K1]"},
		{"INCRBY", "INCREASE THE NUMBER BY N", "INCRBY [K1] N"},
		{"DECRBY", "DECREASE THE NUMBER BY N", "DECRBY [K1] N"},
		{"MULL", "MULTIPLY THE NUMBER BY N", "MULL [K1] N"},
		{"DIV", "DIVIDE THE NUMBER BY N", "DIV [K1] N"},
		{"EXISTS", "CHECK IF THE KEYS EXISTS AND RETURNS THEIR QT", "EXISTS [K1] [K2] ..."},
		{"LEXISTS", "CHECK IF THE KEYS EXISTS AND RETURNS ARRAY", "LEXISTS [K1] [K2] ..."},
		{"CONTAINS", "CHECK IF THE VALUES EXISTS IN THE ARRAY AND RETURNS THEIR QT", "CONTAINS ARR_NAME [V1] [V2] ..."},
		{"LCONTAINS", "CHECK IF THE VALUES EXISTS IN THE ARRAY AND RETURNS ARRAY", "LCONTAINS ARR_NAME [V1] [V2] ..."},
		{"HCONTAINS", "CHECK IF THE KEYS EXISTS IN THE HASHSET AND RETURNS THEIR QT", "HCONTAINS H_NAME [K1] [K2] ..."},
		{"LHCONTAINS", "CHECK IF THE KEYS EXISTS IN THE HASHSET AND RETURNS ARRAY", "LHCONTAINS H_NAME [K1] [K2] ..."},
		{"TYPE", "SHOW THE TYPE OF THE VALUE", "TYPE [K1]"},
		{"MODE", "CHANGE THE MODE OF THE CACHE", "MODE [READONLY/READWRITE]"},
		{"PSET", "SET THE PASSWORD", "PSET [PASSWORD]"},
		{"CLS", "ERASE ALL PREVIOUS TEXT AND OUTPUT", ""},
	}
)

func HandleRequests(w http.ResponseWriter, r *http.Request) {
	// if r.Method == "GET" {
	// 	handlers.Save(w, r)
	// } else if r.Method == "POST" {
	// 	handlers.Save(w, r)
	// }
}

func serverStart() {
	http.HandleFunc("/", HandleRequests)
	if err := http.ListenAndServe(":8085", nil); err != nil {
		log.Fatal(err)
	}
}

func main() {
	go serverStart()

	c := storage.CreateCache()

	scanner := bufio.NewScanner(os.Stdin)

	protected := secure.SecureConnection(c)

	if protected {
		fmt.Printf(logger.Bold + logger.Cyan + "[Cache] Enter password: " + logger.Reset)

		for scanner.Scan() {
			parts := strings.Fields(scanner.Text())
			res := regexp.MustCompile(`"[^"]*"|\S+`)
			if len(parts) == 0 {
				fmt.Printf(logger.Bold + logger.Cyan + "[Cache] Enter password: " + logger.Reset)
				continue
			}

			password := strings.Join(res.FindAllString(strings.Join(parts, " "), -1), " ")

			if strings.ToLower(password) == "exit" {
				return
			}

			if authenticate := secure.Authenticate(password); authenticate {
				fmt.Println(logger.Green + "Authenticated successfully!" + logger.Reset)
				break
			} else {
				fmt.Printf(logger.Bold + logger.Cyan + "[Cache] Wrong password. Try again: " + logger.Reset)
			}
		}
	}

	fmt.Printf(logger.Bold + logger.Cyan + "[Cache] Enter command: " + logger.Reset)

	for scanner.Scan() {
		input := scanner.Text()

		if input == "exit" {
			break
		}
		cmdName, cmdArgs := utils.ParseCommand(input)

		switch strings.ToUpper(cmdName) {
		case "SET":
			if len(cmdArgs) < 2 || len(cmdArgs)%2 != 0 {
				logger.Error("Invalid syntax")
				break
			}

			res := regexp.MustCompile(`"[^"]*"|\S+`)
			secure.ReadOnlyMiddleware(c, func() {
				cmd.SET(c, res.FindAllString(strings.Join(cmdArgs, " "), -1))
			})
		case "NUMSET":
			if len(cmdArgs) == 0 || len(cmdArgs[1:]) == 0 {
				logger.Error("Invalid syntax")
				break
			}
			secure.ReadOnlyMiddleware(c, func() {
				cmd.NUMSET(c, cmdArgs)
			})
		case "GET":
			if len(cmdArgs) == 0 {
				logger.Error("Invalid syntax")
				break
			}
			cmd.GET(c, cmdArgs)
		case "KEYS":
			if len(cmdArgs) == 0 {
				logger.Error("Invalid syntax")
				break
			}
			cmd.KEYS(c, cmdArgs)
		case "HSET":
			if len(cmdArgs[1:]) == 0 {
				logger.Error("Invalid syntax")
				break
			}
			res := regexp.MustCompile(`"[^"]*"|\S+`)
			matches := res.FindAllString(strings.Join(cmdArgs, " "), -1)
			secure.ReadOnlyMiddleware(c, func() {
				cmd.HSET(c, cmdArgs[0], matches[1:])
			})
		case "HGET":
			if len(cmdArgs[1:]) == 0 {
				logger.Error("Invalid syntax")
				break
			}
			cmd.HGET(c, cmdArgs[0], cmdArgs[1:])
		case "HDEL":
			if len(cmdArgs[1:]) == 0 {
				logger.Error("Invalid syntax")
				break
			}
			cmd.HDEL(c, cmdArgs[0], cmdArgs[1:])
		case "HLEN":
			if len(cmdArgs) > 1 {
				logger.Error("Invalid syntax")
				break
			}
			cmd.HLEN(c, cmdArgs[0])
		case "HKEYS":
			if len(cmdArgs) > 1 {
				logger.Error("Invalid syntax")
				break
			}
			cmd.HKEYS(c, cmdArgs[0])
		case "HVALUES":
			if len(cmdArgs) > 1 {
				logger.Error("Invalid syntax")
				break
			}
			cmd.HVALUES(c, cmdArgs[0])
		case "DEL":
			if len(cmdArgs) == 0 {
				logger.Error("Invalid syntax")
				break
			}
			secure.ReadOnlyMiddleware(c, func() {
				cmd.DEL(c, cmdArgs[0])
			})
		case "COPY":
			if len(cmdArgs) != 2 {
				logger.Error("Invalid syntax")
				break
			}
			secure.ReadOnlyMiddleware(c, func() {
				cmd.COPY(c, cmdArgs[0], cmdArgs[1])
			})
		case "LSET":
			if len(cmdArgs) < 2 {
				logger.Error("Invalid syntax")
				break
			}
			res := regexp.MustCompile(`"[^"]*"|\S+`)
			matches := res.FindAllString(strings.Join(cmdArgs, " "), -1)
			secure.ReadOnlyMiddleware(c, func() {
				cmd.LSET(c, cmdArgs[0], matches[1:])
			})
		case "LGET":
			if len(cmdArgs) < 1 {
				logger.Error("Invalid syntax")
				break
			}
			res := regexp.MustCompile(`"[^"]*"|\S+`)
			matches := res.FindAllString(strings.Join(cmdArgs, " "), -1)
			cmd.LGET(c, cmdArgs[0], matches[1:])
		case "LCLEAR":
			if len(cmdArgs) > 1 {
				logger.Error("Invalid syntax")
				break
			}
			cmd.LCLEAR(c, cmdArgs[0])
		case "LLEN":
			if len(cmdArgs) > 1 {
				logger.Error("Invalid syntax")
				break
			}
			cmd.LLEN(c, cmdArgs[0])
		case "SPUSH":
			if len(cmdArgs) < 2 {
				logger.Error("Invalid syntax")
				break
			}
			res := regexp.MustCompile(`"[^"]*"|\S+`)
			matches := res.FindAllString(strings.Join(cmdArgs, " "), -1)
			secure.ReadOnlyMiddleware(c, func() {
				cmd.SPUSH(c, cmdArgs[0], matches[1:])
			})
		case "EPUSH":
			if len(cmdArgs) < 2 {
				logger.Error("Invalid syntax")
				break
			}
			res := regexp.MustCompile(`"[^"]*"|\S+`)
			matches := res.FindAllString(strings.Join(cmdArgs, " "), -1)
			secure.ReadOnlyMiddleware(c, func() {
				cmd.SPUSH(c, cmdArgs[0], matches[1:])
			})
		case "SPOP":
			if len(cmdArgs) > 2 {
				logger.Error("Invalid syntax")
				break
			}
			secure.ReadOnlyMiddleware(c, func() {
				cmd.SPOP(c, cmdArgs[0], cmdArgs[1])
			})
		case "EPOP":
			if len(cmdArgs) > 2 {
				logger.Error("Invalid syntax")
				break
			}
			secure.ReadOnlyMiddleware(c, func() {
				cmd.EPOP(c, cmdArgs[0], cmdArgs[1])
			})
		case "SRANGE":
			if len(cmdArgs) != 3 {
				logger.Error("Invalid syntax")
				break
			}
			cmd.SRANGE(c, cmdArgs[0], cmdArgs[1:])
		case "TTK":
			if len(cmdArgs) != 2 {
				logger.Error("Invalid syntax")
				break
			}
			secure.ReadOnlyMiddleware(c, func() {
				cmd.TTK(c, cmdArgs[0], cmdArgs[1])
			})
		case "TTL":
			if len(cmdArgs) != 1 {
				logger.Error("Invalid syntax")
				break
			}
			cmd.TTL(c, cmdArgs[0])
		case "COMPARE":
			if len(cmdArgs) > 2 {
				logger.Error("Invalid syntax")
				break
			}
			cmd.COMPARE(c, cmdArgs)
		case "SIZEOF":
			if len(cmdArgs) == 0 {
				logger.Error("Invalid syntax")
				break
			}
			cmd.SIZEOF(c, cmdArgs)
		case "LIST":
			cmd.LIST(c)
		case "LISTLEN":
			cmd.LISTLEN(c)
		case "INCR":
			if len(cmdArgs) > 2 {
				logger.Error("Invalid syntax")
				break
			}
			secure.ReadOnlyMiddleware(c, func() {
				cmd.INCR(c, cmdArgs[0])
			})
		case "DECR":
			if len(cmdArgs) > 2 {
				logger.Error("Invalid syntax")
				break
			}
			secure.ReadOnlyMiddleware(c, func() {
				cmd.DECR(c, cmdArgs[0])
			})
		case "INCRBY":
			if len(cmdArgs) > 2 {
				logger.Error("Invalid syntax")
				break
			}
			secure.ReadOnlyMiddleware(c, func() {
				cmd.INCRBY(c, cmdArgs[0], cmdArgs[1])
			})
		case "DECRBY":
			if len(cmdArgs) > 2 {
				logger.Error("Invalid syntax")
				break
			}
			secure.ReadOnlyMiddleware(c, func() {
				cmd.DECRBY(c, cmdArgs[0], cmdArgs[1])
			})
		case "MULL":
			if len(cmdArgs) > 2 {
				logger.Error("Invalid syntax")
				break
			}
			secure.ReadOnlyMiddleware(c, func() {
				cmd.MULL(c, cmdArgs[0], cmdArgs[1])
			})
		case "DIV":
			if len(cmdArgs) > 2 {
				logger.Error("Invalid syntax")
				break
			}
			secure.ReadOnlyMiddleware(c, func() {
				cmd.DIV(c, cmdArgs[0], cmdArgs[1])
			})
		case "EXISTS":
			if len(cmdArgs) == 0 {
				logger.Error("Invalid syntax")
				break
			}
			cmd.EXISTS(c, cmdArgs)
		case "LEXISTS":
			if len(cmdArgs[1:]) == 0 {
				logger.Error("Invalid syntax")
				break
			}
			cmd.LEXISTS(c, cmdArgs)
		case "HCONTAINS":
			if len(cmdArgs[1:]) == 0 {
				logger.Error("Invalid syntax")
				break
			}
			cmd.HCONTAINS(c, cmdArgs[0], cmdArgs[1:])
		case "LHCONTAINS":
			if len(cmdArgs[1:]) == 0 {
				logger.Error("Invalid syntax")
				break
			}
			cmd.LHCONTAINS(c, cmdArgs[0], cmdArgs[1:])
		case "CONTAINS":
			if len(cmdArgs[1:]) == 0 {
				logger.Error("Invalid syntax")
				break
			}
			cmd.CONTAINS(c, cmdArgs[0], cmdArgs[1:])
		case "LCONTAINS":
			if len(cmdArgs[1:]) == 0 {
				logger.Error("Invalid syntax")
				break
			}
			cmd.LCONTAINS(c, cmdArgs[0], cmdArgs[1:])
		case "TYPE":
			if len(cmdArgs) > 1 {
				logger.Error("Invalid syntax")
				break
			}
			cmd.TYPE(c, cmdArgs[0])
		case "MODE":
			if len(cmdArgs) == 0 {
				fmt.Println(c.GetMode())
				break
			} else if len(cmdArgs) > 1 {
				logger.Error("Invalid syntax")
				break
			}
			secure.ChangeMode(c, strings.ToUpper(cmdArgs[0]))
		case "PSET":
			if len(cmdArgs) < 1 {
				logger.Error("Invalid syntax")
				break
			}
			res := regexp.MustCompile(`"[^"]*"|\S+`)
			secure.SetPassword(res.FindAllString(strings.Join(cmdArgs, " "), -1)[0])
		case "CLS":
			clearScreen()
		case "\\HELP", "?":
			fmt.Println("Available commands:")

			maxLen := struct {
				lenCmd  int
				lenDesc int
			}{
				lenCmd:  0,
				lenDesc: 0,
			}
			for _, c := range commands {
				if len(c.cmd) > maxLen.lenCmd {
					maxLen.lenCmd = len(c.cmd)
				}

				if len(c.desc) > maxLen.lenDesc {
					maxLen.lenDesc = len(c.desc)
				}
			}

			for _, c := range commands {
				logger.Commad(c.cmd, c.desc, c.format, maxLen.lenCmd+10, maxLen.lenDesc+5)
			}
		default:
			fmt.Println("\033[33;1mUnknown command (type \\HELP)\033[0m")
		}

		fmt.Printf(logger.Bold + logger.Cyan + "[Cache] Enter command: " + logger.Reset)
	}
}

func clearScreen() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	_ = cmd.Run()
}
