package dsock_test

import (
	"bytes"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"testing"
)

func TestUserSend(t *testing.T) {
	claim, err := createClaim(claimOptions{
		User: "send_user",
	})

	if err != nil {
		t.Fatal("Error during claim creation:", err.Error())
	}

	if claim["success"].(bool) != true {
		t.Fatal("Application error during claim creation:", claim["error"])
	}

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim="+claim["claim"].(map[string]interface{})["id"].(string), nil)
	if err != nil {
		t.Error("Error during connection ("+resp.Status+"):", err.Error())
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			t.Fatal("Body:", string(body))
		} else {
			t.Fatal("Could not read body")
		}
	}

	defer conn.Close()

	message, err := sendMessage(sendOptions{
		User:    "send_user",
		Type:    "text",
		Message: []byte("Hello world!"),
	})

	if err != nil {
		t.Fatal("Error during sending:", err.Error())
	}

	if message["success"].(bool) != true {
		t.Fatal("Application error during sending:", message["error"])
	}

	messageType, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatal("Error during receiving message:", err.Error())
	}

	if messageType != websocket.TextMessage {
		t.Fatal("Message type does not match. Got:", messageType)
	}

	if string(data) != "Hello world!" {
		t.Fatal("Message does not match. Got:", string(data))
	}
}

func TestUserSessionSend(t *testing.T) {
	claim, err := createClaim(claimOptions{
		User:    "send",
		Session: "session",
	})

	if err != nil {
		t.Fatal("Error during claim creation:", err.Error())
	}

	if claim["success"].(bool) != true {
		t.Fatal("Application error during claim creation:", claim["error"])
	}

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim="+claim["claim"].(map[string]interface{})["id"].(string), nil)
	if err != nil {
		t.Error("Error during connection ("+resp.Status+"):", err.Error())
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			t.Fatal("Body:", string(body))
		} else {
			t.Fatal("Could not read body")
		}
	}

	defer conn.Close()

	message, err := sendMessage(sendOptions{
		User:    "send",
		Session: "session",
		Type:    "text",
		Message: []byte("Hello world!"),
	})

	if err != nil {
		t.Fatal("Error during sending:", err.Error())
	}

	if message["success"].(bool) != true {
		t.Fatal("Application error during sending:", message["error"])
	}

	messageType, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatal("Error during receiving message:", err.Error())
	}

	if messageType != websocket.TextMessage {
		t.Fatal("Message type does not match. Got:", messageType)
	}

	if string(data) != "Hello world!" {
		t.Fatal("Message does not match. Got:", string(data))
	}
}

func TestConnectionSend(t *testing.T) {
	claim, err := createClaim(claimOptions{
		User:    "send",
		Session: "connection",
	})

	if err != nil {
		t.Fatal("Error during claim creation:", err.Error())
	}

	if claim["success"].(bool) != true {
		t.Fatal("Application error during claim creation:", claim["error"])
	}

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim="+claim["claim"].(map[string]interface{})["id"].(string), nil)
	if err != nil {
		t.Error("Error during connection ("+resp.Status+"):", err.Error())
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			t.Fatal("Body:", string(body))
		} else {
			t.Fatal("Could not read body")
		}
	}

	defer conn.Close()

	info, err := getInfo(infoOptions{
		User:    "send",
		Session: "connection",
	})

	if err != nil {
		t.Fatal("Error during getting info:", err.Error())
	}

	if info["success"].(bool) != true {
		t.Fatal("Application error during getting info:", info["error"])
	}

	connections := info["connections"].([]interface{})

	if len(connections) != 1 {
		t.Fatal("Invalid number of connections, expected 1:", len(connections))
	}

	id := connections[0].(map[string]interface{})["id"].(string)

	message, err := sendMessage(sendOptions{
		Id:      id,
		Type:    "binary",
		Message: []byte{1, 2, 3, 4},
	})

	if err != nil {
		t.Fatal("Error during sending:", err.Error())
	}

	if message["success"].(bool) != true {
		t.Fatal("Application error during sending:", message["error"])
	}

	messageType, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatal("Error during receiving message:", err.Error())
	}

	if messageType != websocket.BinaryMessage {
		t.Fatal("Message type does not match. Got:", messageType)
	}

	if bytes.Compare(data, []byte{1, 2, 3, 4}) != 0 {
		t.Fatal("Message does not match. Got:", string(data))
	}
}

func TestSendNoTarget(t *testing.T) {
	message, err := sendMessage(sendOptions{})

	if err != nil {
		t.Fatal("Error during sending message:", err.Error())
	}

	if message["success"].(bool) != false {
		t.Fatal("Application succeeded when it should have failed during sending message:", message["error"])
	}

	if message["errorCode"] != "MISSING_CONNECTION_OR_USER" {
		t.Fatal("Error code did not match expected (MISSING_CONNECTION_OR_USER):", message["error"])
	}
}

func TestSendNoType(t *testing.T) {
	message, err := sendMessage(sendOptions{
		Id: "a",
	})

	if err != nil {
		t.Fatal("Error during sending message:", err.Error())
	}

	if message["success"].(bool) != false {
		t.Fatal("Application succeeded when it should have failed during sending message:", message["error"])
	}

	if message["errorCode"] != "INVALID_MESSAGE_TYPE" {
		t.Fatal("Error code did not match expected (INVALID_MESSAGE_TYPE):", message["error"])
	}
}
