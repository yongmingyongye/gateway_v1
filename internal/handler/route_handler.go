package handler

import (
	"gateway/internal/util"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"gateway/internal/model"
	"gateway/internal/service"
	"gateway/pkg/logger"
	"github.com/gin-gonic/gin"
)

func AddRoute(c *gin.Context) {
	var newProxy model.Proxy
	if err := c.ShouldBindJSON(&newProxy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Generate a unique ID for the new route
	newProxy.ID = util.GenerateUUID()

	// Ensure the prefix starts with a slash
	if !strings.HasPrefix(newProxy.Prefix, "/") {
		newProxy.Prefix = "/" + newProxy.Prefix
	}

	if err := service.AddProxy(newProxy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newProxy)
}

func UpdateRoute(c *gin.Context) {
	id := c.Param("id")
	var updatedProxy model.Proxy
	if err := c.ShouldBindJSON(&updatedProxy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	updatedProxy.ID = id

	// Ensure the prefix starts with a slash
	if !strings.HasPrefix(updatedProxy.Prefix, "/") {
		updatedProxy.Prefix = "/" + updatedProxy.Prefix
	}

	if err := service.UpdateProxy(updatedProxy); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedProxy)
}

func DeleteRoute(c *gin.Context) {
	id := c.Param("id")

	if err := service.DeleteProxy(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Route deleted"})
}

func ListRoutes(c *gin.Context) {
	proxies := service.ListProxies()
	c.JSON(http.StatusOK, proxies)
}

func GetRoute(c *gin.Context) {
	id := c.Param("id")

	proxy, err := service.GetProxyByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, proxy)
}

func Forward(c *gin.Context) {
	HostReverseProxy(c.Writer, c.Request)
}

func HostReverseProxy(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "/favicon.ico" {
		io.WriteString(w, "Request path Error")
		return
	}

	// 从服务层获取转发的url
	proxy, exists := service.GetProxyByPrefix(r.URL.Path)
	if !exists {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	upstream := proxy.Upstream
	rewritePrefix := proxy.RewritePrefix

	// 如果转发的地址是 / 结尾，需要过滤掉
	if strings.HasSuffix(upstream, "/") {
		upstream = strings.TrimRight(upstream, "/")
	}

	// 如果 rewritePrefix 不为空，替换原路径的前缀
	if rewritePrefix != "" {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, proxy.Prefix)
		if !strings.HasPrefix(rewritePrefix, "/") {
			rewritePrefix = "/" + rewritePrefix
		}
		r.URL.Path = rewritePrefix + r.URL.Path
	} else {
		// 如果 rewritePrefix 为空，则保持原来的路径
		r.URL.Path = strings.TrimPrefix(r.URL.Path, proxy.Prefix)
	}

	remote, err := url.Parse(upstream)
	logger.InfoLog.Printf("RequestURI %s upstream %s remote %s", r.RequestURI, upstream, remote.String())
	if err != nil {
		http.Error(w, "Invalid upstream URL", http.StatusInternalServerError)
		return
	}

	r.URL.Host = remote.Host
	r.URL.Scheme = remote.Scheme
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = remote.Host

	httputil.NewSingleHostReverseProxy(remote).ServeHTTP(w, r)
}
