package server

import (
	"net"

	"github.com/pkg/errors"
)

// publish sends data to the specified topic's subscribers.
func (svr *Server) publish(topicName string, data []byte) error {

	// Get the topic.
	svr.mu.RLock()
	topic_, ok := svr.topics[topicName]
	svr.mu.RUnlock()

	if !ok {
		return problemDetail{pdTypeTopicDoesNotExist, ""}
	}

	// Get the topic's subscriber count.
	svr.mu.RLock()
	subs := topic_.subs
	svr.mu.RUnlock()

	// For each subscriber, send the data to the topic's channel.
	for i := 0; i < subs; i++ {
		topic_.ch <- data
	}

	return nil
}

// subscribe listens for messages from the specified topic, then writes them to the connection.
func (svr *Server) subscribe(conn net.Conn, msgChan chan string, topicName string) error {

	// Get the topic.
	svr.mu.RLock()
	topic_, ok := svr.topics[topicName]
	svr.mu.RUnlock()

	// If the topic doesn't exist, create a new one.
	svr.mu.Lock()
	if !ok {
		topic_ = &topic{0, make(chan []byte)}
		svr.topics[topicName] = topic_
	}
	topic_.subs++
	svr.mu.Unlock()

	defer func() {

		// Decrement the topic's subscriber count. If it gose below 1, delete the topic.
		svr.mu.Lock()
		if topic_.subs = topic_.subs - 1; topic_.subs <= 0 {
			delete(svr.topics, topicName)
		}
		svr.mu.Unlock()
	}()

	for {
		select {

		// Listen for a close message from msgChan.
		case _, ok := <-msgChan:
			if !ok {
				return nil
			}

		// Listen for data from the topic.
		case d := <-topic_.ch:

			// Write the data to the connection.
			if _, err := conn.Write(d); err != nil {
				return errors.Wrap(err, "cannot write to connection")
			}
		}
	}
}
