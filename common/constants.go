package common

import "time"

const (
	PathPing                  = "/ping"
	PathSend                  = "/send"
	PathConnect               = "/connect"
	PathClaim                 = "/claim"
	PathInfo                  = "/info"
	PathDisconnect            = "/disconnect"
	PathChannelSubscribe      = "/channel/subscribe/:channel"
	PathChannelUnsubscribe    = "/channel/unsubscribe/:channel"
	PathReceiveMessage        = "/_/message"
	PathReceiveChannelMessage = "/_/message/channel"
)

const ProtobufContentType = "application/protobuf"

const TtlBuffer = 5 * time.Second
