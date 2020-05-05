package common

var (
	ErrorUserIdRequired = "USER_ID_REQUIRED"

	ErrorInvalidExpiration  = "INVALID_EXPIRATION"
	ErrorNegativeExpiration = "NEGATIVE_EXPIRATION"

	ErrorInvalidDuration  = "INVALID_DURATION"
	ErrorNegativeDuration = "NEGATIVE_DURATION"

	ErrorGettingConnection = "ERROR_GETTING_CONNECTION"
	ErrorMissingConnection = "MISSING_CONNECTION"

	ErrorGettingUser = "ERROR_GETTING_USER"
	ErrorMissingUser = "MISSING_USER"

	ErrorMissingConnectionOrUser = "MISSING_CONNECTION_OR_USER"

	ErrorGettingWorker = "ERROR_GETTING_WORKER"

	ErrorInvalidAuthorization  = "INVALID_AUTHORIZATION"
	ErrorMissingAuthentication = "MISSING_AUTHENTICATION"
	ErrorInvalidJwt            = "INVALID_JWT"

	ErrorClaimIdAlreadyUsed = "CLAIM_ID_ALREADY_USED"
	ErrorGettingClaim       = "ERROR_GETTING_CLAIM"
	ErrorMissingClaim       = "MISSING_CLAIM"
	ErrorExpiredClaim       = "EXPIRED_CLAIM"

	ErrorReadingMessage     = "ERROR_READING_MESSAGE"
	ErrorMarshallingMessage = "ERROR_MARSHALLING_MESSAGE"

	ErrorInvalidMessageType = "INVALID_MESSAGE_TYPE"
)
