package main

import (
	"github.com/Cretezy/dSock/common"
	"github.com/Cretezy/dSock/common/protos"
)

func send(message *protos.Message) {
	// Resolve all local connections for message target
	connections, ok := resolveConnections(common.ResolveOptions{
		Connection: message.Connection,
		User:       message.User,
		Session:    message.Session,
	})

	if !ok {
		return
	}

	// Send to all connections for target
	for _, connection := range connections {
		if connection.Sender == nil || connection.CloseChannel == nil {
			continue
		}

		connection := connection

		go func() {
			if message.Type == protos.Message_DISCONNECT {
				connection.CloseChannel <- struct{}{}
			} else {
				connection.Sender <- message
			}
		}()
	}
}
