package server

import (
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
	name      string
	listeners int
	msgChan   chan []byte
}

// Server uses TCP to implement a Pub/Sub messaging pattern.
type Server struct {
	logger *zerolog.Logger

	mu     *sync.RWMutex
	topics map[string]*topic
}

func NewServer(l *zerolog.Logger) *Server {
	return &Server{
		logger: l,
		mu:     &sync.RWMutex{},
		topics: map[string]*topic{},
	}
}

// Start listens on the given TCP network address and asynchronously handles connections.
func (svr *Server) Start(addr string) error {

	// Create a TCP listener.
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return errors.Wrapf(err, "cannot listen on %s", addr)
	}
	defer ln.Close()

	for {

		// Get the next connection.
		conn, err := ln.Accept()
		if err != nil {
			svr.logger.Error().Stack().Err(err).Msg("Cannot accept connection")
			continue
		}

		// Handle the connection in another go-routine.
		go svr.handleConn(conn)
	}
}

// handleConn executes operations based on the messages recived by the connection.
func (svr *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	// Read incoming messages from the connection in another go-routine.
	connChan := make(chan string)
	go svr.read(conn, connChan)

	// Wait for a message from the connection, or for connChan to close.
	msg, ok := <-connChan
	if !ok {
		return
	}

	// Compare the message's first argument.
	switch strings.SplitN(msg, " ", 2)[0] {

	case operationPublish:

		// Split the message into 3 sub-strings: operation, topic, and data.
		args := strings.SplitN(msg, " ", 3)

		// Check for missing arguments.
		if len(args) != 3 {
			svr.writeErr(conn, problemDetail{
				PDType: pdTypeInvalidCommand,
				Detail: fmt.Sprintf("Operation '%s' requires a topic and data as arguments.", operationPublish),
			})

			return
		}

		// Publish the data to the topic.
		if err := svr.publish(args[1], []byte(args[2])); err != nil {
			svr.writeErr(conn, errors.Wrap(err, "cannot publish"))
			return
		}

	case operationSubscribe:

		// Split the message into at least 2 sub-strings: operation and topics.
		args := strings.Split(msg, " ")

		// Check for missing arguments.
		if len(args) < 2 {
			svr.writeErr(conn, problemDetail{
				PDType: pdTypeInvalidCommand,
				Detail: fmt.Sprintf("Operation '%s' requires at least one topic as an argument.", operationSubscribe),
			})

			return
		}

		// Subscribe to the topics.
		if err := svr.subscribe(conn, connChan, args[1:]...); err != nil {
			svr.writeErr(conn, errors.Wrap(err, "cannot subscribe"))
			return
		}

	default:
		svr.writeErr(conn, problemDetail{
			PDType: pdTypeInvalidCommand,
			Detail: "Operation must be 'pub' or 'sub'.",
		})
	}
}
