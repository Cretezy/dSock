package common

import "github.com/gin-gonic/gin"

/// Validates that token matches from query parameter or from Authorization header
func TokenMiddleware(token string) gin.HandlerFunc {
	isAuthorized := func(c *gin.Context) bool {
		if token == c.Query("token") {
			return true
		}

		tokenHeader := c.GetHeader("Authorization")
		if len(tokenHeader) > 7 {
			// Removes "Bearer "
			if token == tokenHeader[7:] {
				return true
			}
		}

		return false
	}

	return func(c *gin.Context) {
		if !isAuthorized(c) {
			apiError := ApiError{
				StatusCode: 400,
				ErrorCode:  ErrorInvalidAuthorization,
			}

			apiError.Send(c)
		}
	}
}
