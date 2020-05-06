package dsock_test

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type DisconnectSuite struct {
	suite.Suite
}

func TestDisconnectSuite(t *testing.T) {
	suite.Run(t, new(DisconnectSuite))
}

func (suite *DisconnectSuite) TestUserDisconnect() {
	claim, err := createClaim(claimOptions{
		User: "disconnect_user",
	})
	if !checkRequestError(suite.Suite, err, claim, "claim creation") {
		return
	}

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim="+claim["claim"].(map[string]interface{})["id"].(string), nil)
	if !checkConnectionError(suite.Suite, err, resp) {
		return
	}

	defer conn.Close()

	disconnectResponse, err := disconnect(disconnectOptions{
		User: "disconnect_user",
	})
	if !checkRequestError(suite.Suite, err, disconnectResponse, "disconnection") {
		return
	}

	_, _, err = conn.ReadMessage()
	if !suite.Error(err, "Didn't get error when was expecting") {
		return
	}

	if closeErr, ok := err.(*websocket.CloseError); ok {
		if !suite.Equal(websocket.CloseNormalClosure, closeErr.Code, "Incorrect close type") {
			return
		}
	} else {
		suite.Failf("Incorrect error type: %s", err.Error())
	}
}

func (suite *DisconnectSuite) TestSessionDisconnect() {
	claim, err := createClaim(claimOptions{
		User:    "disconnect",
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

	disconnectResponse, err := disconnect(disconnectOptions{
		User:    "disconnect",
		Session: "session",
	})
	if !checkRequestError(suite.Suite, err, disconnectResponse, "disconnection") {
		return
	}

	_, _, err = conn.ReadMessage()
	if !suite.Error(err, "Didn't get error when was expecting") {
		return
	}

	if closeErr, ok := err.(*websocket.CloseError); ok {
		if !suite.Equal(websocket.CloseNormalClosure, closeErr.Code, "Incorrect close type") {
			return
		}
	} else {
		suite.Failf("Incorrect error type: %s", err.Error())
	}
}

func (suite *DisconnectSuite) TestConnectionDisconnect() {
	claim, err := createClaim(claimOptions{
		User:    "disconnect",
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
		User:    "disconnect",
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

	disconnectResponse, err := disconnect(disconnectOptions{
		Id: id,
	})
	if !checkRequestError(suite.Suite, err, disconnectResponse, "disconnection") {
		return
	}

	_, _, err = conn.ReadMessage()
	if !suite.Error(err, "Didn't get error when was expecting") {
		return
	}

	if closeErr, ok := err.(*websocket.CloseError); ok {
		if !suite.Equal(websocket.CloseNormalClosure, closeErr.Code, "Incorrect close type") {
			return
		}
	} else {
		suite.Failf("Incorrect error type: %s", err.Error())
	}
}

func (suite *DisconnectSuite) TestNoSessionDisconnect() {
	claim, err := createClaim(claimOptions{
		User:    "disconnect",
		Session: "no_session",
	})
	if !checkRequestError(suite.Suite, err, claim, "claim creation") {
		return
	}

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim="+claim["claim"].(map[string]interface{})["id"].(string), nil)
	if !checkConnectionError(suite.Suite, err, resp) {
		return
	}

	defer conn.Close()

	var websocketErr error

	go func() {
		_, _, websocketErr = conn.ReadMessage()
	}()

	disconnectResponse, err := disconnect(disconnectOptions{
		User:    "disconnect",
		Session: "no_session_bad",
	})
	if !checkRequestError(suite.Suite, err, disconnectResponse, "disconnection") {
		return
	}

	// Wait a bit to see if connection was closed
	time.Sleep(time.Millisecond * 100)

	if !suite.NoError(websocketErr, "Got Websocket error (disconnect) when wasn't expecting") {
		return
	}
}

func (suite *DisconnectSuite) TestDisconnectExpireClaim() {
	claim, err := createClaim(claimOptions{
		User:    "disconnect",
		Session: "expire_claim",
	})
	if !checkRequestError(suite.Suite, err, claim, "claim creation") {
		return
	}

	infoBefore, err := getInfo(infoOptions{
		User:    "disconnect",
		Session: "expire_claim",
	})
	if !checkRequestError(suite.Suite, err, infoBefore, "getting before info") {
		return
	}

	claimsBefore := infoBefore["claims"].([]interface{})

	if !suite.Len(claimsBefore, 1, "Incorrect number of claims before") {
		return
	}

	disconnectResponse, err := disconnect(disconnectOptions{
		User:    "disconnect",
		Session: "expire_claim",
	})
	if !checkRequestError(suite.Suite, err, disconnectResponse, "disconnection") {
		return
	}

	infoAfter, err := getInfo(infoOptions{
		User:    "disconnect",
		Session: "expire_claim",
	})
	if !checkRequestError(suite.Suite, err, infoBefore, "getting after info") {
		return
	}

	claimsAfter := infoAfter["claims"].([]interface{})

	if !suite.Len(claimsAfter, 0, "Incorrect number of claims after") {
		return
	}
}

func (suite *DisconnectSuite) TestDisconnectKeepClaim() {
	claim, err := createClaim(claimOptions{
		User:    "disconnect",
		Session: "keep_claim",
	})
	if !checkRequestError(suite.Suite, err, claim, "claim creation") {
		return
	}
	infoBefore, err := getInfo(infoOptions{
		User:    "disconnect",
		Session: "keep_claim",
	})
	if !checkRequestError(suite.Suite, err, infoBefore, "getting info") {
		return
	}

	claimsBefore := infoBefore["claims"].([]interface{})

	if !suite.Len(claimsBefore, 1, "Incorrect number of claims before") {
		return
	}

	disconnectResponse, err := disconnect(disconnectOptions{
		User:       "disconnect",
		Session:    "keep_claim_invalid",
		KeepClaims: true,
	})
	if !checkRequestError(suite.Suite, err, disconnectResponse, "disconnection") {
		return
	}

	infoAfter, err := getInfo(infoOptions{
		User:    "disconnect",
		Session: "keep_claim",
	})
	if !checkRequestError(suite.Suite, err, infoAfter, "getting after info") {
		return
	}

	claimsAfter := infoAfter["claims"].([]interface{})

	if !suite.Len(claimsAfter, 1, "Incorrect number of claims after") {
		return
	}
}
