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
