package formatter

import "fmt"

type Error struct {
	Code    string
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("-ERR %s: %s", e.Code, e.Message)
}

var ErrWrongKey = Error{Code: "WRONG_KEY", Message: "key can`t be empty or *"}
var ErrNotEnoughValues = Error{Code: "NOT_ENOUGH_VALUES", Message: "not enough values provided"}
var ErrNotANumber = Error{Code: "NOT_A_NUMBER", Message: "value is not a number"}
var ErrInvalidTTL = Error{Code: "INVALID_TTL", Message: "ttl must be a number and >= -1"}
var ErrInvalidSyntax = Error{Code: "INVALID_SYNTAX", Message: "cmd has invalid syntax"}
var ErrMismatchType = Error{Code: "MISMATCH_TYPE", Message: "value doesn`t match stored type"}
var ErrInvalidRange = Error{Code: "INVALID_RANGE", Message: "range is invalid"}
var ErrInvalidScore = Error{Code: "INVALID_SCORE", Message: "score must be >= 0"}
