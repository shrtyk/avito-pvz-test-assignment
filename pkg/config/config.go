package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

var path string

func init() {
	flag.StringVar(&path, "cfg_path", "", "Path to connfig file")

	_ = godotenv.Load()
}

type Config struct {
	PvzCfg        pvzCfg        `yaml:"pvz"`
	HttpServerCfg httpServerCfg `yaml:"http_server"`
	GrpcServerCfg grpcServerCfg `yaml:"grpc_server"`
	PostgresCfg   postgresCfg   `yaml:"postgres"`
}

type pvzCfg struct {
	Env     string        `yaml:"env" env:"PVZ_ENV" env-default:"prod"`
	Timeout time.Duration `yaml:"timeout" env:"PVZ_TIMEOUT" env-default:"5s"`
}

type httpServerCfg struct {
	Host         string        `yaml:"host" env:"HTTP_SERVER_HOST" env-default:"localhost"`
	Port         string        `yaml:"port" env:"HTTP_SERVER_PORT" env-default:"16700"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env:"HTTP_SERVER_IDLE_TIMEOUT" env-default:"5s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"HTTP_SERVER_WRITE_TIMEOUT" env-default:"10s"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env:"HTTP_SERVER_READ_TIMEOUT" env-default:"10s"`
}

type grpcServerCfg struct {
	Host string `yaml:"host" env:"GRPC_SERVER_HOST" env-default:"localhost"`
	Port string `yaml:"port" env:"GRPC_SERVER_PORT" env-default:"16701"`
}

type postgresCfg struct {
	User     string `yaml:"user" env:"PG_USER" env-default:"user"`
	Password string `yaml:"password" env:"PG_PASSWORD" env-default:"password"`
	Host     string `yaml:"host" env:"PG_HOST" env-default:"postgres"`
	Port     string `yaml:"port" env:"PG_PORT" env-default:"5432"`
	DBName   string `yaml:"db_name" env:"PG_DBNAME" env-default:"pvz-db"`
	SSLMode  string `yaml:"sslmode" env:"PG_SSL_MODE" env-default:"disable"`

	MaxOpenConns    int           `yaml:"max_open_conns" env:"PG_MAX_OPEN_CONNS" env-default:"20"`
	MaxIdleConns    int           `yaml:"max_idle_conns" env:"PG_MAX_IDLE_CONS" env-default:"10"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetine" env:"PG_CONN_MAX_LIFETIME" env-default:"30m"`
	ConnMaxIdletime time.Duration `yaml:"conn_max_idletime" env:"PG_CONN_MAX_IDLETIME" env-default:"5m"`
}

func MustInitConfig() *Config {
	cfgPath := cfgPath()

	cfg := new(Config)
	if err := cleanenv.ReadConfig(cfgPath, cfg); err != nil {
		if envErr := cleanenv.ReadEnv(cfg); envErr != nil {
			msg := fmt.Sprintf(
				"couldn't read config data from config file or environment variables: %s: %s", err, envErr)
			panic(msg)
		}
	}
	return cfg
}

func cfgPath() string {
	if !flag.Parsed() {
		flag.Parse()
	}

	if path == "" {
		return os.Getenv("CONFIG_PATH")
	}

	return path
}
