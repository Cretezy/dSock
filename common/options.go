package common

import (
	"crypto/tls"
	"github.com/go-redis/redis/v7"
	"github.com/spf13/viper"
	"os"
	"strings"
)

type JwtOptions struct {
	JwtSecret string
}

type DSockOptions struct {
	RedisOptions *redis.Options
	Address      string
	QuitChannel  chan struct{}
	Debug        bool
	LogRequests  bool
	/// Token for your API -> dSock and between dSock services
	Token string
	/// JWT parsing/verifying options
	Jwt JwtOptions
	/// Default channels to subscribe on join
	DefaultChannels []string
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
	viper.SetDefault("address", ":6241")
	viper.SetDefault("default_channels", "")
	viper.SetDefault("token", "")
	viper.SetDefault("jwt_secret", "")
	viper.SetDefault("debug", false)

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

func GetOptions() (*DSockOptions, error) {
	err := SetupConfig()
	if err != nil {
		return nil, err
	}

	port := os.Getenv("PORT")

	address := ":" + port

	if viper.GetString("address") != "" {
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
	}, nil
}
