package config

import (
	"database/sql"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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

type RedisConfig struct {
	Addr     string `yaml:"Addr"`
	Password string `yaml:"Password"`
	DB       int    `yaml:"DB"`
}

type Config struct {
	Database    DBConfig    `yaml:"Database"`
	RedisConfig RedisConfig `yaml:"RedisConfig"`
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

func SetupDatabase(cfg DBConfig) (*sqlx.SqlConn, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.Options,
	)

	db, err := sql.Open(cfg.Type, dsn)
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetMaxIdleConns(cfg.MaxIdleConnections)
	db.SetConnMaxLifetime(time.Duration(cfg.MaxLifeTime) * time.Second)

	conn := sqlx.NewSqlConnFromDB(db)

	return &conn, nil
}

func SetupRedis(cfg RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,     // Redis 服务器地址
		Password: cfg.Password, // Redis 密码，如果没有密码可留空
		DB:       cfg.DB,       // 使用的 Redis 数据库
	})
	return rdb, nil
}
