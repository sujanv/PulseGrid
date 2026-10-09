package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server    ServerConfig
	Pool      PoolConfig
	Scheduler SchedulerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	Alert     AlertConfig
	Metrics   MetricsConfig
}

type ServerConfig struct {
	Port         int
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	APIKey       string
}

type PoolConfig struct {
	WorkerCount     int
	QueueSize       int
	DefaultTimeout  time.Duration
	PerHostRateLimit float64 // requests per second
}

type SchedulerConfig struct {
	TickInterval  time.Duration
	JitterPercent float64
	BatchSize     int
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Addr      string
	Password  string
	DB        int
	LockTTL   time.Duration
	KeyPrefix string
}

type AlertConfig struct {
	WebhookTimeout time.Duration
	MaxRetries     int
	SMTPHost       string
	SMTPPort       int
	SMTPUser       string
	SMTPPass       string
	FromEmail      string
}

type MetricsConfig struct {
	Enabled bool
	Port    int
	Path    string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnvInt("PULSEGRID_PORT", 8080),
			Host:         getEnv("PULSEGRID_HOST", "0.0.0.0"),
			ReadTimeout:  getEnvDuration("PULSEGRID_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getEnvDuration("PULSEGRID_WRITE_TIMEOUT", 10*time.Second),
			APIKey:       getEnv("PULSEGRID_API_KEY", ""),
		},
		Pool: PoolConfig{
			WorkerCount:      getEnvInt("PULSEGRID_WORKER_COUNT", 50),
			QueueSize:        getEnvInt("PULSEGRID_QUEUE_SIZE", 5000),
			DefaultTimeout:   getEnvDuration("PULSEGRID_CHECK_TIMEOUT", 10*time.Second),
			PerHostRateLimit: getEnvFloat("PULSEGRID_HOST_RATE_LIMIT", 5.0),
		},
		Scheduler: SchedulerConfig{
			TickInterval:  getEnvDuration("PULSEGRID_SCHEDULER_TICK", 1*time.Second),
			JitterPercent: getEnvFloat("PULSEGRID_SCHEDULER_JITTER", 0.10),
			BatchSize:     getEnvInt("PULSEGRID_BATCH_SIZE", 200),
		},
		Database: DatabaseConfig{
			URL:             getEnv("PULSEGRID_DB_URL", "postgres://postgres:postgres@localhost:5432/pulsegrid?sslmode=disable"),
			MaxOpenConns:    getEnvInt("PULSEGRID_DB_MAX_OPEN", 25),
			MaxIdleConns:    getEnvInt("PULSEGRID_DB_MAX_IDLE", 10),
			ConnMaxLifetime: getEnvDuration("PULSEGRID_DB_LIFETIME", 5*time.Minute),
		},
		Redis: RedisConfig{
			Addr:      getEnv("PULSEGRID_REDIS_ADDR", "localhost:6379"),
			Password:  getEnv("PULSEGRID_REDIS_PASS", ""),
			DB:        getEnvInt("PULSEGRID_REDIS_DB", 0),
			LockTTL:   getEnvDuration("PULSEGRID_REDIS_LOCK_TTL", 15*time.Second),
			KeyPrefix: getEnv("PULSEGRID_REDIS_PREFIX", "pulsegrid:"),
		},
		Alert: AlertConfig{
			WebhookTimeout: getEnvDuration("PULSEGRID_ALERT_TIMEOUT", 5*time.Second),
			MaxRetries:     getEnvInt("PULSEGRID_ALERT_MAX_RETRIES", 3),
			SMTPHost:       getEnv("PULSEGRID_SMTP_HOST", ""),
			SMTPPort:       getEnvInt("PULSEGRID_SMTP_PORT", 587),
			SMTPUser:       getEnv("PULSEGRID_SMTP_USER", ""),
			SMTPPass:       getEnv("PULSEGRID_SMTP_PASS", ""),
			FromEmail:      getEnv("PULSEGRID_ALERT_FROM", "alerts@pulsegrid.io"),
		},
		Metrics: MetricsConfig{
			Enabled: getEnvBool("PULSEGRID_METRICS_ENABLED", true),
			Port:    getEnvInt("PULSEGRID_METRICS_PORT", 9090),
			Path:    getEnv("PULSEGRID_METRICS_PATH", "/metrics"),
		},
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvFloat(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		lower := strings.ToLower(val)
		return lower == "true" || lower == "1" || lower == "yes"
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}
