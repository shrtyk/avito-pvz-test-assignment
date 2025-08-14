package postgres

import (
	"database/sql"
	"fmt"
	"net/url"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/shrtyk/avito-backend-spring-2025/pkg/config"
)

func MustCreateConnectionPool(cfg *config.Config) *sql.DB {
	dsn := buildDSN(cfg)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		msg := fmt.Sprintf("Coludn't create connections pool for DSN: '%s': %s", dsn, err)
		panic(msg)
	}

	db.SetMaxOpenConns(cfg.PostgresCfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.PostgresCfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.PostgresCfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.PostgresCfg.ConnMaxIdletime)

	if err = db.Ping(); err != nil {
		panic(fmt.Sprintf("Couldn't ping DB: %s", err))
	}

	return db
}

func buildDSN(cfg *config.Config) string {
	url := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.PostgresCfg.User, cfg.PostgresCfg.Password),
		Host:   fmt.Sprintf("%s:%s", cfg.PostgresCfg.Host, cfg.PostgresCfg.Port),
		Path:   cfg.PostgresCfg.DBName,
	}

	q := url.Query()
	q.Set("sslmode", cfg.PostgresCfg.SSLMode)
	url.RawQuery = q.Encode()

	return url.String()
}
