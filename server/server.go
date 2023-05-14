package server

import (
	"context"
	"net"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

const (
	operationPublish   = "pub"
	operationSubscribe = "sub"
)

type topic struct {
	subs int
	ch   chan []byte
}

type Server struct {
	topics map[string]*topic

	logger *zerolog.Logger
	mu     *sync.Mutex
}

func NewServer(l *zerolog.Logger) *Server {

	svr := &Server{
		topics: map[string]*topic{},
		logger: l,
		mu:     &sync.Mutex{},
	}

	return svr
}

func (svr *Server) Start(addr string) error {

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.Wrapf(err, "cannot listen on %s", addr)
	}
	defer ln.Close()

	for {

		conn, err := ln.Accept()
		if err != nil {
			svr.logger.Error().Stack().Err(err).Msg("Cannot accept connection")
			continue
		}

		go svr.handleConn(conn)
	}
}

func (svr *Server) readAndWait(conn net.Conn, cancel context.CancelFunc, msgChan chan []byte) {
	defer cancel()

	sent := false
	buf := make([]byte, 2048)

	for {
		if _, err := conn.Read(buf); err != nil {
			return
		}

		if !sent {
			msgChan <- buf
			sent = true
		}
	}
}

func (svr *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	ctx, cancel := context.WithCancel(context.Background())
	msgChan := make(chan []byte)

	go svr.readAndWait(conn, cancel, msgChan)

	msg := <-msgChan

	params := strings.Split(string(msg), " ")
	if len(params) != 3 {
		svr.writeErr(conn, problemDetail{
			PDType: pdTypeInvalidCommand,
			Detail: "Command must look like this: [operation] [topic] [data].",
		})
		return
	}

	var err error

	switch params[0] {
	case operationPublish:
		err = svr.publish(params[1], []byte(params[2]))
	case operationSubscribe:
		err = svr.subscribe(ctx, conn, params[1])
	default:
		err = problemDetail{
			PDType: pdTypeInvalidOperation,
			Detail: "Operation must be 'pub' or 'sub'.",
		}
	}

	if err != nil {
		svr.writeErr(conn, errors.Wrapf(err, "cannot handle %s", params[0]))
	}
}
