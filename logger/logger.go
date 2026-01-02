package logger

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

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

func FatalError(format string, a ...any) {
	log.Printf(Red+"Error: "+format+Reset+"\n", a...)
	os.Exit(1)
}

func Success(format string, a ...any) {
	log.Printf(Green+format+Reset+"\n", a...)
}

func Help(format string, a ...any) {
	fmt.Printf(Yellow+format+Reset+"\n", a...)
}

func Commad(cmd, desc, format string, lenCmd, lenDesc int) {
	fmt.Printf(Bold+Green+"%-*s"+Reset+"%-*s"+Magenta+"%s"+Reset+"\n", lenCmd, cmd, lenDesc, desc, format)
}

func Request(r *http.Request, a ...any) {
	fmt.Printf("[Cache] %v 		%v |"+Yellow+r.Method+Reset+"|	%s\n", time.Now().Format("2006-01-02 15:04:05"), r.Host, r.URL.Query().Get("url"))
}
