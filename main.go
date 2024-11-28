package main

import "net/http"

func main() {
	servMux := http.ServeMux{}
	serv := http.Server{
		Addr:    ":8080",
		Handler: &servMux,
	}

	handler := http.FileServer(http.Dir("."))
	servMux.Handle("/", handler)

	serv.ListenAndServe()
}
