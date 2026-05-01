package config

import (
	"fmt"
	"os"
	"time"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Server struct {
		Port         int           `yaml:"port"`
		ReadTimeout  time.Duration `yaml:"read_timeout"`
		WriteTimeout time.Duration `yaml:"write_timeout"`
		IdleTimeout  time.Duration `yaml:"idle_timeout"`
	} `yaml:"server"`
	Model struct {
		Default     string  `yaml:"default"`
		Fallback    string  `yaml:"fallback"`
		NumCtx      int     `yaml:"num_ctx"`
		Temperature float64 `yaml:"temperature"`
		TopP        float64 `yaml:"top_p"`
	} `yaml:"model"`
	Cache struct {
		SQLitePath  string   `yaml:"sqlite_path"`
		LRUMaxItems int      `yaml:"lru_max_items"`
		TTLMinutes  int      `yaml:"ttl_minutes"`
		Pragmas     []string `yaml:"sqlite_pragmas"`
	} `yaml:"cache"`
	System struct {
		GOGC          int    `yaml:"gogc"`
		OllamaHost    string `yaml:"ollama_host"`
		MemoryLimitMB int    `yaml:"memory_hard_limit_mb"`
		LogLevel      string `yaml:"log_level"`
	} `yaml:"system"`
}

func Load(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Model.NumCtx < 512 || cfg.Model.NumCtx > 8192 {
		return nil, fmt.Errorf("num_ctx out of range [512-8192]")
	}
	if cfg.Model.Temperature < 0 || cfg.Model.Temperature > 2.0 {
		return nil, fmt.Errorf("temperature out of range [0.0-2.0]")
	}
	if cfg.System.OllamaHost == "" {
		cfg.System.OllamaHost = os.Getenv("OLLAMA_HOST")
		if cfg.System.OllamaHost == "" {
			cfg.System.OllamaHost = "127.0.0.1:11434"
		}
	}
	return &cfg, nil
}
