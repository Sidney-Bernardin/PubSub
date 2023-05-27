package server

import (
	"context"
	"net"

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
func (svr *Server) subscribe(conn net.Conn, connChan chan string, topicNames ...string) error {

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

				// Return, when the context gets canceled.
				case <-ctx.Done():
					return

				// Listen for and send the topic's messages to writeChan, to be
				// written to the connection.
				case msg := <-topic_.msgChan:
					writeChan <- msg
				}
			}
		}(i, topic_)
	}

	for {
		select {

		// Return, if connChan closes.
		case _, ok := <-connChan:
			if !ok {
				return nil
			}

		// Write messages from writeChan to the connection.
		case msg := <-writeChan:
			if _, err := conn.Write(msg); err != nil {
				return errors.Wrap(err, "cannot write to connection.")
			}
		}
	}
}
