package dsock_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

var token = "abc123"

// Will eventually be separated into it's own Go dSock client

func doApiRequest(method, path string, params url.Values, body io.Reader) (map[string]interface{}, error) {
	request, err := http.NewRequest(method, "http://api"+path+"?"+params.Encode(), body)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}

	err = json.Unmarshal(responseBody, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

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

	return doApiRequest("POST", "/claim", params, nil)
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

	return doApiRequest("POST", "/send", params, bytes.NewReader(options.Message))
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

	return doApiRequest("GET", "/info", params, nil)
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

	return doApiRequest("POST", "/disconnect", params, nil)

}
