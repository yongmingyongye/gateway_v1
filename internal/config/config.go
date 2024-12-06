package config

import (
	"fmt"
	"gateway/internal/util"
	"github.com/caarlos0/env/v6"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
)

type Config struct {
	Port        string `env:"PORT" envDefault:"8000"`
	MySQLDSN    string `env:"MYSQL_DSN" envDefault:"root:root@tcp(localhost:3306)/gateway?charset=utf8mb4&parseTime=True&loc=Local"`
	RedisAddr   string `env:"REDIS_ADDR" envDefault:"127.0.0.1:6379"`
	RedisPasswd string `env:"REDIS_PASSWD" envDefault:"123456"`
}

var Cfg *Config
var DB *gorm.DB
var RedisTool *util.RedisClient

func InitConfig() error {
	Cfg = &Config{}
	if err := env.Parse(Cfg); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	// 初始化 MySQL 连接
	var err error
	DB, err = gorm.Open(mysql.Open(Cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}

	// 初始化 Redis 连接
	redisClient := redis.NewClient(&redis.Options{
		Addr:     Cfg.RedisAddr,
		Password: Cfg.RedisPasswd,
		DB:       0,
	})

	// 创建 Redis 工具类实例
	RedisTool = util.NewRedisClient(redisClient)

	return nil
}
