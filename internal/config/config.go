package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string        `yaml:"env" env-default:"local"`
	StoragePath string        `yaml:"storage_path"`
	TokenTTL    time.Duration `yaml:"token_ttl" env-required:"true"`
	GRPC        GRPCConfig    `yaml:"grpc"`
	DB          DBConfig      `yaml:"db"`
}

type DBConfig struct {
	Username string        `mapstructure:"username"`
	Password string        `mapstructure:"password"`
	Host     string        `mapstructure:"host"`
	Port     string        `mapstructure:"port"`
	DBname   string        `mapstructure:"dbname"`
	SSLmode  string        `mapstructure:"sslmode"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

func MustLoad() *Config {
	path := fetchConfig()
	if path == "" {
		panic("config path is empty")
	}

	return MustByLoad(path)
}

func MustByLoad(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file doesnt exist:" + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("failed to read config:" + err.Error())
	}

	return &cfg
}

// парсинг path-a конфига из командной строки в виде: --config="path/path/..."
func fetchConfig() string {
	var res string
	flag.StringVar(&res, "config", "./config/localv2.yaml", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
