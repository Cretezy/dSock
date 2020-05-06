package dsock_test

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"
	"testing"
)

type SendSuite struct {
	suite.Suite
}

func TestSendSuite(t *testing.T) {
	suite.Run(t, new(SendSuite))
}

func (suite *SendSuite) TestUserSend() {
	claim, err := createClaim(claimOptions{
		User: "send_user",
	})
	if !checkRequestError(suite.Suite, err, claim, "claim creation") {
		return
	}

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim="+claim["claim"].(map[string]interface{})["id"].(string), nil)
	if !checkConnectionError(suite.Suite, err, resp) {
		return
	}

	defer conn.Close()

	message, err := sendMessage(sendOptions{
		User:    "send_user",
		Type:    "text",
		Message: []byte("Hello world!"),
	})
	if !checkRequestError(suite.Suite, err, message, "sending") {
		return
	}

	messageType, data, err := conn.ReadMessage()
	if !suite.NoError(err, "Error during receiving message") {
		return
	}

	if !suite.Equal(websocket.TextMessage, messageType, "Incorrect message type") {
		return
	}

	if !suite.Equal("Hello world!", string(data), "Incorrect message data") {
		return
	}
}

func (suite *SendSuite) TestUserSessionSend() {
	claim, err := createClaim(claimOptions{
		User:    "send",
		Session: "session",
	})
	if !checkRequestError(suite.Suite, err, claim, "claim creation") {
		return
	}

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim="+claim["claim"].(map[string]interface{})["id"].(string), nil)
	if !checkConnectionError(suite.Suite, err, resp) {
		return
	}

	defer conn.Close()

	message, err := sendMessage(sendOptions{
		User:    "send",
		Session: "session",
		Type:    "text",
		Message: []byte("Hello world!"),
	})
	if !checkRequestError(suite.Suite, err, message, "sending") {
		return
	}

	messageType, data, err := conn.ReadMessage()
	if !suite.NoError(err, "Error during receiving message") {
		return
	}

	if !suite.Equal(websocket.TextMessage, messageType, "Incorrect message type") {
		return
	}

	if !suite.Equal("Hello world!", string(data), "Incorrect message data") {
		return
	}
}

func (suite *SendSuite) TestConnectionSend() {
	claim, err := createClaim(claimOptions{
		User:    "send",
		Session: "connection",
	})
	if !checkRequestError(suite.Suite, err, claim, "claim creation") {
		return
	}

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim="+claim["claim"].(map[string]interface{})["id"].(string), nil)
	if !checkConnectionError(suite.Suite, err, resp) {
		return
	}

	defer conn.Close()

	info, err := getInfo(infoOptions{
		User:    "send",
		Session: "connection",
	})
	if !checkRequestError(suite.Suite, err, info, "getting info") {
		return
	}

	infoConnections := info["connections"].([]interface{})
	if !suite.Len(infoConnections, 1, "Invalid number of connections") {
		return
	}

	id := infoConnections[0].(map[string]interface{})["id"].(string)

	message, err := sendMessage(sendOptions{
		Id:      id,
		Type:    "binary",
		Message: []byte{1, 2, 3, 4},
	})
	if !checkRequestError(suite.Suite, err, message, "sending") {
		return
	}

	messageType, data, err := conn.ReadMessage()
	if !suite.NoError(err, "Error during receiving message") {
		return
	}

	if !suite.Equal(websocket.BinaryMessage, messageType, "Incorrect message type") {
		return
	}

	if !suite.Equal([]byte{1, 2, 3, 4}, data, "Incorrect message data") {
		return
	}
}

func (suite *SendSuite) TestSendNoTarget() {
	message, err := sendMessage(sendOptions{})
	if !checkRequestNoError(suite.Suite, err, message, "MISSING_CONNECTION_OR_USER", "sending") {
		return
	}
}

func (suite *SendSuite) TestSendNoType() {
	message, err := sendMessage(sendOptions{
		Id: "a",
	})
	if !checkRequestNoError(suite.Suite, err, message, "INVALID_MESSAGE_TYPE", "sending") {
		return
	}
}
