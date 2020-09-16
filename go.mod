module github.com/Cretezy/dSock

go 1.13

require (
	github.com/Cretezy/dSock-go v1.0.2
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/gin-contrib/requestid v0.0.0-20200512155051-855d6508f0f0
	github.com/gin-contrib/zap v0.0.1
	github.com/gin-gonic/gin v1.6.3
	github.com/go-redis/redis/v7 v7.2.0
	github.com/golang/protobuf v1.4.0-rc.4.0.20200313231945-b860323f09d0
	github.com/google/uuid v1.1.1
	github.com/gorilla/websocket v1.4.2
	github.com/spf13/viper v1.6.3
	github.com/stretchr/testify v1.4.0
	go.uber.org/zap v1.10.0
	google.golang.org/protobuf v1.21.0
)

replace github.com/Cretezy/common/protos => ./common/build/gen/protos/github.com/Cretezy/dSock/common/protos
