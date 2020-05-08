package dsock_test

import (
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"
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

	// Connection was successful, wait a tiny bit to make sure connection is set in Redis
	time.Sleep(time.Millisecond)

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

func interfaceToStringSlice(slice interface{}) []string {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("interfaceSlice() given a non-slice type")
	}

	ret := make([]string, s.Len())

	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface().(string)
	}

	return ret
}
