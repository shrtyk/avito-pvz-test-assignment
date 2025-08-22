package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/joho/godotenv/autoload"
)

var path string

func init() {
	flag.StringVar(&path, "cfg_path", "", "Path to connfig file")
}

type Config struct {
	AppCfg        AppCfg        `yaml:"app"`
	HttpServerCfg HttpServerCfg `yaml:"http_server"`
	GrpcServerCfg GrpcServerCfg `yaml:"grpc_server"`
	PostgresCfg   PostgresCfg   `yaml:"postgres"`
	AuthTokenCfg  AuthTokensCfg `yaml:"auth_tokens"`
}

type AppCfg struct {
	Env             string        `yaml:"env" env:"APP_ENV" env-default:"prod"`
	Timeout         time.Duration `yaml:"timeout" env:"APP_TIMEOUT" env-default:"5s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env:"APP_SHUTDOWN_TIMEOUT" env-default:"10s"`
}

type HttpServerCfg struct {
	Port         string        `yaml:"port" env:"HTTP_SERVER_PORT" env-default:"8080"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env:"HTTP_SERVER_IDLE_TIMEOUT" env-default:"5s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"HTTP_SERVER_WRITE_TIMEOUT" env-default:"10s"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env:"HTTP_SERVER_READ_TIMEOUT" env-default:"10s"`
}

type GrpcServerCfg struct {
	Port string `yaml:"port" env:"GRPC_SERVER_PORT" env-default:"3000"`
}

type PostgresCfg struct {
	User     string `yaml:"user" env:"PG_USER" env-default:"user"`
	Password string `yaml:"password" env:"PG_PASSWORD" env-default:"password"`
	Host     string `yaml:"host" env:"PG_HOST" env-default:"postgres"`
	Port     string `yaml:"port" env:"PG_PORT" env-default:"5432"`
	DBName   string `yaml:"db_name" env:"PG_DBNAME" env-default:"pvz-db"`
	SSLMode  string `yaml:"sslmode" env:"PG_SSLMODE" env-default:"disable"`

	MaxOpenConns    int           `yaml:"max_open_conns" env:"PG_MAX_OPEN_CONNS" env-default:"20"`
	MaxIdleConns    int           `yaml:"max_idle_conns" env:"PG_MAX_IDLE_CONS" env-default:"10"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetine" env:"PG_CONN_MAX_LIFETIME" env-default:"30m"`
	ConnMaxIdletime time.Duration `yaml:"conn_max_idletime" env:"PG_CONN_MAX_IDLETIME" env-default:"5m"`
}

type AuthTokensCfg struct {
	PublicRSAPath   string        `yaml:"public_key_path" env:"PUBLIC_RSA_PATH" env-default:"./keys/rsa/public_key.pem"`
	PrivateRSAPath  string        `yaml:"private_key_path" env:"PRIVATE_RSA_PATH" env-default:"./keys/rsa/private_key.pem"`
	JWTLifetime     time.Duration `yaml:"jwt_lifetime" env:"JWT_LIFETIME" env-default:"15m"`
	RefreshLifetime time.Duration `yaml:"refresh_lifetime" env:"REFRESH_LIFETIME" env-default:"720h"`
	SecretKey       string        `yaml:"secret_key" env:"SECRET_KEY" env-default:"super-secret-key"`
}

func MustInitConfig() *Config {
	cfgPath := cfgPath()
	cfg := new(Config)

	if cfgPath != "" {
		err := cleanenv.ReadConfig(cfgPath, cfg)
		if err != nil && !os.IsNotExist(err) {
			panic(fmt.Sprintf("failed to read config file: %s", err))
		}
	}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		panic(fmt.Sprintf("failed to read environment variables: %s", err))
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
