package server

import (
	"net"
	"strings"

	"github.com/pkg/errors"
)

// getTopics returns the specified topics from the server and creates any that don't exist.
func (svr *Server) getTopics(topicNames ...string) []*topic {
	topics := make([]*topic, len(topicNames))

	for i, topicName := range topicNames {

		// Get the topic.
		svr.mu.RLock()
		topic_, ok := svr.topics[topicName]
		svr.mu.RUnlock()

		// If the topic doesn't exist, create a new one.
		svr.mu.Lock()
		if !ok {
			topic_ = &topic{
				name:    topicName,
				msgChan: make(chan []byte),
			}
			svr.topics[topicName] = topic_
		}
		svr.mu.Unlock()

		topics[i] = topic_
	}

	return topics
}

type readResponse struct {
	msg string
	err error
}

// read sends messages from the connection to the read-channel. While the
// read-channel is full, messages are ignored.
func (svr *Server) read(conn net.Conn, readChan chan readResponse) {
	buf := make([]byte, 2048)

	for {

		// Listen for messages from the connection.
		n, err := conn.Read(buf)
		if err != nil {
			readChan <- readResponse{
				err: errors.Wrap(err, "cannot read from connection"),
			}

			return
		}

		// If the read-channel has space, send the message to it.
		select {
		case readChan <- readResponse{
			msg: strings.TrimSpace(string(buf[:n])),
		}:
		default:
		}
	}
}

// addListener increments the topic's listener count by one.
func (svr *Server) addListener(topic_ *topic) {
	svr.mu.Lock()
	defer svr.mu.Unlock()
	topic_.listeners++
}

// removeListener decrements the topic's listener count by one. If it gets to
// zero, the topic is deleted.
func (svr *Server) removeListener(topic_ *topic) {
	svr.mu.Lock()
	defer svr.mu.Unlock()

	if topic_.listeners = topic_.listeners - 1; topic_.listeners < 1 {
		delete(svr.topics, topic_.name)
	}
}
