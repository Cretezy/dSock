package dsock_test

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var token = "abc123"

// Will eventually be separated into it's own Go dSock client

type target struct {
	Id      string
	User    string
	Session string
	Channel string
}

func addTargetToParams(target target, params url.Values) {
	if target.User != "" {
		params.Add("user", target.User)
	}
	if target.Session != "" {
		params.Add("session", target.Session)
	}
	if target.Id != "" {
		params.Add("id", target.Id)
	}
	if target.Channel != "" {
		params.Add("channel", target.Channel)
	}
}

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

	result["_raw"] = string(responseBody)

	return result, nil
}

type claimOptions struct {
	Id         string
	User       string
	Session    string
	Duration   int
	Expiration int
	Channels   []string
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
	if len(options.Channels) != 0 {
		params.Add("channels", strings.Join(options.Channels, ","))
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
	target
	Type    string
	Message []byte
}

func sendMessage(options sendOptions) (map[string]interface{}, error) {
	params := url.Values{}

	addTargetToParams(options.target, params)

	params.Add("type", options.Type)

	return doApiRequest("POST", "/send", params, bytes.NewReader(options.Message))
}

type infoOptions struct {
	target
}

func getInfo(options infoOptions) (map[string]interface{}, error) {
	params := url.Values{}

	addTargetToParams(options.target, params)

	return doApiRequest("GET", "/info", params, nil)
}

type disconnectOptions struct {
	target

	KeepClaims bool
}

func disconnect(options disconnectOptions) (map[string]interface{}, error) {
	params := url.Values{}

	addTargetToParams(options.target, params)

	if options.KeepClaims {
		params.Add("keepClaims", "true")
	}

	return doApiRequest("POST", "/disconnect", params, nil)

}

type channelOptions struct {
	target
	Channel      string
	IgnoreClaims bool
}

func subscribeChannel(options channelOptions) (map[string]interface{}, error) {
	params := url.Values{}

	addTargetToParams(options.target, params)

	if options.IgnoreClaims {
		params.Set("ignoreClaims", "true")
	}

	return doApiRequest("POST", "/channel/subscribe/"+options.Channel, params, nil)
}

func unsubscribeChannel(options channelOptions) (map[string]interface{}, error) {
	params := url.Values{}

	addTargetToParams(options.target, params)

	if options.IgnoreClaims {
		params.Set("ignoreClaims", "true")
	}

	return doApiRequest("POST", "/channel/unsubscribe/"+options.Channel, params, nil)
}
