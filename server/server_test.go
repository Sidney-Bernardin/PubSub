package server

import (
	"fmt"
	"net"
	"testing"
)

func TestSubscribe(t *testing.T) {

	tt := []struct {
		request          []byte
		expectedTopics   []string
		expectedResponse readResponse
	}{
		{
			request:        []byte("sub a"),
			expectedTopics: []string{"a"},
			expectedResponse: readResponse{
				msg: "successfully subscribed to a",
			},
		},
		{
			request:        []byte("sub a b"),
			expectedTopics: []string{"a", "b"},
			expectedResponse: readResponse{
				msg: "successfully subscribed to a, b",
			},
		},
		{
			request:        []byte("su a"),
			expectedTopics: []string{},
			expectedResponse: readResponse{
				msg: `{"type":"invalid_command","detail":"Operation must be 'pub' or 'sub'."}`,
			},
		},
	}

	for _, tc := range tt {

		// Create mock network connection.
		serverConn, clientConn := net.Pipe()

		// Create a server.
		svr := NewServer(nil)
		go svr.handleConn(serverConn)

		// Read messages from the client's connection in another go-routine.
		readChan := make(chan readResponse)
		go svr.read(clientConn, readChan)

		// Write the test-case's request to the client's connection.
		if _, err := clientConn.Write(tc.request); err != nil {
			t.Fatalf("Cannot write to connection: %v", err)
		}

		// Wait for a response from the server.
		res := <-readChan

		// Check if the error is expected.
		if res.err != tc.expectedResponse.err {
			t.Fatalf("Unexpected read error - expected '%v' got '%v'", tc.expectedResponse.err, res.err)
		}

		// Check if the message is expected.
		if res.msg != tc.expectedResponse.msg {
			t.Fatalf("Unexpected read message - expected '%v' got '%v'", tc.expectedResponse.msg, res.msg)
		}

		// Loop over the test-case's expected topics to making sure they're all working.
		for _, topicName := range tc.expectedTopics {

			// Check if the topic exists.
			topic_, ok := svr.topics[topicName]
			if !ok {
				t.Fatalf("Topic %s wasn't created", topicName)
			}

			helloTopic := fmt.Sprintf("Hello, %s!", topicName)

			// Send test data to the topic's message-channel.
			topic_.msgChan <- []byte(helloTopic)

			// Wait for a response from server.
			res := <-readChan
			if res.err != nil {
				t.Fatalf("Cannot read from connection: %v", res.err)
			}

			// Check if the message is expected.
			if res.msg != helloTopic {
				t.Fatalf("Unexpected read message - expected '%v' got '%v'", helloTopic, res.msg)
			}
		}
	}
}
