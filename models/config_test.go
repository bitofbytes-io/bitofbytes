package models

import (
	"os"
	"testing"
)

func withTempEnv(env map[string]string, fn func(t *testing.T)) func(t *testing.T) {
	return func(t *testing.T) {
		t.Helper()
		original := make(map[string]string, len(env))
		for k := range env {
			original[k] = os.Getenv(k)
		}
		for k, v := range env {
			if v == "" {
				os.Unsetenv(k)
				continue
			}
			os.Setenv(k, v)
		}
		t.Cleanup(func() {
			for k, v := range original {
				if v == "" {
					os.Unsetenv(k)
					continue
				}
				os.Setenv(k, v)
			}
		})
		fn(t)
	}
}

func TestLoadEnvConfigRequiresEnv(t *testing.T) {
	t.Run("errors when server address missing", withTempEnv(map[string]string{
		"SERVER_ADDRESS": "",
		"CSRF_SECURE":    "true",
	}, func(t *testing.T) {
		if _, err := LoadEnvConfig(); err == nil {
			t.Fatalf("expected error when SERVER_ADDRESS missing")
		}
	}))

	t.Run("errors when csrf secure missing", withTempEnv(map[string]string{
		"SERVER_ADDRESS": "127.0.0.1:8080",
		"CSRF_SECURE":    "",
	}, func(t *testing.T) {
		if _, err := LoadEnvConfig(); err == nil {
			t.Fatalf("expected error when CSRF_SECURE missing")
		}
	}))

	t.Run("loads config from env", withTempEnv(map[string]string{
		"SERVER_ADDRESS": "127.0.0.1:8080",
		"CSRF_SECURE":    "false",
	}, func(t *testing.T) {
		cfg, err := LoadEnvConfig()
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if cfg.Server.Address != "127.0.0.1:8080" {
			t.Fatalf("expected server address from env, got %q", cfg.Server.Address)
		}
		if cfg.CSRF.Secure {
			t.Fatalf("expected CSRF secure false from env")
		}
	}))
}
