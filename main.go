package main

import (
	"log"
	"net/http"
)

func main() {

	svr := http.Server{
		Addr:    "localhost:8080",
		Handler: newPubSub(),
	}

	log.Fatal(svr.ListenAndServe())
}
