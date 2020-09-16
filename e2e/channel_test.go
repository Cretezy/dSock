package dsock_test

import (
	"github.com/Cretezy/dSock-go"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type ChannelSuite struct {
	suite.Suite
}

func TestChannelSuite(t *testing.T) {
	suite.Run(t, new(ChannelSuite))
}

func (suite *ChannelSuite) TestChannelSubscribe() {
	claim, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User:    "channel",
		Session: "subscribe",
	})
	if !checkRequestError(suite.Suite, err, "claim creation") {
		return
	}

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim="+claim.Id, nil)
	if !checkConnectionError(suite.Suite, err, resp) {
		return
	}

	defer conn.Close()

	err = dSockClient.SubscribeChannel(dsock.ChannelOptions{
		Target: dsock.Target{
			User:    "channel",
			Session: "subscribe",
		},
		Channel: "channel_subscribe",
	})
	if !checkRequestError(suite.Suite, err, "subscribing channel") {
		return
	}

	// Give it some time to propagate
	time.Sleep(time.Millisecond * 10)

	info, err := dSockClient.GetInfo(dsock.GetInfoOptions{
		Target: dsock.Target{
			Channel: "channel_subscribe",
		},
	})
	if !checkRequestError(suite.Suite, err, "getting info") {
		return
	}

	if !suite.Len(info.Connections, 1, "Incorrect number of connections") {
		return
	}

	connection := info.Connections[0]

	if !suite.Equal("channel", connection.User, "Incorrect connection user") {
		return
	}

	if !suite.Equal("subscribe", connection.Session, "Incorrect connection user session") {
		return
	}

	// Includes default_channels in info
	if !suite.Equal([]string{"global", "channel_subscribe"}, interfaceToStringSlice(connection.Channels), "Incorrect connection channels") {
		return
	}
}

func (suite *ChannelSuite) TestChannelUnsubscribe() {
	claim, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User:     "channel",
		Session:  "unsubscribe",
		Channels: []string{"channel_unsubscribe"},
	})
	if !checkRequestError(suite.Suite, err, "claim creation") {
		return
	}

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim="+claim.Id, nil)
	if !checkConnectionError(suite.Suite, err, resp) {
		return
	}

	defer conn.Close()

	err = dSockClient.UnsubscribeChannel(dsock.ChannelOptions{
		Target: dsock.Target{
			User:    "channel",
			Session: "unsubscribe",
		},
		Channel: "channel_unsubscribe",
	})
	if !checkRequestError(suite.Suite, err, "unsubscribing channel") {
		return
	}

	info, err := dSockClient.GetInfo(dsock.GetInfoOptions{
		Target: dsock.Target{
			User:    "channel",
			Session: "unsubscribe",
		},
	})
	if !checkRequestError(suite.Suite, err, "getting info") {
		return
	}

	if !suite.Len(info.Connections, 1, "Incorrect number of connections") {
		return
	}

	connection := info.Connections[0]

	if !suite.Equal("channel", connection.User, "Incorrect connection user") {
		return
	}

	if !suite.Equal("unsubscribe", connection.Session, "Incorrect connection user session") {
		return
	}

	if !suite.Equal([]string{"global"}, interfaceToStringSlice(connection.Channels), "Incorrect connection channels") {
		return
	}
}

func (suite *ChannelSuite) TestChannelClaim() {
	_, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User:    "channel",
		Session: "subscribe_claim",
	})
	if !checkRequestError(suite.Suite, err, "claim creation") {
		return
	}

	err = dSockClient.SubscribeChannel(dsock.ChannelOptions{
		Target: dsock.Target{
			User:    "channel",
			Session: "subscribe_claim",
		},
		Channel: "channel_subscribe_claim",
	})
	if !checkRequestError(suite.Suite, err, "subscribing channel") {
		return
	}

	info1, err := dSockClient.GetInfo(dsock.GetInfoOptions{
		Target: dsock.Target{
			Channel: "channel_subscribe_claim",
		},
	})
	if !checkRequestError(suite.Suite, err, "getting info 1") {
		return
	}

	infoClaims1 := info1.Claims
	if !suite.Len(infoClaims1, 1, "Incorrect number of info claims") {
		return
	}

	infoClaim1 := infoClaims1[0]

	if !suite.Equal("channel", infoClaim1.User, "Incorrect info claim 1 user") {
		return
	}

	if !suite.Equal("subscribe_claim", infoClaim1.Session, "Incorrect info claim 1 user session") {
		return
	}

	if !suite.Equal([]string{"channel_subscribe_claim"}, interfaceToStringSlice(infoClaim1.Channels), "Incorrect info claim 1 channels") {
		return
	}

	err = dSockClient.UnsubscribeChannel(dsock.ChannelOptions{
		Target: dsock.Target{
			User:    "channel",
			Session: "subscribe_claim",
		},
		Channel: "channel_subscribe_claim",
	})
	if !checkRequestError(suite.Suite, err, "unsubscribing channel") {
		return
	}

	info2, err := dSockClient.GetInfo(dsock.GetInfoOptions{
		Target: dsock.Target{
			Channel: "channel_subscribe_claim",
		},
	})
	if !checkRequestError(suite.Suite, err, "getting info 2") {
		return
	}

	if !suite.Len(info2.Claims, 0, "Incorrect number of info claims") {
		return
	}
}
func (suite *ChannelSuite) TestChannelClaimIgnore() {
	_, err := dSockClient.CreateClaim(dsock.CreateClaimOptions{
		User:    "channel",
		Session: "subscribe_claim_ignore",
	})
	if !checkRequestError(suite.Suite, err, "claim creation") {
		return
	}

	err = dSockClient.SubscribeChannel(dsock.ChannelOptions{
		Target: dsock.Target{
			User:    "channel",
			Session: "subscribe_claim_ignore",
		},
		Channel:      "channel_subscribe_claim_ignore",
		IgnoreClaims: true,
	})
	if !checkRequestError(suite.Suite, err, "subscribing channel") {
		return
	}

	info, err := dSockClient.GetInfo(dsock.GetInfoOptions{
		Target: dsock.Target{
			Channel: "channel_subscribe_claim_ignore",
		},
	})
	if !checkRequestError(suite.Suite, err, "getting info") {
		return
	}

	if !suite.Len(info.Claims, 0, "Incorrect number of info claims") {
		return
	}
}
