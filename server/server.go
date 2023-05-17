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
	subs int
	ch   chan []byte
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

	// Listen for incoming from the connection in another go-routine.
	msgChan := make(chan string)
	go svr.read(conn, msgChan)

	// Wait for a message.
	msg, ok := <-msgChan
	if !ok {
		return
	}

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

		// Split the message into at least 2 sub-strings: operation, topic.
		args := strings.Split(msg, " ")

		// Check for missing arguments.
		if len(args) < 2 {
			svr.writeErr(conn, problemDetail{
				PDType: pdTypeInvalidCommand,
				Detail: fmt.Sprintf("Operation '%s' requires a topic as an argument.", operationSubscribe),
			})

			return
		}

		// Subscribe to the topic.
		if err := svr.subscribe(conn, msgChan, args[1]); err != nil {
			svr.writeErr(conn, errors.Wrap(err, "cannot publish"))
			return
		}

	default:
		svr.writeErr(conn, problemDetail{
			PDType: pdTypeInvalidCommand,
			Detail: "Operation must be 'pub' or 'sub'.",
		})
	}
}

// read sends any messages recived from the connection to msgChan. If the
// connection can't be read from, msgChan closes and read returns.
func (svr *Server) read(conn net.Conn, msgChan chan string) {
	defer close(msgChan)

	buf := make([]byte, 2048)

	for {

		// Listen for incoming messages from the connection.
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		// If msgChan has space, send the message to it.
		select {
		case msgChan <- strings.TrimSpace(string(buf[:n])):
		default:
		}
	}
}
