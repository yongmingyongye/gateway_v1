package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
)

// RedisClient 是一个通用的 Redis 客户端工具类
type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisClient 创建一个新的 Redis 客户端实例
func NewRedisClient(redisClient *redis.Client) *RedisClient {
	return &RedisClient{
		client: redisClient,
		ctx:    context.Background(),
	}
}

// Set 设置一个键值对
func (r *RedisClient) Set(key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %v", err)
	}
	return r.client.Set(r.ctx, key, data, 0).Err()
}

// Get 获取一个键对应的值
func (r *RedisClient) Get(key string, result interface{}) error {
	val, err := r.client.Get(r.ctx, key).Result()
	if err == redis.Nil {
		return errors.New("key not found")
	}
	if err != nil {
		return fmt.Errorf("failed to get value from Redis: %v", err)
	}
	return json.Unmarshal([]byte(val), result)
}

// Del 删除一个或多个键
func (r *RedisClient) Del(keys ...string) error {
	return r.client.Del(r.ctx, keys...).Err()
}

// HSet 在 Hash 中设置字段和值
func (r *RedisClient) HSet(hashKey, field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %v", err)
	}
	return r.client.HSet(r.ctx, hashKey, field, data).Err()
}

// HGet 从 Hash 中获取字段对应的值
func (r *RedisClient) HGet(hashKey, field string, result interface{}) error {
	val, err := r.client.HGet(r.ctx, hashKey, field).Result()
	if err == redis.Nil {
		return errors.New("field not found")
	}
	if err != nil {
		return fmt.Errorf("failed to get value from Redis Hash: %v", err)
	}
	return json.Unmarshal([]byte(val), result)
}

// HDel 从 Hash 中删除字段
func (r *RedisClient) HDel(hashKey, field string) error {
	return r.client.HDel(r.ctx, hashKey, field).Err()
}

// HGetAll 获取 Hash 中所有的字段和值
func (r *RedisClient) HGetAll(hashKey string) (map[string]string, error) {
	return r.client.HGetAll(r.ctx, hashKey).Result()
}
