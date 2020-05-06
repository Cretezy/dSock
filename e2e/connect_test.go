package dsock_test

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"testing"
)

func TestClaimConnect(t *testing.T) {
	claim, err := createClaim(claimOptions{
		User:    "connect",
		Session: "claim",
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
		User:    "connect",
		Session: "claim",
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

	connection := connections[0].(map[string]interface{})

	if connection["user"] != "connect" {
		t.Fatal("Connection user did not match (expected != actual): connect !=", connection["user"])
	}

	if connection["session"] != "claim" {
		t.Fatal("Connection session did not match (expected != actual): jwt !=", connection["session"])
	}
}

func TestInvalidClaim(t *testing.T) {
	_, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?claim=invalid-claim", nil)
	if err == nil {
		t.Error("Did not error when error is expected during connection")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Could not read body")
	}

	var parsedBody map[string]interface{}

	err = json.Unmarshal(body, &parsedBody)
	if err != nil {
		t.Fatal("Could not parse body")
	}

	if parsedBody["success"] == true {
		t.Fatal("Body said successful when should have failed")
	}

	if parsedBody["errorCode"] != "MISSING_CLAIM" {
		t.Fatal("Incorrect error (expected != actual): MISSING_CLAIM != ", parsedBody["errorCode"])
	}
}

func TestJwtConnect(t *testing.T) {
	// Hard coded JWT with max expiry:
	// {
	//  "sub": "connect",
	//  "sid": "jwt",
	//  "exp": 2147485546
	//}
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJjb25uZWN0Iiwic2lkIjoiand0IiwiZXhwIjoyMTQ3NDg1NTQ2fQ.oMbgPfg86I1sWs6IK25AP0H4ftzUVt9asKr9W9binW0"

	conn, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?jwt="+jwt, nil)
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
		User:    "connect",
		Session: "jwt",
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

	connection := connections[0].(map[string]interface{})

	if connection["user"] != "connect" {
		t.Fatal("Connection user did not match (expected != actual): connect !=", connection["user"])
	}

	if connection["session"] != "jwt" {
		t.Fatal("Connection session did not match (expected != actual): jwt !=", connection["session"])
	}
}

func TestInvalidJwt(t *testing.T) {
	// Hard coded JWT with invalid expiry:
	// {
	//  "sub": "connect",
	//  "sid": "invalid",
	//  "exp": "invalid"
	//}
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJjb25uZWN0Iiwic2lkIjoiaW52YWxpZCIsImV4cCI6ImludmFsaWQifQ.afZ4Mi-K0FeS35n7sivpNlq41JUi-QKVEjkH6mGWOrk"

	_, resp, err := websocket.DefaultDialer.Dial("ws://worker/connect?jwt="+jwt, nil)
	if err == nil {
		t.Error("Did not error when error is expected during connection")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("Could not read body")
	}

	var parsedBody map[string]interface{}

	err = json.Unmarshal(body, &parsedBody)
	if err != nil {
		t.Fatal("Could not parse body")
	}

	if parsedBody["success"] == true {
		t.Fatal("Body said successful when should have failed")
	}

	if parsedBody["errorCode"] != "INVALID_JWT" {
		t.Fatal("Incorrect error (expected != actual): INVALID_JWT != ", parsedBody["errorCode"])
	}
}
