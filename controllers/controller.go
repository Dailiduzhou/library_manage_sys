package controller

import (
	"errors"
	"net/http"

	"github.com/Dailiduzhou/libaray_manage_sys/config"
	"github.com/Dailiduzhou/libaray_manage_sys/models"
	"github.com/Dailiduzhou/libaray_manage_sys/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Register(c *gin.Context) {
	var req models.RegisterRequest
	var err error

	if err = c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "参数设定错误",
		})
		return
	}

	var existingUser models.User
	err = config.DB.Where("username = ?", req.Username).First(&existingUser).Error
	if err == nil {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "用户已存在",
		})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "查询数据库失败",
		})
		return
	}

	hashedpassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "密码加密错误",
		})
	}

	newUser := models.User{
		Username: req.Username,
		Password: hashedpassword,
		Role:     "user",
	}

	if err = config.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "创建用户失败",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "注册成功",
		Data: gin.H{
			"username": newUser.Username,
			"user_id":  newUser.ID,
		},
	})
}

func Login(c *gin.Context) {
	var req models.LoginRequest
	var err error

	if err = c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "参数设定错误",
		})
		return
	}

	var user models.User
	err = config.DB.Where("username = ?", req.Username).First(&user).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "查询数据库失败",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusForbidden, Response{
			Code: 403,
			Msg:  "用户不存在",
		})
		return
	}

	err = utils.ComparePassword(user.Password, req.Password)
	if err != nil {
		c.JSON(http.StatusForbidden, Response{
			Code: 403,
			Msg:  "密码错误",
		})
		return
	}

	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Set("role", user.Role)
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "鉴权组件错误",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "登陆成功",
		Data: gin.H{
			"use_id": user.ID,
			"role":   user.Role,
		},
	})
}

func CreateBook(c *gin.Context) {
	var req models.CreateBookRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "参数设定错误",
		})
		return
	}

	var existingBook models.Book
	err := config.DB.Where("title = ? AND author = ?", req.Title, req.Author).First(&existingBook).Error
	if err == nil {
		c.JSON(http.StatusConflict, Response{
			Code: 409,
			Msg:  "该图书已存在(书名和作者相同)",
		})
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "数据库查询失败",
		})
		return
	}

	finalCoverPath := config.DefaultCoverPath
	if req.Cover != nil {
		savePath, err := utils.SaveImages(c, req.Cover)
		if err != nil {
			c.JSON(http.StatusInternalServerError, Response{
				Code: 500,
				Msg:  "图片保存失败",
			})
			return
		}
		finalCoverPath = savePath
	}

	finalSummary := config.DefaultSummary
	if req.Summary != "" {
		finalSummary = req.Summary
	}

	newBook := models.Book{
		Title:     req.Title,
		Author:    req.Author,
		Summary:   finalSummary,
		CoverPath: finalCoverPath,
	}

	if err := config.DB.Create(&newBook).Error; err != nil {
		if req.Cover != nil {
			utils.RemoveFile(finalCoverPath)
		}

		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "创建图书失败",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "图书创建成功",
		Data: newBook,
	})
}
