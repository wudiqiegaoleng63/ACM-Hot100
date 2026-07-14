package config

import "testing"

func TestValidateProductionRejectsUnsafeDefaults(t *testing.T) {
	base := Config{
		AppEnv:           "production",
		AppBaseURL:       "https://acm.example",
		MySQLDSN:         "app:secret@tcp(mysql:3306)/acmhot100",
		RedisPassword:    "redis-secret",
		JWTAccessSecret:  "01234567890123456789012345678901",
		JWTRefreshSecret: "abcdefghijklmnopqrstuvwxyz123456",
	}
	if err := base.ValidateProduction(); err != nil {
		t.Fatalf("safe production config: %v", err)
	}

	cases := []struct {
		name   string
		mutate func(*Config)
	}{
		{name: "default JWT", mutate: func(cfg *Config) { cfg.JWTAccessSecret = "dev-access-secret-change-me" }},
		{name: "empty Redis password", mutate: func(cfg *Config) { cfg.RedisPassword = "" }},
		{name: "root MySQL", mutate: func(cfg *Config) { cfg.MySQLDSN = "root:secret@tcp(mysql:3306)/acmhot100" }},
		{name: "insecure base URL", mutate: func(cfg *Config) { cfg.AppBaseURL = "http://acm.example" }},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			cfg := base
			test.mutate(&cfg)
			if err := cfg.ValidateProduction(); err == nil {
				t.Fatal("expected unsafe production config to be rejected")
			}
		})
	}
}
