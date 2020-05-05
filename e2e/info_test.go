package dsock_test

import (
	"github.com/gorilla/websocket"
	"io/ioutil"
	"testing"
)

func TestInfoClaim(t *testing.T) {
	claim, err := createClaim(claimOptions{
		User:    "info",
		Session: "claim",
	})

	if err != nil {
		t.Fatal("Error during claim creation:", err.Error())
	}

	if claim["success"].(bool) != true {
		t.Fatal("Application error during claim creation:", claim["error"])
	}

	info, err := getInfo(infoOptions{
		User:    "info",
		Session: "claim",
	})

	if err != nil {
		t.Fatal("Error during getting info:", err.Error())
	}

	if info["success"].(bool) != true {
		t.Fatal("Application error during getting info:", info["error"])
	}

	infoClaims := info["claims"].([]interface{})

	if len(infoClaims) != 1 {
		t.Fatal("Invalid number of claims, expected 1:", len(infoClaims))
	}

	claimData := claim["claim"].(map[string]interface{})

	infoClaimData := infoClaims[0].(map[string]interface{})

	if claimData["id"] != infoClaimData["id"] {
		t.Fatal("Info claim ID doesn't match (expected != actual):", claimData["id"], "!=", infoClaimData["id"])
	}

	if claimData["user"] != infoClaimData["user"] {
		t.Fatal("Info claim user doesn't match (expected != actual):", claimData["user"], "!=", infoClaimData["user"])
	}

	if claimData["session"] != infoClaimData["session"] {
		t.Fatal("Info claim session doesn't match (expected != actual):", claimData["session"], "!=", infoClaimData["session"])
	}
}

func TestInfoConnection(t *testing.T) {
	claim, err := createClaim(claimOptions{
		User:    "info",
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
		User:    "info",
		Session: "connection",
	})

	if err != nil {
		t.Fatal("Error during getting info:", err.Error())
	}

	if info["success"].(bool) != true {
		t.Fatal("Application error during getting info:", info["error"])
	}

	infoConnections := info["connections"].([]interface{})

	if len(infoConnections) != 1 {
		t.Fatal("Invalid number of connections, expected 1:", len(infoConnections))
	}

	claimData := claim["claim"].(map[string]interface{})

	infoConnectionData := infoConnections[0].(map[string]interface{})

	if claimData["user"] != infoConnectionData["user"] {
		t.Fatal("Info claim user doesn't match (expected != actual):", claimData["user"], "!=", infoConnectionData["user"])
	}

	if claimData["session"] != infoConnectionData["session"] {
		t.Fatal("Info claim session doesn't match (expected != actual):", claimData["session"], "!=", infoConnectionData["session"])
	}
}
