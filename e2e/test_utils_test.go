package dsock_test

import (
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
)

func checkConnectionError(suite suite.Suite, err error, resp *http.Response) bool {
	if !suite.NoErrorf(err, "Error during connection (%s)", resp.Status) {
		body, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			suite.T().Logf("Body: %s", string(body))
		} else {
			suite.T().Log("Could not read body")
		}

		return false
	}

	return true
}

func checkRequestError(suite suite.Suite, err error, body map[string]interface{}, during string) bool {
	if !suite.NoErrorf(err, "Error during %s", during) {
		return false
	}

	if !suite.Equalf(true, body["success"], "Application error during %s (%s)", during, body["errorCode"]) {
		return false
	}

	return true
}

func checkRequestNoError(suite suite.Suite, err error, body map[string]interface{}, errorCode, during string) bool {
	if !suite.NoErrorf(err, "Error during %s", during) {
		return false
	}

	if !suite.Equalf(false, body["success"], "Application succeeded during %s when expecting %s", during, errorCode) {
		return false
	}

	if !suite.Equalf(errorCode, body["errorCode"], "Incorrect error code") {
		return false
	}

	return true
}
