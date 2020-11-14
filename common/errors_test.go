package common_test

import (
	"github.com/Cretezy/dSock/common"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ApiErrorSuite struct {
	suite.Suite
}

func TestApiErrorSuite(t *testing.T) {
	suite.Run(t, new(ApiErrorSuite))
}

func (suite *ApiErrorSuite) TestFormat() {
	apiError := common.ApiError{
		StatusCode: 400,
		ErrorCode:  common.ErrorInvalidAuthorization,
		RequestId:  "some-request-id",
	}

	statusCode, body := apiError.Format()

	if !suite.Equal(400, statusCode, "Incorrect status code") {
		return
	}

	if !suite.Equal(common.ErrorInvalidAuthorization, body["errorCode"], "Incorrect error code") {
		return
	}

	if !suite.Equal("Invalid authorization", body["error"], "Incorrect error message") {
		return
	}

	if !suite.Equal("some-request-id", body["requestId"], "Incorrect request ID") {
		return
	}
}

func (suite *ApiErrorSuite) TestFormatNoStatusCode() {
	apiError := common.ApiError{
		ErrorCode: common.ErrorInvalidAuthorization,
	}

	statusCode, _ := apiError.Format()

	if !suite.Equal(500, statusCode, "Incorrect status code") {
		return
	}
}

func (suite *ApiErrorSuite) TestFormatCustomErrorMessage() {
	apiError := common.ApiError{
		ErrorCode:          common.ErrorInvalidAuthorization,
		CustomErrorMessage: "Custom Error",
	}

	_, body := apiError.Format()

	if !suite.Equal(common.ErrorInvalidAuthorization, body["errorCode"], "Incorrect error code") {
		return
	}

	if !suite.Equal("Custom Error", body["error"], "Incorrect error message") {
		return
	}
}

func (suite *ApiErrorSuite) TestFormatNoErrorMessage() {
	apiError := common.ApiError{
		ErrorCode: "_OTHER_ERROR",
	}

	_, body := apiError.Format()

	if !suite.Equal("_OTHER_ERROR", body["errorCode"], "Incorrect error code") {
		return
	}

	if !suite.Equal("_OTHER_ERROR", body["error"], "Incorrect error message") {
		return
	}
}
