package config

import (
	"cmp"
	"os"
)

type Config struct {
	DBPath   string
	DBDriver string
	DBURL    string
	RedisURL string
	Port     string
	TmpDir   string
}

func Load() *Config {
	return &Config{
		DBPath:   cmp.Or(os.Getenv("DB_PATH"), "forms.db"),
		DBDriver: cmp.Or(os.Getenv("DB_DRIVER"), "sqlite"),
		DBURL:    os.Getenv("DATABASE_URL"),
		RedisURL: os.Getenv("REDIS_URL"),
		Port:     cmp.Or(os.Getenv("PORT"), "8080"),
		TmpDir:   cmp.Or(os.Getenv("TMP_DIR"), "tmp"),
	}
}
