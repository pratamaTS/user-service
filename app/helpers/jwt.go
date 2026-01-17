package helpers

import (
	"os"
	"time"
)

func ProvideJWTSecret() string {
	if v := os.Getenv("JWT_SECRET"); v != "" {
		return v
	}
	return "dev-secret"
}

func ProvideAccessTTL() time.Duration {
	if v := os.Getenv("JWT_ACCESS_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	// default: 30 minutes
	return 30 * time.Minute
}

func ProvideRefreshTTL() time.Duration {
	if v := os.Getenv("JWT_REFRESH_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	// default: 7 days
	return 7 * 24 * time.Hour
}
