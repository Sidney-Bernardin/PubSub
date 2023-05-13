package server

import (
	"context"
	"net"

	"github.com/pkg/errors"
)

func (svr *Server) publish(ctx context.Context, conn net.Conn, topicName string, data []byte) error {

	topic_, ok := svr.topics[topicName]
	if !ok {
		return problemDetail{
			PDType: pdTypeTopicDoesNotExist,
		}
	}

	for i := 0; i < topic_.subs; i++ {
		topic_.ch <- data
	}

	return nil
}

func (svr *Server) subscribe(ctx context.Context, conn net.Conn, topicName string, data []byte) error {

	svr.mu.Lock()
	topic_, ok := svr.topics[topicName]
	if !ok {
		topic_ = &topic{0, make(chan []byte)}
		svr.topics[topicName] = topic_
	}
	topic_.subs++
	svr.mu.Unlock()

	defer func() {
		if topic_.subs = topic_.subs - 1; topic_.subs <= 0 {
			svr.mu.Lock()
			delete(svr.topics, topicName)
			svr.mu.Unlock()
		}
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
