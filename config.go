package crayon

import (
	"time"
	"os"
	"io/ioutil"
	"encoding/json"
	"strconv"
)


const (
	defaultHost = "127.0.0.1"
	defaultPort = "8080"
	defaultReadTimeout = 10
	defaultWriteTimeout = 10
	defaultShutdownTimeout = 10
)

type Config struct {
	Host string `json:"host"`
	Port string `json:"port"`
	CertFile string `json:"cert_file"`
	KeyFile string `json:"key_file"`
	HTTPSPort string `json:"http_port"`

	InsecureSkipVerify bool `json:"insecure_skip_verify"`
	ReadTimeout time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout"`
}

func NewConfig(filePath string) (Config, error) {

	var cfg Config

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return cfg, ErrInvalidFilePath
	}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	if cfg.Host == "" {
		cfg.Host = defaultHost
	}
	if cfg.Port == "" {
		cfg.Port = defaultPort
	} else {
		if port, err := strconv.Atoi(cfg.Port); err != nil && (port <= 0 || port > 65535) {
			return cfg, ErrInvalidPort
		}

	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = defaultReadTimeout
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = defaultWriteTimeout
	}
	if cfg.ShutdownTimeout == 0 {
		cfg.ShutdownTimeout = defaultShutdownTimeout
	}
	return cfg, nil
}