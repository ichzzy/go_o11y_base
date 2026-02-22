package config

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/viper"
)

type Config struct {
	App           AppConfig           `mapstructure:"app"`
	Server        ServerConfig        `mapstructure:"server"`
	Observability ObservabilityConfig `mapstructure:"observability"`
	Connections   ConnectionsConfig   `mapstructure:"connections"`

	JWT struct {
		PublicKey        ed25519.PublicKey  `mapstructure:"-"`
		PublicKeyBase64  string             `mapstructure:"publicKeyBase64"`
		PrivateKey       ed25519.PrivateKey `mapstructure:"-"`
		PrivateKeyBase64 string             `mapstructure:"privateKeyBase64"`
		Duration         time.Duration      `mapstructure:"duration"`
		RefreshDuration  time.Duration      `mapstructure:"refreshDuration"`
	} `mapstructure:"jwt"`
}

type AppConfig struct {
	Env         string `mapstructure:"env"`
	ServiceName string `mapstructure:"serviceName"`
	PodName     string `mapstructure:"podName"`
	Log         struct {
		Pretty       bool          `mapstructure:"pretty"`
		Level        string        `mapstructure:"level"`
		Path         string        `mapstructure:"path"`
		MaxAge       time.Duration `mapstructure:"maxAge"`
		RotationTime time.Duration `mapstructure:"rotationTime"`
	} `mapstructure:"log"`
}

type ServerConfig struct {
	Http struct {
		Port    int `mapstructure:"port"`
		Timeout struct {
			ReadTimeout       time.Duration `mapstructure:"readTimeout"`
			ReadHeaderTimeout time.Duration `mapstructure:"readHeaderTimeout"`
			WriteTimeout      time.Duration `mapstructure:"writeTimeout"`
			IdleTimeout       time.Duration `mapstructure:"idleTimeout"`
		} `mapstructure:"timeout"`
	} `mapstructure:"http"`
	Grpc struct {
		Port          int  `mapstructure:"port"`
		EnableReflect bool `mapstructure:"enableReflect"`
	} `mapstructure:"grpc"`
}

type ObservabilityConfig struct {
	Pprof     PprofConfig     `mapstructure:"pprof"`
	OTEL      OTELConfig      `mapstructure:"otel"`
	Pyroscope PyroscopeConfig `mapstructure:"pyroscope"`
}

type OTELConfig struct {
	Enabled     bool    `mapstructure:"enabled"`
	Endpoint    string  `mapstructure:"endpoint"`
	SampleRatio float64 `mapstructure:"sampleRatio"`
}

type PyroscopeConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Endpoint string `mapstructure:"endpoint"`
}

type PprofConfig struct {
	Enabled      bool          `mapstructure:"enabled"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"readTimeout"`
	WriteTimeout time.Duration `mapstructure:"writeTimeout"`
	IdleTimeout  time.Duration `mapstructure:"idleTimeout"`
}

type ConnectionsConfig struct {
	Mysql MySQLConfig `mapstructure:"mysql"`
	Redis RedisConfig `mapstructure:"redis"`
	Kafka struct{}    `mapstructure:"kafka"`
}

type MySQLConfig struct {
	Gorm struct {
		LogEnabled       bool          `mapstructure:"logEnabled"`
		LogSlowThreshold time.Duration `mapstructure:"logSlowThreshold"`
	} `mapstructure:"gorm"`
	Master struct {
		Host            string        `mapstructure:"host"`
		Port            int           `mapstructure:"port"`
		Username        string        `mapstructure:"username"`
		Password        string        `mapstructure:"password"`
		Name            string        `mapstructure:"name"`
		MaxIdleConns    int           `mapstructure:"maxIdleConns"`
		MaxOpenConns    int           `mapstructure:"maxOpenConns"`
		ConnMaxLifetime time.Duration `mapstructure:"connMaxLifetime"`
	} `mapstructure:"master"`
}

type RedisConfig struct {
	Cluster struct {
		Addrs           []string      `mapstructure:"addrs"`
		Username        string        `mapstructure:"username"`
		Password        string        `mapstructure:"password"`
		DialTimeout     time.Duration `mapstructure:"dialTimeout"`
		ReadTimeout     time.Duration `mapstructure:"readTimeout"`
		WriteTimeout    time.Duration `mapstructure:"writeTimeout"`
		PoolSize        int           `mapstructure:"poolSize"`
		MinIdleConns    int           `mapstructure:"minIdleConns"`
		MaxIdleConns    int           `mapstructure:"maxIdleConns"`
		ConnMaxIdleTime time.Duration `mapstructure:"connMaxIdleTime"`
	} `mapstructure:"cluster"`
}

func Load() (*Config, error) {
	viper.AddConfigPath("./config")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	var c Config
	if err := viper.ReadInConfig(); err != nil {
		return &c, fmt.Errorf("viper.ReadInConfig failed: %w", err)
	}
	if err := viper.Unmarshal(&c); err != nil {
		return &c, fmt.Errorf("viper.Unmarshal failed: %w", err)
	}

	if err := parseJwtKeyPair(&c); err != nil {
		return &c, fmt.Errorf("parseJwtKeyPair failed: %w", err)
	}

	return &c, nil
}

func parseJwtKeyPair(c *Config) error {
	privPEM, err := base64.StdEncoding.DecodeString(c.JWT.PrivateKeyBase64)
	if err != nil {
		return fmt.Errorf("decode private key base64 failed: %w", err)
	}
	block, _ := pem.Decode(privPEM)
	if block == nil {
		return fmt.Errorf("decode private key pem block failed")
	}
	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("parse ed25519 private key failed: %w", err)
	}
	edPriv, ok := priv.(ed25519.PrivateKey)
	if !ok {
		return fmt.Errorf("not an ed25519 private key")
	}
	c.JWT.PrivateKey = edPriv

	pubPEM, err := base64.StdEncoding.DecodeString(c.JWT.PublicKeyBase64)
	if err != nil {
		return fmt.Errorf("decode public key base64 failed: %w", err)
	}
	block, _ = pem.Decode(pubPEM)
	if block == nil {
		return fmt.Errorf("decode public key pem block failed")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("parse ed25519 public key failed: %w", err)
	}
	edPub, ok := pub.(ed25519.PublicKey)
	if !ok {
		return fmt.Errorf("not an ed25519 public key")
	}
	c.JWT.PublicKey = edPub

	return nil
}
