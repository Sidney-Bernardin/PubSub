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
	opCmdPublish   = "pub"
	opCmdSubscribe = "sub"
)

type operation func(ctx context.Context, conn net.Conn, topicName string, data []byte) error

type topic struct {
	subs int
	ch   chan []byte
}

type Server struct {
	topics map[string]*topic

	logger     *zerolog.Logger
	operations map[string]operation
	mu         *sync.Mutex
}

func NewServer(l *zerolog.Logger) *Server {

	svr := &Server{
		topics:     map[string]*topic{},
		logger:     l,
		operations: map[string]operation{},
		mu:         &sync.Mutex{},
	}

	svr.operations[opCmdPublish] = svr.publish
	svr.operations[opCmdSubscribe] = svr.subscribe

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

func (svr *Server) handleConn(conn net.Conn) {
	defer func() {
		conn.Close()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	msgChan := make(chan []byte)

	go func() {
		defer cancel()

		msg := make([]byte, 2048)

		for {
			if _, err := conn.Read(msg); err != nil {
				return
			}
			msgChan <- msg
			close(msgChan)
		}
	}()

	msg := <-msgChan

	params := strings.Split(string(msg), " ")
	if len(params) != 3 {
		svr.writeErr(conn, problemDetail{
			PDType: pdTypeInvalidCommand,
			Detail: "Command must look like this: [operation] [topic] [data].",
		})
		return
	}

	op, ok := svr.operations[params[0]]
	if !ok {
		svr.writeErr(conn, problemDetail{
			PDType: pdTypeInvalidOperation,
			Detail: "Operation must be 'pub' or 'sub'.",
		})
		return
	}

	if err := op(ctx, conn, params[1], []byte(params[2])); err != nil {
		svr.writeErr(conn, errors.Wrapf(err, "cannot handle %s", params[0]))
	}

	return
}
