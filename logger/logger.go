package logger

import (
	"fmt"
	"log"
	"time"

	"github.com/arm1n18/nimble/config"
)

type Logger struct {
}

const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Underline = "\033[4m"
	Red       = "\033[31m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Blue      = "\033[34m"
	Magenta   = "\033[95m"
	Cyan      = "\033[36m"
	Gray      = "\033[37m"
	White     = "\033[97m"
)

func Error(format string, a ...any) {
	log.Printf(Red+"Error: "+format+Reset+"\n", a...)
}

func Success(format string, a ...any) {
	log.Printf(Green+format+Reset+"\n", a...)
}

func Help(format string, a ...any) {
	fmt.Printf(Yellow+format+Reset+"\n", a...)
}

func ServerInfo(c config.Config) {
	fmt.Println(Bold + Cyan + "Host" + Reset + "  : " + c.Host)
	fmt.Println(Bold + Cyan + "Port" + Reset + "  : " + fmt.Sprint(c.Port))
	fmt.Println(Bold + Yellow + "Mode" + Reset + "  : " + c.GetMode())
	fmt.Println(Bold + Green + "Users" + Reset + " : " + fmt.Sprint(c.GetUsers()))
	fmt.Println(Bold + Magenta + "Date" + Reset + "  : " + fmt.Sprint(time.Now()))
}

func NewLogger(path string) (*Logger, error) {

	return nil, nil
}
