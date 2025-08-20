package config

import (
	"strings"

	"github.com/spf13/viper"
)

type AppCfg struct {
	Name string
	Env  string
	Host string
	Port int
}

type LogCfg struct {
	Level string
}

type DBCfg struct {
	DSN         string
	MaxOpen     int
	MaxIdle     int
	AutoMigrate bool
}

type RedisCfg struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

type MQCfg struct {
	URL      string
	Queue    string
	Prefetch int
}

type S3Cfg struct {
	Endpoint         string
	Region           string
	AccessKey        string
	SecretKey        string
	Bucket           string
	UsePathStyle     bool
	PresignExpireSec int
	SSE              string
}

type Config struct {
	App      AppCfg
	Log      LogCfg
	Database DBCfg
	Redis    RedisCfg
	RabbitMQ MQCfg
	S3       S3Cfg
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("APP")

	_ = v.ReadInConfig()

	v.SetDefault("app.host", "0.0.0.0")
	v.SetDefault("app.port", 8080)
	v.SetDefault("log.level", "info")
	v.SetDefault("redis.poolSize", 10)
	v.SetDefault("rabbitmq.prefetch", 10)
	v.SetDefault("s3.region", "auto")
	v.SetDefault("s3.usePathStyle", true)
	v.SetDefault("s3.presignExpireSec", 900)

	cfg := new(Config)
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
