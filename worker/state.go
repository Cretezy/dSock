package main

import "sync"

type connectionsState struct {
	Connections map[string]*SockConnection
	Mutex       sync.Mutex
}

var connections = connectionsState{
	Connections: make(map[string]*SockConnection),
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

var users = usersState{
	Users: make(map[string][]string),
}

func (users *usersState) Set(user string, connections []string) {
	users.Mutex.Lock()
	defer users.Mutex.Unlock()

	users.Users[user] = connections
}

type channelsState struct {
	Channels map[string][]string
	Mutex    sync.Mutex
}

var channels = channelsState{
	Channels: make(map[string][]string),
}

func (channels *channelsState) Set(channel string, connections []string) {
	users.Mutex.Lock()
	defer users.Mutex.Unlock()

	channels.Channels[channel] = connections
}
