package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/Cretezy/dSock/common/protos"
	"strings"
)

func handleChannel(channelAction *protos.ChannelAction) {
	// Resolve all local connections for message target
	connections, ok := resolveConnections(common.ResolveOptions{
		Connection: channelAction.Target.Connection,
		User:       channelAction.Target.User,
		Session:    channelAction.Target.Session,
		Channel:    channelAction.Target.Channel,
	})

	if !ok {
		return
	}

	// Apply to all connections for target
	for _, connection := range connections {
		if channelAction.Type == protos.ChannelAction_SUBSCRIBE && !common.IncludesString(connection.Channels, channelAction.Channel) {
			connection.lock.Lock()
			connection.Channels = append(connection.Channels, channelAction.Channel)
			connection.lock.Unlock()

			channelEntry, channelExists := channels.Channels[channelAction.Channel]
			if channelExists {
				channels.Set(channelAction.Channel, append(channelEntry, connection.Id))
			} else {
				channels.Set(channelAction.Channel, []string{connection.Id})
			}

			redisClient.SAdd("channel:"+channelAction.Channel, connection.Id)
		} else if channelAction.Type == protos.ChannelAction_UNSUBSCRIBE && common.IncludesString(connection.Channels, channelAction.Channel) {
			connection.lock.Lock()
			connection.Channels = common.RemoveString(connection.Channels, channelAction.Channel)
			connection.lock.Unlock()

			channelEntry, channelExists := channels.Channels[channelAction.Channel]
			if channelExists {
				channels.Set(channelAction.Channel, common.RemoveString(channelEntry, connection.Id))
			}

			redisClient.SRem("channel:"+channelAction.Channel, connection.Id)
		} else {
			// Don't set in Redis
			return
		}

		redisClient.HSet("conn:"+connection.Id, "channels", strings.Join(connection.Channels, ","))
	}
}
