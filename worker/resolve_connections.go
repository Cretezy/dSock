package main

type ResolveOptions struct {
	Connection string
	User       string
	Session    string
}

func resolveConnections(options ResolveOptions) ([]SockConnection, bool) {
	if options.Connection != "" {
		sockConnection, exists := connections[options.Connection]

		if !exists {
			// Connection doesnt' exist
			return []SockConnection{}, true
		}

		return []SockConnection{sockConnection}, true
	} else if options.User != "" {
		usersEntry, exists := users[options.User]

		if !exists {
			// User doesn't exist
			return []SockConnection{}, true
		}

		senders := make([]SockConnection, 0)

		for _, connectionId := range usersEntry {
			connection, connectionExists := connections[connectionId]
			// Target a specific session for a user if set
			if connectionExists && (options.Session == "" || connection.Session == options.Session) {
				senders = append(senders, connection)
			}
		}

		return senders, true
	} else {
		return []SockConnection{}, false
	}
}
