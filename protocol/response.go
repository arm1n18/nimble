package protocol

import (
	"fmt"
	"log"
	"os"
)

type Response struct {
	Success bool
	Output  string
}

type NumberType interface {
	int | float64
}

// +OK - Success +
// -ERR error - Error +
// _nil - No value +
// $value - String +-
// :number - Number +
// *[array] - Array
// #[hash] - Hash

func FatalError(format string, a ...any) {
	log.Printf("Error: "+format+"\n", a...)
	os.Exit(1)
}

func Ok() string {
	return "+OK"
}

func OkMessage(a ...any) string {
	return fmt.Sprintf("+OK %v", a...)
}

func Success() string {
	return ":1"
}

func Failure() string {
	return ":0"
}

func ErrorMessage(a ...any) string {
	return fmt.Sprintf("-ERR %v", a...)
}

func String(a string) string {
	return fmt.Sprintf("$%s", a)
}

func Number[T NumberType](a T) string {
	return fmt.Sprintf(":%v", a)
}

func Array(a string) string {
	return fmt.Sprintf("*%s", a)
}

func Hash(a string) string {
	return fmt.Sprintf("#%s", a)
}

func Nil() string {
	return "-nil"
}
