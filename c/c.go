package c

import "net/http"

var router = http.NewServeMux()

func GET(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	router.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			handler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func POST(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	router.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func PUT(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	router.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "PUT" {
			handler(w, r)
		}
	})
}

func DELETE(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			handler(w, r)
		}
	})
}

func GetRouter() *http.ServeMux {
	return router
}
