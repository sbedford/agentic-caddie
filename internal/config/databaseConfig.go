package config

import "fmt"

type DatabaseConfig struct {
	Path string
}

func Load() DatabaseConfig {
	return DatabaseConfig{
		Path: "data/app.db",
	}
}

func (c DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s?%s", c.Path,
		"_busy_timeout=5000"+
			"&_journal_mode=WAL"+ // allows concurrent reads while writing
			"&_foreign_keys=ON", // enforce FK constraints (off by default in SQLite)
	)
}
