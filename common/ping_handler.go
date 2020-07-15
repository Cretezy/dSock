package common

import (
	"github.com/gin-gonic/gin"
)

func PingHandler(c *gin.Context) {
	c.String(200, "pong")
}
