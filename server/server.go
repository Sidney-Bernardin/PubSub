package server

import (
	"context"
	"fmt"
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
	logger *zerolog.Logger

	mu     *sync.RWMutex
	topics map[string]*topic
}

func NewServer(l *zerolog.Logger) *Server {

	svr := &Server{
		topics: map[string]*topic{},
		logger: l,
		mu:     &sync.RWMutex{},
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

func (svr *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	var (
		err         error
		ctx, cancel = context.WithCancel(context.Background())
		msgChan     = make(chan []byte)
	)

	go svr.readAndWait(conn, cancel, msgChan)

	msg := strings.TrimSpace(string(<-msgChan))
	operation := strings.SplitN(msg, " ", 2)[0]

	switch operation {

	case operationPublish:

		args := strings.SplitN(msg, " ", 3)

		if len(args) != 3 {
			err = problemDetail{
				PDType: pdTypeInvalidCommand,
				Detail: fmt.Sprintf("Operation '%s' requires a topic and data as arguments.", operationPublish),
			}
			break
		}

		err = svr.publish(args[1], []byte(args[2]))

	case operationSubscribe:

		args := strings.Split(msg, " ")

		if len(args) < 2 {
			err = problemDetail{
				PDType: pdTypeInvalidCommand,
				Detail: fmt.Sprintf("Operation '%s' requires a topic as an argument.", operationSubscribe),
			}
			break
		}

		err = svr.subscribe(ctx, conn, args[1])

	default:
		err = problemDetail{
			PDType: pdTypeInvalidCommand,
			Detail: "Operation must be 'pub' or 'sub'.",
		}
	}

	svr.writeErr(conn, errors.Wrapf(err, "cannot handle operation '%s'", err))
}

func (svr *Server) readAndWait(conn net.Conn, cancel context.CancelFunc, msgChan chan []byte) {
	defer cancel()

	sent := false
	buf := make([]byte, 2048)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		if !sent {
			msgChan <- buf[:n]
			sent = true
		}
	}
}
