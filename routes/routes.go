package routes

import (
	controller "github.com/Dailiduzhou/library_manage_sys/controllers"
	"github.com/Dailiduzhou/library_manage_sys/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.POST("/register", controller.Register)
		api.POST("/login", controller.Login)

		authGroup := api.Group("/")
		authGroup.Use(middleware.AuthRequired())
		{
			authGroup.POST("/logout", controller.Logout)
			authGroup.POST("/borrow", controller.BorrowBook)
			authGroup.POST("/return", controller.ReturnBook)

			// 通过标题、作者和简介模糊搜索
			authGroup.GET("/getbooks", controller.GetBooks)

			adminGroup := authGroup.Group("/")
			adminGroup.Use(middleware.AdminRequired())
			{
				adminGroup.POST("/createbook", controller.CreateBook)
				adminGroup.POST("/updatebook", controller.UpdateBook)
				adminGroup.DELETE("/deletebook", controller.DeleteBooks)
			}
		}
	}
}
