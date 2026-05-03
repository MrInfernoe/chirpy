package main

import (
	"net/http"
	"fmt"
)

func main() {
	// matches the URL of each incoming request against a list of registered patterns and calls the handler for the pattern that most closely matches the URL. 
	serveMux := http.NewServeMux()

	root := http.Dir("./web/static")
	patternHandler := http.FileServer(root)
	serveMux.Handle("/", patternHandler)

	server := http.Server{}
	server.Addr = ":8080"
	server.Handler = serveMux

	err := server.ListenAndServe()
	if err == http.ErrServerClosed {
		fmt.Println("server closed")
	} else {
		fmt.Println(err)
	}
}