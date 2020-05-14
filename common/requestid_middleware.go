package common

import (
	"github.com/gin-contrib/requestid"
	"github.com/google/uuid"
)

var RequestIdMiddleware = requestid.New(requestid.Config{
	Generator: func() string {
		return uuid.New().String()
	},
})
