package dsock_test

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"
	"testing"
)

type InfoSuite struct {
	suite.Suite
}

func TestInfoSuite(t *testing.T) {
	suite.Run(t, new(InfoSuite))
}

func (suite *InfoSuite) TestInfoClaim() {
	claim, err := createClaim(claimOptions{
		User:       "info",
		Session:    "claim",
		Expiration: 2147485545,
	})
	if !checkRequestError(suite.Suite, err, claim, "claim creation") {
		return
	}

	info, err := getInfo(infoOptions{
		User:    "info",
		Session: "claim",
	})
	if !checkRequestError(suite.Suite, err, info, "getting info") {
		return
	}

	infoClaims := info["claims"].([]interface{})

	if !suite.Len(infoClaims, 1, "Invalid number of claims") {
		return
	}

	claimData := claim["claim"].(map[string]interface{})

	infoClaimData := infoClaims[0].(map[string]interface{})

	if !suite.Equal(claimData["id"], infoClaimData["id"], "Info claim ID doesn't match claim") {
		return
	}

	if !suite.Equal("info", claimData["user"], "Invalid claim user") {
		return
	}
	if !suite.Equal("info", infoClaimData["user"], "Invalid info claim user") {
		return
	}

	if !suite.Equal("claim", claimData["session"], "Invalid claim user session") {
		return
	}
	if !suite.Equal("claim", infoClaimData["session"], "Invalid info claim user session") {
		return
	}

	// Has to do some weird casting
	if !suite.Equal(2147485545, int(infoClaimData["expiration"].(float64)), "Info claim expiration doesn't match") {
		return
	}
}

func (suite *InfoSuite) TestInfoClaimInvalidExpiration() {
	claim, err := createClaim(claimOptions{
		User:       "info",
		Session:    "invalid_expiration",
		Expiration: 1,
	})
	if !checkRequestNoError(suite.Suite, err, claim, "INVALID_EXPIRATION", "claim creation") {
		return
	}
}

func (suite *InfoSuite) TestInfoClaimNegativeExpiration() {
	claim, err := createClaim(claimOptions{
		User:       "info",
		Session:    "negative_expiration",
		Expiration: -1,
	})
	if !checkRequestNoError(suite.Suite, err, claim, "NEGATIVE_EXPIRATION", "claim creation") {
		return
	}
}

func (suite *InfoSuite) TestInfoClaimNegativeDuration() {
	claim, err := createClaim(claimOptions{
		User:     "info",
		Session:  "negative_duration",
		Duration: -1,
	})
	if !checkRequestNoError(suite.Suite, err, claim, "NEGATIVE_DURATION", "claim creation") {
		return
	}
}

func (suite *InfoSuite) TestInfoConnection() {
	claim, err := createClaim(claimOptions{
		User:    "info",
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
		User:    "info",
		Session: "connection",
	})
	if !checkRequestError(suite.Suite, err, info, "getting info") {
		return
	}

	infoConnections := info["connections"].([]interface{})
	if !suite.Len(infoConnections, 1, "Invalid number of connections") {
		return
	}

	claimData := claim["claim"].(map[string]interface{})
	infoConnectionData := infoConnections[0].(map[string]interface{})

	suite.Equal("info", claimData["user"], "Invalid claim user")
	suite.Equal("info", infoConnectionData["user"], "Invalid connection user")

	suite.Equal("connection", claimData["session"], "Invalid claim user session")
	suite.Equal("connection", infoConnectionData["session"], "Invalid connection user session")
}

func (suite *InfoSuite) TestInfoMissing() {
	info, err := getInfo(infoOptions{
		User:    "info",
		Session: "missing",
	})
	if !checkRequestError(suite.Suite, err, info, "getting info") {
		return
	}

	infoClaims := info["claims"].([]interface{})
	if !suite.Len(infoClaims, 0, "Invalid number of claims") {
		return
	}

	infoConnections := info["connections"].([]interface{})
	if !suite.Len(infoConnections, 0, "Invalid number of connections") {
		return
	}
}

func (suite *InfoSuite) TestInfoNoTarget() {
	info, err := getInfo(infoOptions{})
	if !checkRequestNoError(suite.Suite, err, info, "MISSING_CONNECTION_OR_USER", "getting info") {
		return
	}
}
