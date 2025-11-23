package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
)

// Config holds all configuration for uinputd.
type Config struct {
	// Socket configuration
	Socket SocketConfig `mapstructure:"socket"`

	// Default keyboard layout
	Layout string `mapstructure:"layout"`

	// Performance tuning
	Performance PerformanceConfig `mapstructure:"performance"`

	// Logging configuration
	Logging LoggingConfig `mapstructure:"logging"`
}

// SocketConfig contains Unix socket settings.
type SocketConfig struct {
	Path        string `mapstructure:"path"`
	Permissions uint32 `mapstructure:"permissions"`
}

// PerformanceConfig contains performance tuning parameters.
type PerformanceConfig struct {
	BufferSize        int `mapstructure:"buffer_size"`
	MaxMessageSize    int `mapstructure:"max_message_size"`
	StreamDelayMs     int `mapstructure:"stream_delay_ms"`
	CharDelayMs       int `mapstructure:"char_delay_ms"`
	MaxConcurrentCmds int `mapstructure:"max_concurrent_cmds"`
}

// LoggingConfig contains logging settings.
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"` // "auto", "json", "text"
}

// Load reads configuration from file and environment variables.
// Priority: flags > env vars > config file > defaults
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Config file setup
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// Look for config in standard locations
		v.SetConfigName("uinputd")
		v.SetConfigType("yaml")
		v.AddConfigPath("/etc/uinputd/")
		v.AddConfigPath("$HOME/.config/uinputd/")
		v.AddConfigPath(".")
	}

	// Environment variables
	v.SetEnvPrefix("UINPUTD")
	v.AutomaticEnv()

	// Read config file (optional - don't error if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		// Config file not found, use defaults + env vars
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values.
func setDefaults(v *viper.Viper) {
	// Socket defaults
	v.SetDefault("socket.path", getDefaultSocketPath())
	v.SetDefault("socket.permissions", 0600)

	// Layout defaults
	v.SetDefault("layout", "us")

	// Performance defaults
	v.SetDefault("performance.buffer_size", 4096)
	v.SetDefault("performance.max_message_size", 1048576) // 1MB
	v.SetDefault("performance.stream_delay_ms", 50)
	v.SetDefault("performance.char_delay_ms", 10)
	v.SetDefault("performance.max_concurrent_cmds", 100)

	// Logging defaults
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "auto") // auto-detect TTY
}

// getDefaultSocketPath returns the default Unix socket path.
func getDefaultSocketPath() string {
	// Use runtime dir if available, fallback to /tmp
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		return filepath.Join(runtimeDir, "uinputd.sock")
	}
	return "/tmp/.uinputd.sock"
}

// ParseLogLevel converts string log level to charmbracelet/log level.
func ParseLogLevel(level string) log.Level {
	switch level {
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	case "warn":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	case "fatal":
		return log.FatalLevel
	default:
		return log.InfoLevel
	}
}
