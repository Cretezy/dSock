package common

import "github.com/gin-gonic/gin"

func TokenMiddleware(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		check := func() bool {
			if token == c.Query("token") {
				return true
			}

			tokenHeader := c.GetHeader("Authorization")
			if len(tokenHeader) > 7 {
				if token == tokenHeader[7:] {
					return true
				}
			}

			return false
		}

		if !check() {
			c.AbortWithStatusJSON(400, map[string]interface{}{
				"success":   false,
				"error":     "Invalid authorization",
				"errorCode": ErrorInvalidAuthorization,
			})

			return
		}
	}
}
