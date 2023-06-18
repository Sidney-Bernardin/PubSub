package server

import (
	"fmt"
	"net"
	"testing"
)

func TestSubscribe(t *testing.T) {

	tt := []struct {
		topics        []string
		topicMessages []string

		expectedFuncErr      error
		expectedNetResponses []readResponse
	}{
		{
			topics:        []string{"a"},
			topicMessages: []string{"Hello, a!"},

			expectedFuncErr: nil,
			expectedNetResponses: []readResponse{
				{msg: "successfully subscribed to a"},
				{msg: "Hello, a!"},
			},
		},
		{
			topics: []string{"a", "b"},
			topicMessages: []string{
				"Hello, a!",
				"Hello, b!",
			},

			expectedFuncErr: nil,
			expectedNetResponses: []readResponse{
				{msg: "successfully subscribed to a, b"},
				{msg: "Hello, a!"},
				{msg: "Hello, b!"},
			},
		},
		{
			topics:        []string{},
			topicMessages: []string{},

			expectedFuncErr: nil,
			expectedNetResponses: []readResponse{
				{msg: "successfully subscribed to"},
			},
		},
	}

	for i, tc := range tt {
		name := fmt.Sprintf("Subtest%v", i)
		t.Run(name, func(t *testing.T) {

			// Create a mock network connection.
			serverConn, clientConn := net.Pipe()

			// Create channels to recive the connection's messages from another go-routine.
			serverReadChan := make(chan readResponse)
			clientReadChan := make(chan readResponse)

			// Create a server.
			svr := NewServer(nil)

			// Read messages from both connections in another go-routine.
			go svr.read(serverConn, serverReadChan)
			go svr.read(clientConn, clientReadChan)

			// Subscribe to the text-case's topics in another go-routine.
			errChan := make(chan error, 1)
			go func(topics ...string) {
				errChan <- svr.subscribe(serverConn, serverReadChan, topics...)
			}(tc.topics...)

			for i, expected := range tc.expectedNetResponses {

				// After the first expected response, start sending topic-messages.
				if i > 0 {
					topicName := tc.topics[i-1]
					topicMessage := tc.topicMessages[i-1]
					svr.topics[topicName].msgChan <- []byte(topicMessage)
				}

				select {

				// Listen for a message from the connection.
				case res := <-clientReadChan:

					// Check if the message is expected.
					if res.msg != expected.msg {
						t.Fatalf("Expected '%s' got '%s'", expected.msg, res.msg)
					}

					// Check if the error is expected.
					if res.err != expected.err {
						t.Fatalf("Expected '%v' got '%v'", expected.err, res.err)
					}

				// Listen for the subscribe error.
				case err := <-errChan:

					// Check if the error is expected.
					if err != tc.expectedFuncErr {
						t.Fatalf("Expected '%v' got '%v'", tc.expectedFuncErr, err)
					}
				}
			}

			// Close the connection and check if the subscribe error is expected.
			clientConn.Close()
			if err := (<-errChan); err != tc.expectedFuncErr {
				t.Fatalf("Expected '%v' got '%v'", tc.expectedFuncErr, err)
			}
		})
	}
}

func TestPublish(t *testing.T) {

	tt := []struct {
		existingTopics []string

		topic    string
		topicMsg string

		expectedErr error
	}{
		{
			existingTopics: []string{"a"},

			topic:    "a",
			topicMsg: "Hello, from a!",

			expectedErr: nil,
		},
		{
			existingTopics: []string{},

			topic:    "a",
			topicMsg: "Hello, from a!",

			expectedErr: problemDetail{
				PDType: pdTypeTopicDoesNotExist,
			},
		},
	}

	for i, tc := range tt {
		name := fmt.Sprintf("Subtest%v", i)
		t.Run(name, func(t *testing.T) {

			// Create a server.
			svr := NewServer(nil)

			// Create the test-case's existing topics.
			svr.getTopics(tc.existingTopics...)

			// Get the test-case's topic.
			topic_, ok := svr.topics[tc.topic]

			// If the test-case's topic exists, listen for a topic-message from
			// it in another go-routine.
			topicMsgChan := make(chan string, 1)
			if ok {
				topic_.listeners += 1
				go func() {
					topicMsgChan <- string(<-topic_.msgChan)
				}()
			}

			// Publish the test-case's topic-message.
			if err := svr.publish(tc.topic, []byte(tc.topicMsg)); err != tc.expectedErr {
				t.Fatalf("Expected '%v' got '%v'", tc.expectedErr, err)
			}

			// If the test-case's topic exists, listen for a message from it to
			// ckeck if it's expected.
			if topic_ != nil {
				if topicMsg := (<-topicMsgChan); topicMsg != tc.topicMsg {
					t.Fatalf("Expected '%s' got '%s'", tc.topicMsg, topicMsg)
				}
			}
		})
	}
}
