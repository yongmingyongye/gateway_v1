package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"gateway/internal/config"
	"gateway/internal/model"
	"gateway/pkg/logger"
	"gorm.io/gorm"
)

var (
	proxyMap   = make(map[string]model.Proxy)
	proxyMutex sync.RWMutex
	ctx        = context.Background()
)

// AddProxy 添加一个新的代理配置
func AddProxy(proxy model.Proxy) error {
	proxyMutex.Lock()
	defer proxyMutex.Unlock()

	// 更新 MySQL
	result := config.DB.Create(&proxy)
	if result.Error != nil {
		logger.ErrorLog.Printf("Failed to add proxy to MySQL: %v", result.Error)
		return result.Error
	}

	// 更新 Redis Hash
	err := setProxyToRedisHash(&proxy)
	if err != nil {
		logger.ErrorLog.Printf("Failed to set proxy to Redis Hash: %v", err)
	}

	// 更新内存中的代理映射
	proxyMap[proxy.Prefix] = proxy
	logger.InfoLog.Printf("Added proxy: %v", proxy)
	return nil
}

// UpdateProxy 更新现有的代理配置
func UpdateProxy(proxy model.Proxy) error {
	proxyMutex.Lock()
	defer proxyMutex.Unlock()

	// 更新 MySQL
	result := config.DB.Model(&model.Proxy{}).Where("prefix = ?", proxy.Prefix).Updates(proxy)
	if result.Error != nil {
		logger.ErrorLog.Printf("Failed to update proxy in MySQL: %v", result.Error)
		return result.Error
	}

	// 更新 Redis Hash
	err := setProxyToRedisHash(&proxy)
	if err != nil {
		logger.ErrorLog.Printf("Failed to set proxy to Redis Hash: %v", err)
	}

	// 更新内存中的代理映射
	proxyMap[proxy.Prefix] = proxy
	logger.InfoLog.Printf("Updated proxy: %v", proxy)
	return nil
}

// DeleteProxy 根据 ID 删除代理配置
func DeleteProxy(id string) error {
	proxyMutex.Lock()
	defer proxyMutex.Unlock()

	// 查找并删除 MySQL 中的代理
	var proxy model.Proxy
	result := config.DB.Where("id = ?", id).First(&proxy)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return fmt.Errorf("route not found")
		}
		logger.ErrorLog.Printf("Failed to find proxy by ID %s: %v", id, result.Error)
		return result.Error
	}

	// 删除 MySQL 中的代理
	result = config.DB.Delete(&proxy)
	if result.Error != nil {
		logger.ErrorLog.Printf("Failed to delete proxy from MySQL: %v", result.Error)
		return result.Error
	}

	// 从 Redis Hash 中删除
	err := deleteProxyFromRedisHash(proxy.Prefix)
	if err != nil {
		logger.ErrorLog.Printf("Failed to delete proxy from Redis Hash: %v", err)
	}

	// 从内存中删除代理映射
	delete(proxyMap, proxy.Prefix)
	logger.InfoLog.Printf("Deleted proxy with ID: %s", id)
	return nil
}

// ListProxies 返回所有代理配置
func ListProxies() []model.Proxy {
	proxyMutex.RLock()
	defer proxyMutex.RUnlock()

	var proxies []model.Proxy
	for _, proxy := range proxyMap {
		proxies = append(proxies, proxy)
	}
	return proxies
}

// GetProxyByID 根据 ID 获取代理配置
func GetProxyByID(id string) (*model.Proxy, error) {
	proxyMutex.RLock()
	defer proxyMutex.RUnlock()

	for _, proxy := range proxyMap {
		if proxy.ID == id {
			return &proxy, nil
		}
	}
	return nil, fmt.Errorf("route not found")
}

// GetProxyByPrefix 根据前缀获取代理配置
func GetProxyByPrefix(prefix string) (*model.Proxy, bool) {
	proxyMutex.RLock()
	defer proxyMutex.RUnlock()

	// 优先从内存中查找
	if proxy, exists := proxyMap[prefix]; exists {
		return &proxy, true
	}

	// 如果内存中不存在，尝试从 Redis Hash 中获取
	proxy, err := getProxyFromRedisHash(prefix)
	if err == nil && proxy != nil {
		return proxy, true
	}

	// 如果 Redis Hash 中也不存在，尝试从 MySQL 中获取
	proxy, err = getProxyFromMySQL(prefix)
	if err != nil {
		logger.ErrorLog.Printf("Failed to get proxy from MySQL: %v", err)
		return nil, false
	}

	// 将结果缓存到 Redis Hash 中
	if proxy != nil {
		err = setProxyToRedisHash(proxy)
		if err != nil {
			logger.ErrorLog.Printf("Failed to set proxy to Redis Hash: %v", err)
		}
	}

	// 更新内存中的代理映射
	proxyMutex.Lock()
	proxyMap[prefix] = *proxy
	proxyMutex.Unlock()

	return proxy, true
}

// LoadProxiesFromMySQL 从 MySQL 中加载所有路由信息
func LoadProxiesFromMySQL() error {
	var proxies []model.Proxy
	result := config.DB.Find(&proxies)
	if result.Error != nil {
		logger.ErrorLog.Printf("Failed to load proxies from MySQL: %v", result.Error)
		return result.Error
	}

	proxyMutex.Lock()
	defer proxyMutex.Unlock()

	for _, proxy := range proxies {
		proxyMap[proxy.Prefix] = proxy
		logger.InfoLog.Printf("Loaded proxy from MySQL: %v", proxy)
	}

	// 将所有路由信息缓存到 Redis Hash 中
	for _, proxy := range proxies {
		err := setProxyToRedisHash(&proxy)
		if err != nil {
			logger.ErrorLog.Printf("Failed to set proxy to Redis Hash: %v", err)
		}
	}

	return nil
}

// setProxyToRedisHash 将代理配置缓存到 Redis Hash 中
func setProxyToRedisHash(proxy *model.Proxy) error {
	return config.RedisTool.HSet("routes", proxy.Prefix, proxy)
}

// getProxyFromRedisHash 从 Redis Hash 中获取代理配置
func getProxyFromRedisHash(prefix string) (*model.Proxy, error) {
	var proxy model.Proxy
	err := config.RedisTool.HGet("routes", prefix, &proxy)
	if err != nil {
		return nil, err
	}
	return &proxy, nil
}

// deleteProxyFromRedisHash 从 Redis Hash 中删除代理配置
func deleteProxyFromRedisHash(prefix string) error {
	return config.RedisTool.HDel("routes", prefix)
}

// getAllProxiesFromRedisHash 从 Redis Hash 中获取所有代理配置
func getAllProxiesFromRedisHash() (map[string]model.Proxy, error) {
	// 获取所有字段和值
	values, err := config.RedisTool.HGetAll("routes")
	if err != nil {
		return nil, err
	}

	proxies := make(map[string]model.Proxy)
	for prefix, val := range values {
		var proxy model.Proxy
		if err := json.Unmarshal([]byte(val), &proxy); err != nil {
			logger.ErrorLog.Printf("Failed to unmarshal proxy from Redis Hash: %v", err)
			continue
		}
		proxies[prefix] = proxy
	}

	return proxies, nil
}

// getProxyFromMySQL 从 MySQL 中获取代理配置
func getProxyFromMySQL(prefix string) (*model.Proxy, error) {
	var proxy model.Proxy
	result := config.DB.Where("prefix = ?", prefix).First(&proxy)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // 没有找到记录，返回 nil
		}
		return nil, result.Error // 其他错误
	}

	return &proxy, nil
}
