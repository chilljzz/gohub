package config

import (
	"errors"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Mysql  MySQLConfig  `mapstructure:"mysql"`
	JWT    JWTConfig    `mapstructure:"jwt"`
	Redis  RedisConfig  `mapstructure:"redis"`
}
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	Charset  string `mapstructure:"charset"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

var Conf Config

func InitConfig() error {

	if err := godotenv.Load(); err != nil {

	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.SetEnvPrefix("GOHUB")

	viper.SetEnvKeyReplacer(
		strings.NewReplacer(
			".",
			"_",
		),
	)

	viper.AutomaticEnv()

	bindings := map[string]string{
		"server.port":      "GOHUB_SERVER_PORT",
		"server.mode":      "GOHUB_SERVER_MODE",
		"mysql.host":       "GOHUB_MYSQL_HOST",
		"mysql.port":       "GOHUB_MYSQL_PORT",
		"mysql.username":   "GOHUB_MYSQL_USERNAME",
		"mysql.password":   "GOHUB_MYSQL_PASSWORD",
		"mysql.database":   "GOHUB_MYSQL_DATABASE",
		"mysql.charset":    "GOHUB_MYSQL_CHARSET",
		"redis.addr":       "GOHUB_REDIS_ADDR",
		"redis.password":   "GOHUB_REDIS_PASSWORD",
		"redis.db":         "GOHUB_REDIS_DB",
		"jwt.secret":       "GOHUB_JWT_SECRET",
		"jwt.expire_hours": "GOHUB_JWT_EXPIRE_HOURS",
	}

	for key, env := range bindings {
		if err := viper.BindEnv(key, env); err != nil {
			return err
		}
	}

	if err := viper.ReadInConfig(); err != nil {
		var notFoundErr viper.ConfigFileNotFoundError

		if !errors.As(err, &notFoundErr) {
			return err
		}
	}

	if err := viper.Unmarshal(&Conf); err != nil {
		return err
	}

	return nil
}
