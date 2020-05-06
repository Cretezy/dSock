package common

import "github.com/gin-gonic/gin"

var (
	ErrorUserIdRequired          = "USER_ID_REQUIRED"
	ErrorInvalidExpiration       = "INVALID_EXPIRATION"
	ErrorNegativeExpiration      = "NEGATIVE_EXPIRATION"
	ErrorInvalidDuration         = "INVALID_DURATION"
	ErrorNegativeDuration        = "NEGATIVE_DURATION"
	ErrorGettingConnection       = "ERROR_GETTING_CONNECTION"
	ErrorGettingUser             = "ERROR_GETTING_USER"
	ErrorMissingConnectionOrUser = "MISSING_CONNECTION_OR_USER"
	ErrorInvalidAuthorization    = "INVALID_AUTHORIZATION"
	ErrorMissingAuthentication   = "MISSING_AUTHENTICATION"
	ErrorInvalidJwt              = "INVALID_JWT"
	ErrorClaimIdAlreadyUsed      = "CLAIM_ID_ALREADY_USED"
	ErrorCheckingClaim           = "ERROR_CHECKING_CLAIM"
	ErrorGettingClaim            = "ERROR_GETTING_CLAIM"
	ErrorMissingClaim            = "MISSING_CLAIM"
	ErrorExpiredClaim            = "EXPIRED_CLAIM"
	ErrorReadingMessage          = "ERROR_READING_MESSAGE"
	ErrorMarshallingMessage      = "ERROR_MARSHALLING_MESSAGE"
	ErrorInvalidMessageType      = "INVALID_MESSAGE_TYPE"
)

var ErrorMessages = map[string]string{
	ErrorUserIdRequired:          "User ID is required",
	ErrorInvalidExpiration:       "Error parsing expiration (must be a integer)",
	ErrorNegativeExpiration:      "Can not use 0 or negative expiration",
	ErrorInvalidDuration:         "Could not parse duration (must be a integer)",
	ErrorNegativeDuration:        "Can not use 0 or negative duration",
	ErrorGettingConnection:       "Error getting connection",
	ErrorGettingUser:             "Error getting user",
	ErrorMissingConnectionOrUser: "Connection ID or user ID is missing",
	ErrorInvalidAuthorization:    "Invalid authorization",
	ErrorMissingAuthentication:   "Did not provide an authentication method",
	ErrorInvalidJwt:              "Could not validate JWT",
	ErrorClaimIdAlreadyUsed:      "Claim ID is already used",
	ErrorCheckingClaim:           "Error checking if claim already exists",
	ErrorGettingClaim:            "Error getting claim",
	ErrorMissingClaim:            "Could not find claim",
	ErrorExpiredClaim:            "Claim has expired",
	ErrorReadingMessage:          "Error reading message",
	ErrorMarshallingMessage:      "Error marshalling message",
	ErrorInvalidMessageType:      "Invalid message type, must be text or binary",
}

type ApiError struct {
	/// Wrapped error. Can be nil if application/validation error
	InternalError error
	/// Error code. Used to take error message from ErrorMessages
	ErrorCode string
	/// If set, overrides error message from ErrorCode
	CustomErrorMessage string
	/// HTTP status code
	StatusCode int
}

func (apiError *ApiError) Error() string {
	if apiError.InternalError != nil {
		return apiError.ErrorCode
	}

	return apiError.ErrorCode + ": " + apiError.InternalError.Error()
}

func (apiError *ApiError) Format() (int, gin.H) {
	statusCode := apiError.StatusCode
	if statusCode == 0 {
		statusCode = 500
	}

	errorMessage := ErrorMessages[apiError.ErrorCode]

	if apiError.CustomErrorMessage != "" {
		errorMessage = apiError.CustomErrorMessage
	}

	if errorMessage == "" {
		// Default message if unknown error code
		errorMessage = apiError.ErrorCode
	}

	return statusCode, gin.H{
		"success":   false,
		"errorCode": apiError.ErrorCode,
		"error":     errorMessage,
	}
}

func (apiError *ApiError) Send(c *gin.Context) {
	c.AbortWithStatusJSON(apiError.Format())
}
