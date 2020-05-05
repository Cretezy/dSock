package main

import (
	"github.com/Cretezy/dSock/common/protos"
)

func send(message *protos.Message) {
	connections, ok := resolveConnections(ResolveOptions{
		Connection: message.Connection,
		User:       message.User,
		Session:    message.Session,
	})

	if !ok {
		return
	}

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
