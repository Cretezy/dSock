package dsock_test

import (
	dsock "github.com/Cretezy/dSock-go"
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
	claim, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User: "disconnect_user",
	})
	if !checkRequestError(suite.Suite, err, "claim creation") {
		return
	}

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim="+claim.Id, nil)
	if !checkConnectionError(suite.Suite, err, resp) {
		return
	}

	defer conn.Close()

	err = dSockClient.Disconnect(dsock.DisconnectOptions{
		Target: dsock.Target{
			User: "disconnect_user",
		},
	})
	if !checkRequestError(suite.Suite, err, "disconnection") {
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
	claim, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User:    "disconnect",
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

	err = dSockClient.Disconnect(dsock.DisconnectOptions{
		Target: dsock.Target{
			User:    "disconnect",
			Session: "session",
		},
	})
	if !checkRequestError(suite.Suite, err, "disconnection") {
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
	claim, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User:    "disconnect",
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
			User:    "disconnect",
			Session: "connection",
		},
	})
	if !checkRequestError(suite.Suite, err, "getting info") {
		return
	}

	infoConnections := info.Connections
	if !suite.Len(infoConnections, 1, "Incorrect number of connections") {
		return
	}

	id := infoConnections[0].Id

	err = dSockClient.Disconnect(dsock.DisconnectOptions{
		Target: dsock.Target{
			Id: id,
		},
	})
	if !checkRequestError(suite.Suite, err, "disconnection") {
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
	claim, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User:    "disconnect",
		Session: "no_session",
	})
	if !checkRequestError(suite.Suite, err, "claim creation") {
		return
	}

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim="+claim.Id, nil)
	if !checkConnectionError(suite.Suite, err, resp) {
		return
	}

	defer conn.Close()

	var websocketErr error

	go func() {
		_, _, websocketErr = conn.ReadMessage()
	}()

	err = dSockClient.Disconnect(dsock.DisconnectOptions{
		Target: dsock.Target{
			User:    "disconnect",
			Session: "no_session_bad",
		},
	})
	if !checkRequestError(suite.Suite, err, "disconnection") {
		return
	}

	// Wait a bit to see if connection was closed
	time.Sleep(time.Millisecond * 100)

	if !suite.NoError(websocketErr, "Got Websocket error (disconnect) when wasn't expecting") {
		return
	}
}

func (suite *DisconnectSuite) TestDisconnectExpireClaim() {
	_, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User:    "disconnect",
		Session: "expire_claim",
	})
	if !checkRequestError(suite.Suite, err, "claim creation") {
		return
	}

	infoBefore, err := dSockClient.GetInfo(dsock.GetInfoOptions{
		Target: dsock.Target{
			User:    "disconnect",
			Session: "expire_claim",
		},
	})
	if !checkRequestError(suite.Suite, err, "getting before info") {
		return
	}

	if !suite.Len(infoBefore.Claims, 1, "Incorrect number of claims before") {
		return
	}

	err = dSockClient.Disconnect(dsock.DisconnectOptions{
		Target: dsock.Target{
			User:    "disconnect",
			Session: "expire_claim",
		},
	})
	if !checkRequestError(suite.Suite, err, "disconnection") {
		return
	}

	infoAfter, err := dSockClient.GetInfo(dsock.GetInfoOptions{
		Target: dsock.Target{
			User:    "disconnect",
			Session: "expire_claim",
		},
	})
	if !checkRequestError(suite.Suite, err, "getting after info") {
		return
	}

	if !suite.Len(infoAfter.Claims, 0, "Incorrect number of claims after") {
		return
	}
}

func (suite *DisconnectSuite) TestDisconnectKeepClaim() {
	_, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User:    "disconnect",
		Session: "keep_claim",
	})
	if !checkRequestError(suite.Suite, err, "claim creation") {
		return
	}
	infoBefore, err := dSockClient.GetInfo(dsock.GetInfoOptions{
		Target: dsock.Target{
			User:    "disconnect",
			Session: "keep_claim",
		},
	})
	if !checkRequestError(suite.Suite, err, "getting info") {
		return
	}

	if !suite.Len(infoBefore.Claims, 1, "Incorrect number of claims before") {
		return
	}

	err = dSockClient.Disconnect(dsock.DisconnectOptions{
		Target: dsock.Target{
			User:    "disconnect",
			Session: "keep_claim_invalid",
		},
		KeepClaims: true,
	})
	if !checkRequestError(suite.Suite, err, "disconnection") {
		return
	}

	infoAfter, err := dSockClient.GetInfo(dsock.GetInfoOptions{
		Target: dsock.Target{
			User:    "disconnect",
			Session: "keep_claim",
		},
	})
	if !checkRequestError(suite.Suite, err, "getting after info") {
		return
	}

	if !suite.Len(infoAfter.Claims, 1, "Incorrect number of claims after") {
		return
	}
}

func (suite *DisconnectSuite) TestChannelDisconnect() {
	claim, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User:     "disconnect",
		Session:  "channel",
		Channels: []string{"disconnect_channel"},
	})
	if !checkRequestError(suite.Suite, err, "claim creation") {
		return
	}

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim="+claim.Id, nil)
	if !checkConnectionError(suite.Suite, err, resp) {
		return
	}

	defer conn.Close()

	err = dSockClient.Disconnect(dsock.DisconnectOptions{
		Target: dsock.Target{
			Channel: "disconnect_channel",
		},
	})
	if !checkRequestError(suite.Suite, err, "disconnection") {
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
