package handlers

import (
	"cache/logger"
	"net/http"
)

func HandleRequests(w http.ResponseWriter, r *http.Request) {
	// switch r.Method {
	// case "GET":
	// 	fmt.Printf("[Cache] %v 		%v |\033[42m %v \033[0m|	%s\n", time.Now().Format("2006-01-02 15:04:05"), r.Host, r.Method, r.URL.Query().Get("url"))
	// case "POST":
	// 	fmt.Printf("[Cache] %v 		%v |\033[44m %v \033[0m|\n", time.Now().Format("2006-01-02 15:04:05"), r.Host, r.Method)
	// case "DELETE":
	// 	fmt.Printf("[Cache] %v 		%v |\033[41m %v \033[0m|\n", time.Now().Format("2006-01-02 15:04:05"), r.Host, r.Method)
	// }

	logger.Request(r)
}

// Validate and remove Quotes
func removeQuotes(s *[]string, st, n int) bool {
	for i := st; i < len(*s); i += n {
		if len((*s)[i]) >= 2 {

			if (*s)[i][0] == '\'' || (*s)[i][len((*s)[i])-1] == '\'' {
				logger.Error("Invalid syntax")
				return false
			}

			if (*s)[i][0] != '"' && (*s)[i][len((*s)[i])-1] != '"' {
				// (*s)[i] = (*s)[i][1 : len((*s)[i])-1]
				continue
			} else if (*s)[i][0] != '"' || (*s)[i][len((*s)[i])-1] != '"' {
				logger.Error("Invalid syntax")
				return false
			} else {
				(*s)[i] = (*s)[i][1 : len((*s)[i])-1]
			}
		}
	}

	return true
}
