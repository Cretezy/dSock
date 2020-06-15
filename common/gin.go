package common

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)
import "github.com/gin-contrib/zap"

func NewGinEngine(logger *zap.Logger, stack bool) *gin.Engine {
	engine := gin.New()
	engine.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	engine.Use(ginzap.RecoveryWithZap(logger, stack))
	return engine
}
