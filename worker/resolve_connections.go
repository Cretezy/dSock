package main

import "github.com/Cretezy/dSock/common"

func resolveConnections(options common.ResolveOptions) ([]*SockConnection, bool) {
	if options.Connection != "" {
		connectionEntry, connectionExists := connections.Get(options.Connection)

		if !connectionExists {
			// Connection doesnt' exist
			return []*SockConnection{}, true
		}

		return []*SockConnection{connectionEntry}, true
	} else if options.Channel != "" {
		channelEntry, exists := channels.Get(options.Channel)

		if !exists {
			// User doesn't exist
			return []*SockConnection{}, true
		}

		senders := make([]*SockConnection, 0)

		for _, connectionId := range channelEntry {
			connection, connectionExists := connections.Get(connectionId)
			// Target a specific session for a user if set
			if connectionExists && (options.Session == "" || connection.Session == options.Session) {
				senders = append(senders, connection)
			}
		}

		return senders, true
	} else if options.User != "" {
		usersEntry, exists := users.Get(options.User)

		if !exists {
			// User doesn't exist
			return []*SockConnection{}, true
		}

		senders := make([]*SockConnection, 0)

		for _, connectionId := range usersEntry {
			connection, connectionExists := connections.Get(connectionId)
			// Target a specific session for a user if set
			if connectionExists && (options.Session == "" || connection.Session == options.Session) {
				senders = append(senders, connection)
			}
		}

		return senders, true
	} else {
		// No target
		return []*SockConnection{}, false
	}
}
