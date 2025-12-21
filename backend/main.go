// @title Library Management System API
// @version 1.0
// @description This is a library management system API
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email dailiduzhou@example.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost
// @BasePath /api
// @schemes   https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Dailiduzhou/library_manage_sys/config"
	"github.com/Dailiduzhou/library_manage_sys/middleware"
	"github.com/Dailiduzhou/library_manage_sys/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	_ "github.com/Dailiduzhou/library_manage_sys/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	config.ConnectDB()
	config.InitAdmin(config.DB)

	r := gin.Default()

	corsConfig := cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://127.0.0.1:3000",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	r.Use(cors.New(corsConfig))
	r.Static("/uploads", "./uploads")

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	if err := middleware.InitSession(r); err != nil {
		log.Printf("会话创建失败: %q", err)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	routes.RegisterUserRoutes(r)
	routes.RegisterBookRouters(r)

	log.Println("服务器启动")
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")

	// 设置 5 秒超时，处理未完成的请求
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("服务器强制关闭:", err)
	}

	// 关闭数据库连接
	sqlDB, err := config.DB.DB()
	if err == nil {
		sqlDB.Close()
	}

	log.Println("服务器已优雅退出")
}
