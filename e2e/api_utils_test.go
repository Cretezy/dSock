package dsock_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

var token = "abc123"

// Will eventually be separated into it's own Go dSock client

type claimOptions struct {
	Id         string
	User       string
	Session    string
	Duration   int
	Expiration int
}

func createClaim(options claimOptions) (map[string]interface{}, error) {
	params := url.Values{}

	params.Add("user", options.User)

	if options.Session != "" {
		params.Add("session", options.Session)
	}
	if options.Id != "" {
		params.Add("id", options.Id)
	}

	if options.Duration != 0 {
		params.Add("duration", strconv.Itoa(options.Duration))
	}
	if options.Expiration != 0 {
		params.Add("expiration", strconv.Itoa(options.Expiration))
	}

	req, err := http.NewRequest("POST", "http://api/claim?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type sendOptions struct {
	Id      string
	User    string
	Session string
	Type    string
	Message []byte
}

func sendMessage(options sendOptions) (map[string]interface{}, error) {
	params := url.Values{}

	params.Add("user", options.User)
	params.Add("type", options.Type)

	if options.Session != "" {
		params.Add("session", options.Session)
	}
	if options.Id != "" {
		params.Add("id", options.Id)
	}

	req, err := http.NewRequest("POST", "http://api/send?"+params.Encode(), bytes.NewReader(options.Message))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type infoOptions struct {
	Id      string
	User    string
	Session string
}

func getInfo(options infoOptions) (map[string]interface{}, error) {
	params := url.Values{}

	if options.User != "" {
		params.Add("user", options.User)
	}
	if options.Session != "" {
		params.Add("session", options.Session)
	}
	if options.Id != "" {
		params.Add("id", options.Id)
	}

	req, err := http.NewRequest("GET", "http://api/info?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

type disconnectOptions struct {
	Id         string
	User       string
	Session    string
	KeepClaims bool
}

func disconnect(options disconnectOptions) (map[string]interface{}, error) {
	params := url.Values{}

	params.Add("user", options.User)

	if options.Session != "" {
		params.Add("session", options.Session)
	}
	if options.Id != "" {
		params.Add("id", options.Id)
	}
	if options.KeepClaims {
		params.Add("keepClaims", "true")
	}

	req, err := http.NewRequest("POST", "http://api/disconnect?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
