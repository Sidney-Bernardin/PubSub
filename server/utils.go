package server

import (
	"net"
	"strings"
)

// read sends any messages read from the connection to msgChan. If the
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
