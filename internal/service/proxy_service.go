package service

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"gateway/internal/model"
)

var (
	ProxyMap   = make(map[string]model.Proxy)
	proxyMutex sync.RWMutex
)

func AddProxy(proxy model.Proxy) error {
	proxyMutex.Lock()
	defer proxyMutex.Unlock()
	ProxyMap[proxy.Prefix] = proxy
	return nil
}

func UpdateProxy(proxy model.Proxy) error {
	proxyMutex.Lock()
	defer proxyMutex.Unlock()
	if _, exists := ProxyMap[proxy.Prefix]; !exists {
		return fmt.Errorf("route not found")
	}
	ProxyMap[proxy.Prefix] = proxy
	return nil
}

func DeleteProxy(id string) error {
	proxyMutex.Lock()
	defer proxyMutex.Unlock()
	for prefix, p := range ProxyMap {
		if p.ID == id {
			delete(ProxyMap, prefix)
			return nil
		}
	}
	return fmt.Errorf("route not found")
}

func ListProxies() []model.Proxy {
	proxyMutex.RLock()
	defer proxyMutex.RUnlock()

	var proxies []model.Proxy
	for _, proxy := range ProxyMap {
		proxies = append(proxies, proxy)
	}
	return proxies
}

func GetProxyByID(id string) (*model.Proxy, error) {
	proxyMutex.RLock()
	defer proxyMutex.RUnlock()

	for _, proxy := range ProxyMap {
		if proxy.ID == id {
			return &proxy, nil
		}
	}
	return nil, fmt.Errorf("route not found")
}

func LoadProxiesFromFile(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	var respond model.Respond
	if err := json.NewDecoder(f).Decode(&respond); err != nil {
		return err
	}

	proxyMutex.Lock()
	defer proxyMutex.Unlock()

	for _, proxy := range respond.Data {
		ProxyMap[proxy.Prefix] = proxy
	}
	return nil
}
