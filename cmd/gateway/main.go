package main

import (
	"flag"
	"fmt"
	"log"

	"gateway/internal/config"
	"gateway/internal/handler"
	"gateway/pkg/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	// 解析命令行参数
	flag.Parse()

	// 初始化配置
	config.InitConfig()

	// 初始化日志
	logger.InitLogger()

	// 创建 Gin 路由
	router := gin.Default()

	// 注册中间件
	router.Use(logger.Logger())

	// 注册路由管理API
	api := router.Group("/api")
	{
		api.POST("/routes", handler.AddRoute)
		api.PUT("/routes/:id", handler.UpdateRoute)
		api.DELETE("/routes/:id", handler.DeleteRoute)
		api.GET("/routes", handler.ListRoutes)
		api.GET("/routes/:id", handler.GetRoute)
	}

	// 处理未匹配到任何具体路径的请求
	router.NoRoute(handler.Forward)

	// 启动服务器
	port := ":8000"
	fmt.Printf("Starting server on port %s\n", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
