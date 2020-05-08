package common

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/spf13/viper"
	"log"
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
	/// Token for your API -> dSock and between dSock services
	Token string
	/// JWT parsing/verifying options
	Jwt JwtOptions
	/// Default channels to subscribe on join
	DefaultChannels []string
}

func SetupConfig() {
	viper.SetConfigName("config")
	viper.SetEnvPrefix("DSOCK")
	viper.AutomaticEnv()

	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config/dsock")
	viper.AddConfigPath("/etc/dsock")

	viper.SetDefault("redis_host", "localhost:6379")
	viper.SetDefault("redis_password", "")
	viper.SetDefault("redis_db", 0)
	viper.SetDefault("address", ":6241")
	viper.SetDefault("default_channels", "")
	viper.SetDefault("token", "")
	viper.SetDefault("jwt_secret", "")
	viper.SetDefault("debug", false)

	err := viper.ReadInConfig()

	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No config found, writing new config.")

			err = viper.SafeWriteConfigAs("config.toml")

			if err != nil {
				panic(fmt.Errorf("Fatal error saving default config file: %s \n", err))
			}
		} else {
			panic(fmt.Errorf("Fatal error reading config file: %s \n", err))
		}
	}
}

func GetOptions() DSockOptions {
	SetupConfig()

	port := os.Getenv("PORT")

	address := ":" + port

	if viper.GetString("address") != "" {
		address = viper.GetString("address")
	}

	return DSockOptions{
		RedisOptions: &redis.Options{
			Addr:     viper.GetString("redis_host"),
			Password: viper.GetString("redis_password"),
			DB:       viper.GetInt("redis_db"),
		},
		Address:     address,
		Token:       viper.GetString("token"),
		QuitChannel: make(chan struct{}, 0),
		Jwt: JwtOptions{
			JwtSecret: viper.GetString("jwt_secret"),
		},
		DefaultChannels: UniqueString(RemoveEmpty(
			strings.Split(viper.GetString("default_channels"), ","),
		)),
	}
}
