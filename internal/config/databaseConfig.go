package config

import (
	"fmt"
	"os"
)

type DatabaseConfig struct {
	Path string
}

func Load() DatabaseConfig {
	path := os.Getenv("DATABASE_PATH")
	if path == "" {
		path = "data/app.db"
	}
	return DatabaseConfig{Path: path}
}

func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s?%s", c.Path,
		"_busy_timeout=5000"+
			"&_journal_mode=WAL"+ // allows concurrent reads while writing
			"&_foreign_keys=ON", // enforce FK constraints (off by default in SQLite)
	)
}
