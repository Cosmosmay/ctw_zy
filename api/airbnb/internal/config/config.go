package config

import (
	"database/sql"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

type DBConfig struct {
	Type               string `yaml:"Type"`
	Host               string `yaml:"Host"`
	Port               int    `yaml:"Port"`
	User               string `yaml:"User"`
	Password           string `yaml:"Password"`
	Name               string `yaml:"Name"`
	SslMode            string `yaml:"SslMode"`
	Options            string `yaml:"Options"`
	MaxConnections     int    `yaml:"MaxConnections"`
	MaxIdleConnections int    `yaml:"MaxIdleConnections"`
	MaxLifeTime        int    `yaml:"MaxLifeTime"`
}

type Config struct {
	Database DBConfig `yaml:"Database"`
}

func LoadConfig(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	var cfg Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func SetupDatabase(cfg DBConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.Options,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(time.Duration(cfg.MaxLifeTime) * time.Second)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
