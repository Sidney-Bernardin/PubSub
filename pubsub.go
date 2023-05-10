package main

import (
	"log"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

const (
	PS = "/ps/"
)

type pubSub struct {
	mux *http.ServeMux
}

func newPubSub() *pubSub {

	ps := &pubSub{
		mux: http.NewServeMux(),
	}

	ps.mux.HandleFunc("/pub", ps.publish)
	ps.mux.Handle("/sub", websocket.Server{
		Handler: ps.subscribe,
	})

	return ps
}

func (ps *pubSub) publish(w http.ResponseWriter, r *http.Request) {
	log.Println("HANDSHAKE")
}

func (ps *pubSub) subscribe(conn *websocket.Conn) {
	defer conn.Close()
	for {
		log.Println("SUBSCRIBE")
		time.Sleep(3 * time.Second)
	}
}

func (ps *pubSub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ps.mux.ServeHTTP(w, r)
}
