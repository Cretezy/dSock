package dsock_test

import (
	dsock "github.com/Cretezy/dSock-go"
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
	claim, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User: "send_user",
	})
	if !checkRequestError(suite.Suite, err, "claim creation") {
		return
	}

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim="+claim.Id, nil)
	if !checkConnectionError(suite.Suite, err, resp) {
		return
	}

	defer conn.Close()

	err = dSockClient.SendMessage(dsock.SendMessageOptions{
		Target: dsock.Target{
			User: "send_user",
		},
		Type:    "text",
		Message: []byte("Hello world!"),
	})
	if !checkRequestError(suite.Suite, err, "sending") {
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
	claim, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User:    "send",
		Session: "session",
	})
	if !checkRequestError(suite.Suite, err, "claim creation") {
		return
	}

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim="+claim.Id, nil)
	if !checkConnectionError(suite.Suite, err, resp) {
		return
	}

	defer conn.Close()

	err = dSockClient.SendMessage(dsock.SendMessageOptions{
		Target: dsock.Target{
			User:    "send",
			Session: "session",
		},
		Type:    "text",
		Message: []byte("Hello world!"),
	})
	if !checkRequestError(suite.Suite, err, "sending") {
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
	claim, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User:    "send",
		Session: "connection",
	})
	if !checkRequestError(suite.Suite, err, "claim creation") {
		return
	}

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim="+claim.Id, nil)
	if !checkConnectionError(suite.Suite, err, resp) {
		return
	}

	defer conn.Close()

	info, err := dSockClient.GetInfo(dsock.GetInfoOptions{
		Target: dsock.Target{
			User:    "send",
			Session: "connection",
		},
	})
	if !checkRequestError(suite.Suite, err, "getting info") {
		return
	}

	infoConnections := info.Connections
	if !suite.Len(infoConnections, 1, "Invalid number of connections") {
		return
	}

	err = dSockClient.SendMessage(dsock.SendMessageOptions{
		Target: dsock.Target{
			Id: infoConnections[0].Id,
		},
		Type:    "binary",
		Message: []byte{1, 2, 3, 4},
	})
	if !checkRequestError(suite.Suite, err, "sending") {
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
	err := dSockClient.SendMessage(dsock.SendMessageOptions{})
	if !checkRequestNoError(suite.Suite, err, "MISSING_TARGET", "sending") {
		return
	}
}

func (suite *SendSuite) TestSendNoType() {
	err := dSockClient.SendMessage(dsock.SendMessageOptions{
		Target: dsock.Target{
			Id: "a",
		},
	})
	if !checkRequestNoError(suite.Suite, err, "INVALID_MESSAGE_TYPE", "sending") {
		return
	}
}

func (suite *SendSuite) TestConnectionChannel() {
	claim, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User:     "send",
		Session:  "channel",
		Channels: []string{"send_channel"},
	})
	if !checkRequestError(suite.Suite, err, "claim creation") {
		return
	}

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim="+claim.Id, nil)
	if !checkConnectionError(suite.Suite, err, resp) {
		return
	}

	defer conn.Close()

	err = dSockClient.SendMessage(dsock.SendMessageOptions{
		Target: dsock.Target{
			Channel: "send_channel",
		},
		Type:    "text",
		Message: []byte("Hello world!"),
	})
	if !checkRequestError(suite.Suite, err, "sending") {
		return
	}

	messageType, data, err := conn.ReadMessage()
	if !suite.NoError(err, "Error during receiving message") {
		return
	}

	if !suite.Equal(websocket.TextMessage, messageType, "Incorrect message type") {
		return
	}

	if !suite.Equal([]byte("Hello world!"), data, "Incorrect message data") {
		return
	}
}
