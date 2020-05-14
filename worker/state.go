package main

import (
	"github.com/Cretezy/dSock/common"
	"sync"
)

type connectionsState struct {
	state map[string]*SockConnection
	mutex sync.RWMutex
}

func (connections *connectionsState) Add(connection *SockConnection) {
	connections.mutex.Lock()
	defer connections.mutex.Unlock()

	connections.state[connection.Id] = connection
}

func (connections *connectionsState) Remove(id string) {
	connections.mutex.Lock()
	defer connections.mutex.Unlock()

	delete(connections.state, id)
}

func (connections *connectionsState) Get(id string) (*SockConnection, bool) {
	connections.mutex.RLock()
	defer connections.mutex.RUnlock()

	connectionEntry, connectionExists := connections.state[id]
	return connectionEntry, connectionExists
}

type usersState struct {
	state map[string][]string
	mutex sync.RWMutex
}

func (users *usersState) Add(user string, connection string) {
	users.mutex.Lock()
	defer users.mutex.Unlock()

	usersEntry, usersExists := users.state[user]
	if usersExists {
		users.state[user] = append(usersEntry, connection)
	} else {
		users.state[user] = []string{connection}
	}
}

func (users *usersState) Remove(user string, connection string) {
	users.mutex.Lock()
	defer users.mutex.Unlock()

	usersEntry, usersExists := users.state[user]
	if usersExists {
		users.state[user] = common.RemoveString(usersEntry, connection)
	}
}

func (users *usersState) Get(user string) ([]string, bool) {
	users.mutex.RLock()
	defer users.mutex.RUnlock()

	userEntry, userExists := users.state[user]
	return userEntry, userExists
}

type channelsState struct {
	state map[string][]string
	mutex sync.RWMutex
}

func (channels *channelsState) Add(channel string, connection string) {
	channels.mutex.Lock()
	defer channels.mutex.Unlock()

	channelEntry, channelExists := channels.state[channel]
	if channelExists {
		channels.state[channel] = append(channelEntry, connection)
	} else {
		channels.state[channel] = []string{connection}
	}
}

func (channels *channelsState) Remove(channel string, connection string) {
	channels.mutex.Lock()
	defer channels.mutex.Unlock()

	channelEntry, channelExists := channels.state[channel]
	if channelExists {
		channels.state[channel] = common.RemoveString(channelEntry, connection)
	}
}

func (channels *channelsState) Get(channel string) ([]string, bool) {
	channels.mutex.RLock()
	defer channels.mutex.RUnlock()

	channelEntry, channelExists := channels.state[channel]
	return channelEntry, channelExists
}
