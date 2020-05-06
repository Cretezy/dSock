package common_test

import (
	"github.com/Cretezy/dSock/common"
	"testing"
)

func TestApiError_Format(t *testing.T) {
	apiError := common.ApiError{
		StatusCode: 400,
		ErrorCode:  common.ErrorInvalidAuthorization,
	}

	statusCode, body := apiError.Format()

	if statusCode != 400 {
		t.Fatalf("Status code did not match (expected != action): 400 != %v", statusCode)
	}

	if body["errorCode"] != common.ErrorInvalidAuthorization {
		t.Fatalf("Error code did not match (expected != action): %s != %s", common.ErrorInvalidAuthorization, body["errorCode"])
	}

	if body["error"] != "Invalid authorization" {
		t.Fatalf("Error message did not match (expected != action): Invalid authorization != %s", body["error"])
	}
}

func TestApiError_Format_NoStatusCode(t *testing.T) {
	apiError := common.ApiError{
		ErrorCode: common.ErrorInvalidAuthorization,
	}

	statusCode, _ := apiError.Format()

	if statusCode != 500 {
		t.Fatalf("Status code did not match (expected != action): 500 != %v", statusCode)
	}
}

func TestApiError_Format_CustomErrorMessage(t *testing.T) {
	apiError := common.ApiError{
		ErrorCode:          common.ErrorInvalidAuthorization,
		CustomErrorMessage: "Custom Error",
	}

	_, body := apiError.Format()

	if body["errorCode"] != common.ErrorInvalidAuthorization {
		t.Fatalf("Error code did not match (expected != action): %s != %s", common.ErrorInvalidAuthorization, body["errorCode"])
	}

	if body["error"] != "Custom Error" {
		t.Fatalf("Error message did not match (expected != action): Custom Error != %s", body["error"])
	}
}

func TestApiError_Format_NoErrorMessage(t *testing.T) {
	apiError := common.ApiError{
		ErrorCode: "_OTHER_ERROR",
	}

	_, body := apiError.Format()

	if body["errorCode"] != "_OTHER_ERROR" {
		t.Fatalf("Error code did not match (expected != action): _OTHER_ERROR != %s", body["errorCode"])
	}

	if body["error"] != "_OTHER_ERROR" {
		t.Fatalf("Error message did not match (expected != action): _OTHER_ERROR != %s", body["error"])
	}
}
