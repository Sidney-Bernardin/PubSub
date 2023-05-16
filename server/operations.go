package server

import (
	"context"
	"net"

	"github.com/pkg/errors"
)

func (svr *Server) publish(topicName string, data []byte) error {

	svr.mu.RLock()
	topic_, ok := svr.topics[topicName]
	svr.mu.RUnlock()

	if !ok {
		return problemDetail{
			PDType: pdTypeTopicDoesNotExist,
		}
	}

	svr.mu.RLock()
	subs := topic_.subs
	svr.mu.RUnlock()

	for i := 0; i < subs; i++ {
		topic_.ch <- data
	}

	return nil
}

func (svr *Server) subscribe(ctx context.Context, conn net.Conn, topicName string) error {

	svr.mu.RLock()
	topic_, ok := svr.topics[topicName]
	svr.mu.RUnlock()

	svr.mu.Lock()
	if !ok {
		topic_ = &topic{0, make(chan []byte)}
		svr.topics[topicName] = topic_
	}
	topic_.subs++
	svr.mu.Unlock()

	defer func() {
		svr.mu.Lock()
		if topic_.subs = topic_.subs - 1; topic_.subs <= 0 {
			delete(svr.topics, topicName)
		}
		svr.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case d := <-topic_.ch:
			if _, err := conn.Write(d); err != nil {
				return errors.Wrap(err, "cannot write to connection")
			}
		}
	}
}
