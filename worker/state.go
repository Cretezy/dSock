package main

import (
	"github.com/Cretezy/dSock/common"
	"sync"
)

type connectionsState struct {
	Connections map[string]*SockConnection
	Mutex       sync.Mutex
}

func (connections *connectionsState) Add(connection *SockConnection) {
	connections.Mutex.Lock()
	defer connections.Mutex.Unlock()

	connections.Connections[connection.Id] = connection
}

func (connections *connectionsState) Remove(id string) {
	connections.Mutex.Lock()
	defer connections.Mutex.Unlock()

	delete(connections.Connections, id)
}

type usersState struct {
	Users map[string][]string
	Mutex sync.Mutex
}

func (users *usersState) Add(user string, connection string) {
	users.Mutex.Lock()
	defer users.Mutex.Unlock()

	usersEntry, usersExists := users.Users[user]
	if usersExists {
		users.Users[user] = append(usersEntry, connection)
	} else {
		users.Users[user] = []string{connection}
	}
}

func (users *usersState) Remove(user string, connection string) {
	users.Mutex.Lock()
	defer users.Mutex.Unlock()

	usersEntry, usersExists := users.Users[user]
	if usersExists {
		users.Users[user] = common.RemoveString(usersEntry, connection)
	}
}

type channelsState struct {
	Channels map[string][]string
	Mutex    sync.Mutex
}

func (channels *channelsState) Add(channel string, connection string) {
	channels.Mutex.Lock()
	defer channels.Mutex.Unlock()

	channelEntry, channelExists := channels.Channels[channel]
	if channelExists {
		channels.Channels[channel] = append(channelEntry, connection)
	} else {
		channels.Channels[channel] = []string{connection}
	}
}

func (channels *channelsState) Remove(channel string, connection string) {
	channels.Mutex.Lock()
	defer channels.Mutex.Unlock()

	channelEntry, channelExists := channels.Channels[channel]
	if channelExists {
		channels.Channels[channel] = common.RemoveString(channelEntry, connection)
	}
}