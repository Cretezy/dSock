package dsock_test

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ChannelSuite struct {
	suite.Suite
}

func TestChannelSuite(t *testing.T) {
	suite.Run(t, new(ChannelSuite))
}

func (suite *ChannelSuite) TestChannelSubscribe() {
	claim, err := createClaim(claimOptions{
		User:    "channel",
		Session: "subscribe",
	})
	if !checkRequestError(suite.Suite, err, claim, "claim creation") {
		return
	}

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim="+claim["claim"].(map[string]interface{})["id"].(string), nil)
	if !checkConnectionError(suite.Suite, err, resp) {
		return
	}

	defer conn.Close()

	subscribe, err := subscribeChannel(channelOptions{
		target: target{
			User:    "channel",
			Session: "subscribe",
		},
		Channel: "channel_subscribe",
	})
	if !checkRequestError(suite.Suite, err, subscribe, "subscribing channel") {
		return
	}

	info, err := getInfo(infoOptions{
		target: target{
			Channel: "channel_subscribe",
		},
	})
	if !checkRequestError(suite.Suite, err, info, "getting info") {
		return
	}

	connections := info["connections"].([]interface{})
	if !suite.Len(connections, 1, "Incorrect number of connections") {
		return
	}

	connection := connections[0].(map[string]interface{})

	if !suite.Equal("channel", connection["user"], "Incorrect connection user") {
		return
	}

	if !suite.Equal("subscribe", connection["session"], "Incorrect connection user session") {
		return
	}

	// Includes default_channels in info
	if !suite.Equal([]string{"global", "channel_subscribe"}, interfaceToStringSlice(connection["channels"]), "Incorrect connection channels") {
		return
	}
}

func (suite *ChannelSuite) TestChannelUnsubscribe() {
	claim, err := createClaim(claimOptions{
		User:     "channel",
		Session:  "unsubscribe",
		Channels: []string{"channel_unsubscribe"},
	})
	if !checkRequestError(suite.Suite, err, claim, "claim creation") {
		return
	}

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim="+claim["claim"].(map[string]interface{})["id"].(string), nil)
	if !checkConnectionError(suite.Suite, err, resp) {
		return
	}

	defer conn.Close()

	subscribe, err := unsubscribeChannel(channelOptions{
		target: target{
			User:    "channel",
			Session: "unsubscribe",
		},
		Channel: "channel_unsubscribe",
	})
	if !checkRequestError(suite.Suite, err, subscribe, "unsubscribing channel") {
		return
	}

	info, err := getInfo(infoOptions{
		target: target{
			User:    "channel",
			Session: "unsubscribe",
		},
	})
	if !checkRequestError(suite.Suite, err, info, "getting info") {
		return
	}

	connections := info["connections"].([]interface{})
	if !suite.Len(connections, 1, "Incorrect number of connections") {
		return
	}

	connection := connections[0].(map[string]interface{})

	if !suite.Equal("channel", connection["user"], "Incorrect connection user") {
		return
	}

	if !suite.Equal("unsubscribe", connection["session"], "Incorrect connection user session") {
		return
	}

	if !suite.Equal([]string{"global"}, interfaceToStringSlice(connection["channels"]), "Incorrect connection channels") {
		return
	}
}

func (suite *ChannelSuite) TestChannelClaim() {
	claim, err := createClaim(claimOptions{
		User:    "channel",
		Session: "subscribe_claim",
	})
	if !checkRequestError(suite.Suite, err, claim, "claim creation") {
		return
	}

	subscribe, err := subscribeChannel(channelOptions{
		target: target{
			User:    "channel",
			Session: "subscribe_claim",
		},
		Channel: "channel_subscribe_claim",
	})
	if !checkRequestError(suite.Suite, err, subscribe, "subscribing channel") {
		return
	}

	info1, err := getInfo(infoOptions{
		target: target{
			Channel: "channel_subscribe_claim",
		},
	})
	if !checkRequestError(suite.Suite, err, info1, "getting info 1") {
		return
	}

	infoClaims1 := info1["claims"].([]interface{})
	if !suite.Len(infoClaims1, 1, "Incorrect number of info claims") {
		return
	}

	infoClaim1 := infoClaims1[0].(map[string]interface{})

	if !suite.Equal("channel", infoClaim1["user"], "Incorrect info claim 1 user") {
		return
	}

	if !suite.Equal("subscribe_claim", infoClaim1["session"], "Incorrect info claim 1 user session") {
		return
	}

	if !suite.Equal([]string{"channel_subscribe_claim"}, interfaceToStringSlice(infoClaim1["channels"]), "Incorrect info claim 1 channels") {
		return
	}

	unsubscribe, err := unsubscribeChannel(channelOptions{
		target: target{
			User:    "channel",
			Session: "subscribe_claim",
		},
		Channel: "channel_subscribe_claim",
	})
	if !checkRequestError(suite.Suite, err, unsubscribe, "unsubscribing channel") {
		return
	}

	info2, err := getInfo(infoOptions{
		target: target{
			Channel: "channel_subscribe_claim",
		},
	})
	if !checkRequestError(suite.Suite, err, info2, "getting info 2") {
		return
	}

	infoClaims2 := info2["claims"].([]interface{})
	if !suite.Len(infoClaims2, 0, "Incorrect number of info claims") {
		return
	}
}
func (suite *ChannelSuite) TestChannelClaimIgnore() {
	claim, err := createClaim(claimOptions{
		User:    "channel",
		Session: "subscribe_claim_ignore",
	})
	if !checkRequestError(suite.Suite, err, claim, "claim creation") {
		return
	}

	subscribe, err := subscribeChannel(channelOptions{
		target: target{
			User:    "channel",
			Session: "subscribe_claim_ignore",
		},
		Channel:      "channel_subscribe_claim_ignore",
		IgnoreClaims: true,
	})
	if !checkRequestError(suite.Suite, err, subscribe, "subscribing channel") {
		return
	}

	info, err := getInfo(infoOptions{
		target: target{
			Channel: "channel_subscribe_claim_ignore",
		},
	})
	if !checkRequestError(suite.Suite, err, info, "getting info") {
		return
	}

	infoClaims := info["claims"].([]interface{})
	if !suite.Len(infoClaims, 0, "Incorrect number of info claims") {
		return
	}
}
