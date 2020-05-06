package dsock_test

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"testing"
)

type ConnectSuite struct {
	suite.Suite
}

func TestConnectSuite(t *testing.T) {
	suite.Run(t, new(ConnectSuite))
}

func (suite *ConnectSuite) TestClaimConnect() {
	claim, err := createClaim(claimOptions{
		User:    "connect",
		Session: "claim",
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
		User:    "connect",
		Session: "claim",
	})
	if !checkRequestError(suite.Suite, err, info, "getting info") {
		return
	}

	connections := info["connections"].([]interface{})
	if !suite.Len(connections, 1, "Invalid number of connections") {
		return
	}

	claimData := claim["claim"].(map[string]interface{})
	connection := connections[0].(map[string]interface{})

	suite.Equal("connect", claimData["user"], "Invalid claim user")
	suite.Equal("connect", connection["user"], "Invalid connection user")

	suite.Equal("claim", claimData["session"], "Invalid claim user session")
	suite.Equal("claim", connection["session"], "Invalid connection user session")
}

func (suite *ConnectSuite) TestInvalidClaim() {
	_, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim=invalid-claim", nil)
	if !suite.Error(err, "Did not error when expected during connection") {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if !suite.NoError(err, "Could not read body") {
		return
	}

	var parsedBody map[string]interface{}

	err = json.Unmarshal(body, &parsedBody)
	if !suite.NoError(err, "Could not parse body") {
		return
	}

	if !suite.Equal(false, parsedBody["success"], "Succeeded when should have failed") {
		return
	}

	if !suite.Equal("MISSING_CLAIM", parsedBody["errorCode"], "Invalid error code") {
		return
	}
}

func (suite *ConnectSuite) TestJwtConnect() {
	// Hard coded JWT with max expiry:
	// {
	//  "sub": "connect",
	//  "sid": "jwt",
	//  "exp": 2147485546
	//}
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJjb25uZWN0Iiwic2lkIjoiand0IiwiZXhwIjoyMTQ3NDg1NTQ2fQ.oMbgPfg86I1sWs6IK25AP0H4ftzUVt9asKr9W9binW0"

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?jwt="+jwt, nil)
	if !checkConnectionError(suite.Suite, err, resp) {
		return
	}

	defer conn.Close()

	info, err := getInfo(infoOptions{
		User:    "connect",
		Session: "jwt",
	})
	if !checkRequestError(suite.Suite, err, info, "getting info") {
		return
	}

	connections := info["connections"].([]interface{})
	if !suite.Len(connections, 1, "Invalid number of connections") {
		return
	}

	connection := connections[0].(map[string]interface{})

	suite.Equal("connect", connection["user"], "Invalid connection user")
	suite.Equal("jwt", connection["session"], "Invalid connection user session")
}

func (suite *ConnectSuite) TestInvalidJwt() {
	// Hard coded JWT with invalid expiry:
	// {
	//  "sub": "connect",
	//  "sid": "invalid",
	//  "exp": "invalid"
	//}
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJjb25uZWN0Iiwic2lkIjoiaW52YWxpZCIsImV4cCI6ImludmFsaWQifQ.afZ4Mi-K0FeS35n7sivpNlq41JUi-QKVEjkH6mGWOrk"

	_, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?jwt="+jwt, nil)
	if !suite.Error(err, "Did not error when expecting during connection") {
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if !suite.NoError(err, "Could not read body") {
		return
	}

	var parsedBody map[string]interface{}

	err = json.Unmarshal(body, &parsedBody)
	if !suite.NoError(err, "Could not parse body") {
		return
	}

	if !suite.Equal(false, parsedBody["success"], "Application succeeded when expected to fail") {
		return
	}

	if !suite.Equal("INVALID_JWT", parsedBody["errorCode"], "Incorrect error code") {
		return
	}
}
