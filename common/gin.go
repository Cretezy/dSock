package common

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)
import "github.com/gin-contrib/zap"

func NewGinEngine(logger *zap.Logger, options *DSockOptions) *gin.Engine {
	engine := gin.New()
	if options.LogRequests {
		engine.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	}
	engine.Use(ginzap.RecoveryWithZap(logger, options.Debug))
	return engine
}
