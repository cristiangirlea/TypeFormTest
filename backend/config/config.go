package config

import (
	"cmp"
	"os"
)

type Config struct {
	DBPath string
	Port   string
	TmpDir string
}

func Load() *Config {
	return &Config{
		DBPath: cmp.Or(os.Getenv("DB_PATH"), "forms.db"),
		Port:   cmp.Or(os.Getenv("PORT"), "8080"),
		TmpDir: cmp.Or(os.Getenv("TMP_DIR"), "tmp"),
	}
}
