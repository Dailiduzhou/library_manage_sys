package routes

import (
	controller "github.com/Dailiduzhou/library_manage_sys/controllers"
	"github.com/Dailiduzhou/library_manage_sys/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{

		auth := api.Group("/auth")
		{
			auth.POST("/register", controller.Register)
			auth.POST("/login", controller.Login)
		}

		authGroup := api.Group("/")
		authGroup.Use(middleware.AuthRequired())
		{
			authGroup.POST("/logout", controller.Logout)
			borrows := api.Group("/borrows")
			{
				// 创建借阅记录 (借书)
				borrows.POST("", controller.BorrowBook)

				borrows.POST("/return", controller.ReturnBook)
			}

			authGroup.GET("/books", controller.GetBooks)

			adminGroup := authGroup.Group("/")
			adminGroup.Use(middleware.AdminRequired())
			{
				// POST /books 创建
				adminGroup.POST("/books", controller.CreateBook)

				// PUT /books/:id 更新
				adminGroup.PUT("/books/:id", controller.UpdateBook)

				// DELETE /books/:id 删除
				adminGroup.DELETE("/books/:id", controller.DeleteBooks)
			}
		}
	}
}
