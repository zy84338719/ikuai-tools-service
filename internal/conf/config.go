package conf

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Log      LogConfig      `mapstructure:"log"`
	App      AppConfig      `mapstructure:"app"`
	IKuai    IKuaiConfig    `mapstructure:"ikuai"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
	Jobs     JobsConfig     `mapstructure:"jobs"`
}

type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

func (c *DatabaseConfig) DSN() string {
	switch c.Driver {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.User, c.Password, c.Host, c.Port, c.DBName)
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
	case "sqlite":
		return c.DBName
	default:
		return ""
	}
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

type IKuaiConfig struct {
	BaseURL  string `mapstructure:"base_url"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Timeout  int    `mapstructure:"timeout"`
}

type MetricsConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Namespace string `mapstructure:"namespace"`
	Path      string `mapstructure:"path"`
	Port      int    `mapstructure:"port"` // default 9100
}

type JobsConfig struct {
	SkipStart    bool                     `mapstructure:"skip_start"`
	GhProxy      string                   `mapstructure:"gh_proxy"`
	MaxPerRecord MaxPerRecordConfig       `mapstructure:"max_per_record"`
	CustomISP    []CronCustomISPConfig    `mapstructure:"custom_isp"`
	StreamDomain []CronStreamDomainConfig `mapstructure:"stream_domain"`
}

// MaxPerRecordConfig controls the maximum number of entries per iKuai rule record.
// These match the limits used by ikuai-bypass and the iKuai router firmware.
type MaxPerRecordConfig struct {
	ISP    int `mapstructure:"isp"`    // default 5000
	Domain int `mapstructure:"domain"` // default 1000
}

func (c *MaxPerRecordConfig) ISPLimit() int {
	if c.ISP <= 0 {
		return 5000
	}
	return c.ISP
}

func (c *MaxPerRecordConfig) DomainLimit() int {
	if c.Domain <= 0 {
		return 1000
	}
	return c.Domain
}

type CronCustomISPConfig struct {
	Cron    string   `mapstructure:"cron"`
	Tag     string   `mapstructure:"tag"`
	Name    string   `mapstructure:"name"`    // legacy alias for Tag
	Url     []string `mapstructure:"url"`
	Comment string   `mapstructure:"comment"` // written as chunk comment (auto-generated if empty)
}

// GetTag returns the effective tag, falling back to Name for backward compatibility.
func (c *CronCustomISPConfig) GetTag() string {
	if c.Tag != "" {
		return c.Tag
	}
	return c.Name
}

type CronStreamDomainConfig struct {
	Cron               string   `mapstructure:"cron"`
	Tag                string   `mapstructure:"tag"`
	Interface          []string `mapstructure:"interface"`
	Url                []string `mapstructure:"url"`
	SrcAddr            string   `mapstructure:"src_addr"`
	SrcAddrOptIpGroup  string   `mapstructure:"src_addr_opt_ipgroup"` // IP group name used as src filter
	Comment            string   `mapstructure:"comment"`              // legacy; Tag is preferred
}

// GetTag returns the effective tag, falling back to Comment for backward compatibility.
func (c *CronStreamDomainConfig) GetTag() string {
	if c.Tag != "" {
		return c.Tag
	}
	return c.Comment
}

var GlobalConfig *Config

func Init(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	GlobalConfig = &Config{}
	if err := viper.Unmarshal(GlobalConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return GlobalConfig.Validate()
}

// Validate checks that critical config values are sensible.
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port %d is out of range (1-65535)", c.Server.Port)
	}
	if c.IKuai.BaseURL != "" {
		if u, err := url.Parse(c.IKuai.BaseURL); err != nil || !strings.HasPrefix(u.Scheme, "http") {
			return fmt.Errorf("ikuai.base_url %q is not a valid HTTP URL", c.IKuai.BaseURL)
		}
	}
	for i, job := range c.Jobs.CustomISP {
		if job.GetTag() == "" {
			return fmt.Errorf("jobs.custom_isp[%d]: tag is required", i)
		}
		if job.Cron == "" {
			return fmt.Errorf("jobs.custom_isp[%d]: cron is required", i)
		}
	}
	for i, job := range c.Jobs.StreamDomain {
		if job.GetTag() == "" {
			return fmt.Errorf("jobs.stream_domain[%d]: tag is required", i)
		}
		if job.Cron == "" {
			return fmt.Errorf("jobs.stream_domain[%d]: cron is required", i)
		}
	}
	return nil
}

func InitWithDefault() error {
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8888)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)
	viper.SetDefault("database.driver", "mysql")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.max_size", 100)
	viper.SetDefault("log.max_backups", 10)
	viper.SetDefault("log.max_age", 30)
	viper.SetDefault("log.compress", true)
	viper.SetDefault("app.name", "ikuai-tools-service")
	viper.SetDefault("app.version", "1.0.0")
	viper.SetDefault("ikuai.base_url", "http://192.168.1.1")
	viper.SetDefault("ikuai.username", "admin")
	viper.SetDefault("ikuai.password", "")
	viper.SetDefault("ikuai.timeout", 30)
	viper.SetDefault("metrics.enabled", true)
	viper.SetDefault("metrics.namespace", "ikuai")
	viper.SetDefault("metrics.path", "/metrics")
	viper.SetDefault("metrics.port", 9100)

	GlobalConfig = &Config{}
	if err := viper.Unmarshal(GlobalConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return GlobalConfig.Validate()
}
