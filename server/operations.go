package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/pkg/errors"
)

// publish sends the given message to the specified topic for each one of it's listeners.
func (svr *Server) publish(topicName string, msg []byte) error {

	// Get the topic.
	svr.mu.RLock()
	topic_, ok := svr.topics[topicName]
	svr.mu.RUnlock()

	if !ok {
		return problemDetail{pdTypeTopicDoesNotExist, ""}
	}

	// Send the message to the topic for each one of it's listeners.
	svr.mu.Lock()
	for i := 0; i < topic_.listeners; i++ {
		topic_.msgChan <- msg
	}
	svr.mu.Unlock()

	return nil
}

// subscribe asynchronously listens for messages from the specified topics, and
// writes them to the connection.
func (svr *Server) subscribe(conn net.Conn, readChan chan readResponse, topicNames ...string) error {

	var (
		ctx, cancel = context.WithCancel(context.Background())
		writeChan   = make(chan []byte)
	)

	defer cancel()

	// Get the topics and listen for their messages from seperate go-routines.
	for i, topic_ := range svr.getTopics(topicNames...) {
		go func(i int, topic_ *topic) {

			// Add a listener to the topic and defer it's removal.
			svr.addListener(topic_)
			defer svr.removeListener(topic_)

			for {
				select {

				case <-ctx.Done():
					return

				// Listen for and send the topic's messages to the write-channel, to be
				// written to the connection.
				case msg := <-topic_.msgChan:
					writeChan <- msg
				}
			}
		}(i, topic_)
	}

	// Write an OK message to the connection.
	msg := fmt.Sprintf("successfully subscribed to %s", strings.Join(topicNames, ", "))
	if _, err := conn.Write([]byte(msg)); err != nil {
		return errors.Wrap(err, "cannot write to connection")
	}

	for {
		select {

		// Ignore messages and handle errors from the connection.
		case res := <-readChan:

			if res.err == nil {
				continue
			}

			if errors.Cause(res.err) == io.EOF {
				return nil
			}

			return errors.Wrap(res.err, "cannot read from connection")

		// Write messages from the write-channel to the connection.
		case msg := <-writeChan:
			if _, err := conn.Write(msg); err != nil {
				return errors.Wrap(err, "cannot write to connection")
			}
		}
	}
}
