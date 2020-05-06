package dsock_test

import (
	"github.com/gorilla/websocket"
	"io/ioutil"
	"testing"
	"time"
)

func TestUserDisconnect(t *testing.T) {
	claim, err := createClaim(claimOptions{
		User: "disconnect_user",
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

	disconnectResponse, err := disconnect(disconnectOptions{
		User: "disconnect_user",
	})

	if err != nil {
		t.Fatal("Error during disconnection:", err.Error())
	}

	if disconnectResponse["success"].(bool) != true {
		t.Fatal("Application error during disconnection:", disconnectResponse["error"])
	}

	_, _, err = conn.ReadMessage()
	if err == nil {
		t.Fatal("Did not receive error when was expecting")
	}

	if closeErr, ok := err.(*websocket.CloseError); ok {
		if closeErr.Code != websocket.CloseNormalClosure {
			t.Fatalf("Close type did not match (expected != action): %v != %v", websocket.CloseNormalClosure, closeErr.Code)
		}
	} else {
		t.Fatal("Error type does not match. Got:", err)
	}
}

func TestSessionDisconnect(t *testing.T) {
	claim, err := createClaim(claimOptions{
		User:    "disconnect",
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

	disconnectResponse, err := disconnect(disconnectOptions{
		User:    "disconnect",
		Session: "session",
	})

	if err != nil {
		t.Fatal("Error during disconnection:", err.Error())
	}

	if disconnectResponse["success"].(bool) != true {
		t.Fatal("Application error during disconnection:", disconnectResponse["error"])
	}

	_, _, err = conn.ReadMessage()
	if err == nil {
		t.Fatal("Did not receive error when was expecting")
	}

	if closeErr, ok := err.(*websocket.CloseError); ok {
		if closeErr.Code != websocket.CloseNormalClosure {
			t.Fatalf("Close type did not match (expected != action): %v != %v", websocket.CloseNormalClosure, closeErr.Code)
		}
	} else {
		t.Fatal("Error type does not match. Got:", err)
	}
}

func TestConnectionDisconnect(t *testing.T) {
	claim, err := createClaim(claimOptions{
		User:    "disconnect",
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
		User:    "disconnect",
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

	disconnectResponse, err := disconnect(disconnectOptions{
		Id: id,
	})

	if err != nil {
		t.Fatal("Error during disconnection:", err.Error())
	}

	if disconnectResponse["success"].(bool) != true {
		t.Fatal("Application error during disconnection:", disconnectResponse["error"])
	}

	_, _, err = conn.ReadMessage()
	if err == nil {
		t.Fatal("Did not receive error when was expecting")
	}

	if closeErr, ok := err.(*websocket.CloseError); ok {
		if closeErr.Code != websocket.CloseNormalClosure {
			t.Fatalf("Close type did not match (expected != action): %v != %v", websocket.CloseNormalClosure, closeErr.Code)
		}
	} else {
		t.Fatal("Error type does not match. Got:", err)
	}
}

func TestNoSessionDisconnect(t *testing.T) {
	claim, err := createClaim(claimOptions{
		User:    "disconnect",
		Session: "no_session",
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

	var websocketErr error

	go func() {
		_, _, websocketErr = conn.ReadMessage()
	}()

	disconnectResponse, err := disconnect(disconnectOptions{
		User:    "disconnect",
		Session: "no_session_bad",
	})

	if err != nil {
		t.Fatal("Error during disconnection:", err.Error())
	}

	if disconnectResponse["success"].(bool) != true {
		t.Fatal("Application error during disconnection:", disconnectResponse["error"])
	}

	// Wait a bit to see if connection was closed
	time.Sleep(time.Millisecond * 100)

	if websocketErr != nil {
		t.Fatalf("Got websocket error when wasn't expected: %s", websocketErr)
	}
}

func TestDisconnectExpireClaim(t *testing.T) {
	claim, err := createClaim(claimOptions{
		User:    "disconnect",
		Session: "expire_claim",
	})

	if err != nil {
		t.Fatal("Error during claim creation:", err.Error())
	}

	if claim["success"].(bool) != true {
		t.Fatal("Application error during claim creation:", claim["error"])
	}

	infoBefore, err := getInfo(infoOptions{
		User:    "disconnect",
		Session: "expire_claim",
	})

	if err != nil {
		t.Fatal("Error during getting info:", err.Error())
	}

	if infoBefore["success"].(bool) != true {
		t.Fatal("Application error during getting info:", infoBefore["error"])
	}

	claimsBefore := infoBefore["claims"].([]interface{})

	if len(claimsBefore) != 1 {
		t.Fatalf("Invalid number of claims (expected != actual): 1 != %v", len(claimsBefore))
	}

	disconnectResponse, err := disconnect(disconnectOptions{
		User:    "disconnect",
		Session: "expire_claim",
	})

	if err != nil {
		t.Fatal("Error during disconnection:", err.Error())
	}

	if disconnectResponse["success"].(bool) != true {
		t.Fatal("Application error during disconnection:", disconnectResponse["error"])
	}

	infoAfter, err := getInfo(infoOptions{
		User:    "disconnect",
		Session: "expire_claim",
	})

	if infoAfter["success"].(bool) != true {
		t.Fatal("Application error during getting info:", infoAfter["error"])
	}

	claimsAfter := infoAfter["claims"].([]interface{})

	if len(claimsAfter) != 0 {
		t.Fatalf("Invalid number of claims (expected != actual): 0 != %v", len(claimsAfter))
	}
}

func TestDisconnectKeepClaim(t *testing.T) {
	claim, err := createClaim(claimOptions{
		User:    "disconnect",
		Session: "keep_claim",
	})

	if err != nil {
		t.Fatal("Error during claim creation:", err.Error())
	}

	if claim["success"].(bool) != true {
		t.Fatal("Application error during claim creation:", claim["error"])
	}

	infoBefore, err := getInfo(infoOptions{
		User:    "disconnect",
		Session: "keep_claim",
	})

	if err != nil {
		t.Fatal("Error during getting info:", err.Error())
	}

	if infoBefore["success"].(bool) != true {
		t.Fatal("Application error during getting info:", infoBefore["error"])
	}

	claimsBefore := infoBefore["claims"].([]interface{})

	if len(claimsBefore) != 1 {
		t.Fatalf("Invalid number of claims (expected != actual): 1 != %v", len(claimsBefore))
	}

	disconnectResponse, err := disconnect(disconnectOptions{
		User:       "disconnect",
		Session:    "keep_claim",
		KeepClaims: true,
	})

	if err != nil {
		t.Fatal("Error during disconnection:", err.Error())
	}

	if disconnectResponse["success"].(bool) != true {
		t.Fatal("Application error during disconnection:", disconnectResponse["error"])
	}

	infoAfter, err := getInfo(infoOptions{
		User:    "disconnect",
		Session: "keep_claim",
	})

	if infoAfter["success"].(bool) != true {
		t.Fatal("Application error during getting info:", infoAfter["error"])
	}

	claimsAfter := infoAfter["claims"].([]interface{})

	if len(claimsAfter) != 1 {
		t.Fatalf("Invalid number of claims (expected != actual): 1 != %v", len(claimsAfter))
	}
}
