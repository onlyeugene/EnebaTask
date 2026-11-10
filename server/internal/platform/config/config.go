package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv          string
	Port            string
	MongoURI        string
	MongoDB         string
	JWTSecret       []byte
	JWTIssuer       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	CORSOrigins     []string
	ReservationTTL  time.Duration
	BcryptCost      int
}

func (c Config) IsDev() bool { return c.AppEnv == "development" }

// Load reads .env if present (local dev), then the process env.
// godotenv never overrides vars that are already set, so on Render the dashboard values win.
func Load() (Config, error) {
	_ = godotenv.Load()

	var errs []error
	c := Config{
		AppEnv:      getOr("APP_ENV", "development"),
		Port:        getOr("PORT", "8080"),
		MongoURI:    os.Getenv("MONGODB_URI"),
		MongoDB:     os.Getenv("MONGODB_DB"),
		JWTSecret:   []byte(os.Getenv("JWT_SECRET")),
		JWTIssuer:   getOr("JWT_ISSUER", "eugene-api"),
		CORSOrigins: splitList(os.Getenv("CORS_ORIGINS")),
	}
	c.AccessTokenTTL = duration("ACCESS_TOKEN_TTL", "15m", &errs)
	c.RefreshTokenTTL = duration("REFRESH_TOKEN_TTL", "720h", &errs)
	c.ReservationTTL = duration("RESERVATION_TTL", "10m", &errs)

	cost, err := strconv.Atoi(getOr("BCRYPT_COST", "12"))
	if err != nil {
		errs = append(errs, fmt.Errorf("BCRYPT_COST: %w", err))
	}
	c.BcryptCost = cost

	if c.AppEnv != "development" && c.AppEnv != "production" {
		errs = append(errs, fmt.Errorf("APP_ENV must be development or production, got %q", c.AppEnv))
	}
	if c.MongoURI == "" {
		errs = append(errs, errors.New("MONGODB_URI is required"))
	}
	if c.MongoDB == "" {
		errs = append(errs, errors.New("MONGODB_DB is required"))
	}
	// HS256 is only as strong as its key; 32 bytes matches the hash size.
	if len(c.JWTSecret) < 32 {
		errs = append(errs, errors.New("JWT_SECRET must be at least 32 bytes"))
	}
	if len(c.CORSOrigins) == 0 {
		errs = append(errs, errors.New("CORS_ORIGINS is required"))
	}

	return c, errors.Join(errs...)
}

func getOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func duration(key, fallback string, errs *[]error) time.Duration {
	d, err := time.ParseDuration(getOr(key, fallback))
	if err != nil {
		*errs = append(*errs, fmt.Errorf("%s: %w", key, err))
	}
	return d
}

func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	return out
}
