package dsock_test

import (
	dsock "github.com/Cretezy/dSock-go"
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
	claim, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User:       "info",
		Session:    "claim",
		Expiration: 2147485545,
	})
	if !checkRequestError(suite.Suite, err, "claim creation") {
		return
	}

	info, err := dSockClient.GetInfo(dsock.GetInfoOptions{
		Target: dsock.Target{
			User:    "info",
			Session: "claim",
		},
	})
	if !checkRequestError(suite.Suite, err, "getting info") {
		return
	}

	infoClaims := info.Claims

	if !suite.Len(infoClaims, 1, "Incorrect number of claims") {
		return
	}

	infoClaim := infoClaims[0]

	if !suite.Equal(claim.Id, infoClaim.Id, "Info claim ID doesn't match claim") {
		return
	}

	if !suite.Equal("info", claim.User, "Incorrect claim user") {
		return
	}
	if !suite.Equal("info", infoClaim.User, "Incorrect info claim user") {
		return
	}

	if !suite.Equal("claim", claim.Session, "Incorrect claim user session") {
		return
	}
	if !suite.Equal("claim", infoClaim.Session, "Incorrect info claim user session") {
		return
	}

	// Has to do some weird casting
	if !suite.Equal(2147485545, infoClaim.Expiration, "Info claim expiration doesn't match") {
		return
	}
}

func (suite *InfoSuite) TestInfoClaimInvalidExpiration() {
	_, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User:       "info",
		Session:    "invalid_expiration",
		Expiration: 1,
	})
	if !checkRequestNoError(suite.Suite, err, "INVALID_EXPIRATION", "claim creation") {
		return
	}
}

func (suite *InfoSuite) TestInfoClaimNegativeExpiration() {
	_, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User:       "info",
		Session:    "negative_expiration",
		Expiration: -1,
	})
	if !checkRequestNoError(suite.Suite, err, "NEGATIVE_EXPIRATION", "claim creation") {
		return
	}
}

func (suite *InfoSuite) TestInfoClaimNegativeDuration() {
	_, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User:     "info",
		Session:  "negative_duration",
		Duration: -1,
	})
	if !checkRequestNoError(suite.Suite, err, "NEGATIVE_DURATION", "claim creation") {
		return
	}
}

func (suite *InfoSuite) TestInfoConnection() {
	claim, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User:    "info",
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
			User:    "info",
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

	infoConnection := infoConnections[0]

	if !suite.Equal("info", claim.User, "Incorrect claim user") {
		return
	}
	if !suite.Equal("info", infoConnection.User, "Incorrect connection user") {
		return
	}

	if !suite.Equal("connection", claim.Session, "Incorrect claim user session") {
		return
	}
	if !suite.Equal("connection", infoConnection.Session, "Incorrect connection user session") {
		return
	}
}

func (suite *InfoSuite) TestInfoMissing() {
	info, err := dSockClient.GetInfo(dsock.GetInfoOptions{
		Target: dsock.Target{
			User:    "info",
			Session: "missing",
		},
	})
	if !checkRequestError(suite.Suite, err, "getting info") {
		return
	}

	if !suite.Len(info.Claims, 0, "Incorrect number of claims") {
		return
	}

	if !suite.Len(info.Connections, 0, "Incorrect number of connections") {
		return
	}
}

func (suite *InfoSuite) TestInfoNoTarget() {
	_, err := dSockClient.GetInfo(dsock.GetInfoOptions{})
	if !checkRequestNoError(suite.Suite, err, "MISSING_TARGET", "getting info") {
		return
	}
}

func (suite *InfoSuite) TestInfoChannel() {
	claim, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User:     "info",
		Session:  "channel",
		Channels: []string{"info_channel"},
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
			Channel: "info_channel",
		},
	})
	if !checkRequestError(suite.Suite, err, "getting info") {
		return
	}

	infoConnections := info.Connections
	if !suite.Len(infoConnections, 1, "Incorrect number of connections") {
		return
	}

	infoConnection := infoConnections[0]

	if !suite.Equal("info", claim.User, "Incorrect claim user") {
		return
	}
	if !suite.Equal("info", infoConnection.User, "Incorrect connection user") {
		return
	}

	if !suite.Equal([]string{"info_channel"}, interfaceToStringSlice(claim.Channels), "Incorrect claim channels") {
		return
	}

	// Includes default_channels in info
	if !suite.Equal([]string{"info_channel", "global"}, interfaceToStringSlice(infoConnection.Channels), "Incorrect connection channels") {
		return
	}
}

func (suite *InfoSuite) TestInfoChannelClaim() {
	claim, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User:     "info",
		Session:  "channel_claim",
		Channels: []string{"info_channel_claim"},
	})
	if !checkRequestError(suite.Suite, err, "claim creation") {
		return
	}

	info, err := dSockClient.GetInfo(dsock.GetInfoOptions{
		Target: dsock.Target{
			Channel: "info_channel_claim",
		},
	})
	if !checkRequestError(suite.Suite, err, "getting info") {
		return
	}

	infoClaims := info.Claims
	if !suite.Len(infoClaims, 1, "Incorrect number of claims") {
		return
	}

	infoClaim := infoClaims[0]

	if !suite.Equal("info", claim.User, "Incorrect claim user") {
		return
	}
	if !suite.Equal("info", infoClaim.User, "Incorrect info claim user") {
		return
	}

	if !suite.Equal([]string{"info_channel_claim"}, interfaceToStringSlice(claim.Channels), "Incorrect claim channels") {
		return
	}

	// Includes default_channels in info
	if !suite.Equal([]string{"info_channel_claim"}, interfaceToStringSlice(infoClaim.Channels), "Incorrect info claim channels") {
		return
	}
}
