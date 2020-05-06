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
	if !suite.NoError(err, "Error during sending") {
		return
	}

	if !suite.Equalf(true, message["success"], "Application error during sending (%s)", message["errorCode"]) {
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
	if !suite.NoError(err, "Error during sending") {
		return
	}

	if !suite.Equalf(true, message["success"], "Application error during sending (%s)", message["errorCode"]) {
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
	if !suite.NoError(err, "Error during sending") {
		return
	}

	if !suite.Equalf(true, message["success"], "Application error during sending (%s)", message["errorCode"]) {
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
	if !suite.NoError(err, "Error during sending") {
		return
	}

	if !suite.Equalf(false, message["success"], "Application succeeded when expected to fail") {
		return
	}

	if !suite.Equalf("MISSING_CONNECTION_OR_USER", message["errorCode"], "Incorrect error code") {
		return
	}
}

func (suite *SendSuite) TestSendNoType() {
	message, err := sendMessage(sendOptions{
		Id: "a",
	})
	if !suite.NoError(err, "Error during sending") {
		return
	}

	if !suite.Equalf(false, message["success"], "Application succeeded when expected to fail") {
		return
	}

	if !suite.Equalf("INVALID_MESSAGE_TYPE", message["errorCode"], "Incorrect error code") {
		return
	}
}
