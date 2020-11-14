package common

import (
	"crypto/tls"
	"errors"
	"github.com/go-redis/redis/v7"
	"github.com/spf13/viper"
	"net"
	"os"
	"strconv"
	"strings"
)

const MessageMethodRedis = "redis"
const MessageMethodDirect = "direct"

type JwtOptions struct {
	JwtSecret string
}

type DSockOptions struct {
	RedisOptions *redis.Options
	Address      string
	Port         int
	QuitChannel  chan struct{}
	Debug        bool
	LogRequests  bool
	/// Token for your API -> dSock and between dSock services
	Token string
	/// JWT parsing/verifying options
	Jwt JwtOptions
	/// Default channels to subscribe on join
	DefaultChannels []string
	/// The message method between the API to the worker
	MessagingMethod string
	/// The worker hostname
	DirectHostname string
	/// The worker port
	DirectPort int
}

func SetupConfig() error {
	viper.SetConfigName("config")
	viper.SetEnvPrefix("DSOCK")
	viper.AutomaticEnv()

	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config/dsock")
	viper.AddConfigPath("/etc/dsock")

	viper.SetDefault("redis_host", "localhost:6379")
	viper.SetDefault("redis_password", "")
	viper.SetDefault("redis_db", 0)
	viper.SetDefault("redis_max_retries", 10)
	viper.SetDefault("redis_tls", false)
	viper.SetDefault("port", 6241)
	viper.SetDefault("default_channels", "")
	viper.SetDefault("token", "")
	viper.SetDefault("jwt_secret", "")
	viper.SetDefault("debug", false)
	viper.SetDefault("log_requests", false)
	viper.SetDefault("messaging_method", "redis")
	viper.SetDefault("direct_message_hostname", "")
	viper.SetDefault("direct_message_port", "")

	err := viper.ReadInConfig()

	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			err = viper.SafeWriteConfigAs("config.toml")

			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func GetOptions(worker bool) (*DSockOptions, error) {
	err := SetupConfig()
	if err != nil {
		return nil, err
	}

	port := viper.GetInt("port")

	if os.Getenv("PORT") != "" {
		println("Port env set", os.Getenv("PORT"))
		port, err = strconv.Atoi(os.Getenv("PORT"))

		if err != nil {
			return nil, errors.New("invalid port: could not parse integer")
		}
	}

	address := ":" + strconv.Itoa(port)

	if viper.IsSet("address") {
		println("DEPRECATED: address is deprecated, use port")
		address = viper.GetString("address")
	}

	redisOptions := redis.Options{
		Addr:       viper.GetString("redis_host"),
		Password:   viper.GetString("redis_password"),
		DB:         viper.GetInt("redis_db"),
		MaxRetries: viper.GetInt("redis_max_retries"),
	}

	if viper.GetBool("redis_tls") {
		redisOptions.TLSConfig = &tls.Config{}
	}

	directHostname := ""
	directPort := port
	messagingMethod := viper.GetString("messaging_method")

	if messagingMethod == MessageMethodRedis {
		// OK
	} else if messagingMethod == MessageMethodDirect {
		if worker {
			directHostname = viper.GetString("direct_message_hostname")

			if viper.IsSet("direct_message_port") {
				directPort = viper.GetInt("direct_message_port")
			}
		}
	} else {
		return nil, errors.New("invalid messaging method")
	}

	return &DSockOptions{
		Debug:        viper.GetBool("debug"),
		LogRequests:  viper.GetBool("log_requests"),
		RedisOptions: &redisOptions,
		Address:      address,
		Token:        viper.GetString("token"),
		QuitChannel:  make(chan struct{}, 0),
		Jwt: JwtOptions{
			JwtSecret: viper.GetString("jwt_secret"),
		},
		DefaultChannels: UniqueString(RemoveEmpty(
			strings.Split(viper.GetString("default_channels"), ","),
		)),
		MessagingMethod: messagingMethod,
		DirectHostname:  directHostname,
		DirectPort:      directPort,
		Port:            port,
	}, nil
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
