package config

import (
	"testing"
	"time"
)

func TestLoadMissingMySQLDSN(t *testing.T) {
	t.Setenv("MYSQL_DSN", "")

	cfg, err := Load()
	if err == nil {
		t.Fatalf("expected error, got nil (cfg=%v)", cfg)
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("MYSQL_DSN", "user:pass@tcp(localhost:3306)/profile")
	t.Setenv("HTTP_HOST", "")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("GRPC_HOST", "")
	t.Setenv("GRPC_PORT", "")
	t.Setenv("MYSQL_MAX_OPEN_CONNS", "")
	t.Setenv("MYSQL_MAX_IDLE_CONNS", "")
	t.Setenv("MYSQL_CONN_MAX_LIFETIME_MINUTES", "")
	t.Setenv("LOG_LEVEL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.HTTPHost != "0.0.0.0" || cfg.HTTPPort != "8080" {
		t.Fatalf("unexpected HTTP defaults: %+v", cfg)
	}
	if cfg.GRPCHost != "0.0.0.0" || cfg.GRPCPort != "9090" {
		t.Fatalf("unexpected gRPC defaults: %+v", cfg)
	}
	if cfg.MySQLMaxOpen != 10 || cfg.MySQLMaxIdle != 5 {
		t.Fatalf("unexpected MySQL pool defaults: %+v", cfg)
	}
	if cfg.MySQLMaxLife != 30*time.Minute {
		t.Fatalf("unexpected MySQL max life default: %v", cfg.MySQLMaxLife)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("expected LOG_LEVEL default 'info', got %q", cfg.LogLevel)
	}
}

func TestLoadCustomValues(t *testing.T) {
	t.Setenv("MYSQL_DSN", "dsn")
	t.Setenv("HTTP_HOST", "127.0.0.1")
	t.Setenv("HTTP_PORT", "8081")
	t.Setenv("GRPC_HOST", "127.0.0.2")
	t.Setenv("GRPC_PORT", "9091")
	t.Setenv("MYSQL_MAX_OPEN_CONNS", "42")
	t.Setenv("MYSQL_MAX_IDLE_CONNS", "12")
	t.Setenv("MYSQL_CONN_MAX_LIFETIME_MINUTES", "17")
	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.HTTPHost != "127.0.0.1" || cfg.HTTPPort != "8081" {
		t.Fatalf("unexpected HTTP config: %+v", cfg)
	}
	if cfg.GRPCHost != "127.0.0.2" || cfg.GRPCPort != "9091" {
		t.Fatalf("unexpected gRPC config: %+v", cfg)
	}
	if cfg.MySQLMaxOpen != 42 || cfg.MySQLMaxIdle != 12 {
		t.Fatalf("unexpected MySQL pool config: %+v", cfg)
	}
	if cfg.MySQLMaxLife != 17*time.Minute {
		t.Fatalf("unexpected MySQL max life: %v", cfg.MySQLMaxLife)
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("unexpected LOG_LEVEL: %q", cfg.LogLevel)
	}
}

func TestGetIntAndDurationFallback(t *testing.T) {
	t.Setenv("BROKEN_INT", "x")
	t.Setenv("BROKEN_MIN", "y")

	if got := getIntEnv("BROKEN_INT", 7); got != 7 {
		t.Fatalf("expected fallback int 7, got %d", got)
	}
	if got := getDurationEnv("BROKEN_MIN", 3*time.Minute); got != 3*time.Minute {
		t.Fatalf("expected fallback duration 3m, got %v", got)
	}
}
