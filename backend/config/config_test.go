package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// 1. Test defaults
	os.Unsetenv("DB_PATH")
	os.Unsetenv("DB_DRIVER")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("REDIS_URL")
	os.Unsetenv("PORT")

	cfg := Load()

	if cfg.DBPath != "forms.db" {
		t.Errorf("expected forms.db, got %s", cfg.DBPath)
	}
	if cfg.DBDriver != "sqlite" {
		t.Errorf("expected sqlite, got %s", cfg.DBDriver)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected 8080, got %s", cfg.Port)
	}

	// 2. Test environment variables
	os.Setenv("PORT", "9090")
	os.Setenv("DB_DRIVER", "postgres")

	cfg = Load()
	if cfg.Port != "9090" {
		t.Errorf("expected 9090, got %s", cfg.Port)
	}
	if cfg.DBDriver != "postgres" {
		t.Errorf("expected postgres, got %s", cfg.DBDriver)
	}
}
