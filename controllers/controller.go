package controller

import (
	"errors"
	"net/http"
	"time"

	"github.com/Dailiduzhou/library_manage_sys/config"
	"github.com/Dailiduzhou/library_manage_sys/models"
	"github.com/Dailiduzhou/library_manage_sys/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrBookNotFound   = errors.New("图书不存在")
	ErrNoStock        = errors.New("图书库存不足")
	ErrRecordNotFound = errors.New("借书记录查询失败")
	ErrBookBorrowed   = errors.New("图书仍在借")
	ErrDeleteBook     = errors.New("图书删除失败")
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func Register(c *gin.Context) {
	var req models.RegisterRequest
	var err error

	if err = c.ShouldBindJSON(&req); err != nil {
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

	if err = c.ShouldBindJSON(&req); err != nil {
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

func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	if err := session.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "登出失败",
		})
	}

	// 希望前端实现跳转登录界面的功能
	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "登出成功",
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
		Title:        req.Title,
		Author:       req.Author,
		Summary:      finalSummary,
		CoverPath:    finalCoverPath,
		InitialStock: req.InitialStock,
		Stock:        req.InitialStock,
		TotalStock:   req.InitialStock,
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

func GetBooks(c *gin.Context) {
	var books []models.Book

	title := c.Query("title")
	author := c.Query("author")
	summary := c.Query("summary")

	query := config.DB.Model(&models.Book{})

	if title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}
	if author != "" {
		query = query.Where("author LIKE ?", "%"+author+"%")
	}
	if summary != "" {
		query = query.Where("summary LIKE ?", "%"+summary+"%")
	}

	if err := query.Order("id desc").Find(&books).Error; err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "数据库查询失败",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "查询成功",
		Data: books,
	})
}

func UpdateBook(c *gin.Context) {
	var req models.UpdateBookRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusNotFound, Response{
			Code: 404,
			Msg:  "参数设定错误",
		})
		return
	}

	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var book models.Book
	result := tx.Set("gorm:query_option", "FOR UPDATE").First(&book, req.ID)

	if result.Error != nil {
		tx.Rollback()
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, Response{
				Code: 404,
				Msg:  "图书不存在",
			})
		} else {
			c.JSON(http.StatusInternalServerError, Response{
				Code: 500,
				Msg:  "数据库查询失败: ",
			})
		}
		return
	}

	updates := make(map[string]interface{})

	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Author != "" {
		updates["author"] = req.Author
	}
	if req.Summary != "" {
		updates["summary"] = req.Summary
	}

	if req.Stock >= 0 || req.TotalStock > 0 {
		newStock := book.Stock
		newTotalStock := book.TotalStock

		if req.Stock >= 0 {
			newStock = req.Stock
		}
		if req.TotalStock > 0 {
			newTotalStock = req.TotalStock
		}

		if newStock > newTotalStock {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, Response{
				Code: 400,
				Msg:  "当前库存不能大于总库存",
			})
			return
		}

		updates["stock"] = newStock
		updates["total_stock"] = newTotalStock
	}

	if req.Cover != nil {
		coverPath, err := utils.SaveImages(c, req.Cover)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, Response{
				Code: 500,
				Msg:  "封面图片保存失败",
			})
			return
		}
		if book.CoverPath != "" {
			utils.RemoveFile(book.CoverPath)
		}

		updates["cover_path"] = coverPath
	}

	if len(updates) > 0 {
		if err := tx.Model(&book).Updates(updates).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, Response{
				Code: 500,
				Msg:  "更新失败: " + err.Error(),
			})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "提交事务失败: ",
		})
		return
	}
	config.DB.First(&book, req.ID)

	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "图书更新成功",
		Data: book,
	})
}

func DeleteBooks(c *gin.Context) {
	var req models.FindBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "参数设定错误",
		})
		return
	}

	err := config.DB.Transaction(func(tx *gorm.DB) error {
		var existingBook models.Book

		if err := tx.First(&existingBook, req.ID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrBookNotFound
			}
			return err
		}

		if existingBook.Stock != existingBook.TotalStock {
			return ErrBookBorrowed
		}

		if err := tx.Delete(&existingBook).Error; err != nil {
			return ErrDeleteBook
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, ErrBookNotFound) {
			c.JSON(http.StatusNotFound, Response{
				Code: 404,
				Msg:  "图书不存在",
			})
			return
		}
		if errors.Is(err, ErrDeleteBook) {
			c.JSON(http.StatusInternalServerError, Response{
				Code: 500,
				Msg:  "删除图书失败",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "系统繁忙,请稍后再试",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "删除图书成功",
	})
}

func BorrowBook(c *gin.Context) {
	var req models.FindBookRequest
	userID := c.GetUint("user_id")
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "参数设定失败",
		})
	}

	var borrowRecord models.BorrowRecord
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		var book models.Book

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&book).Error; err != nil {
			if errors.Is(err, ErrBookNotFound) {
				return ErrBookNotFound
			}
			return err
		}

		if book.Stock <= 0 {
			return ErrNoStock
		}

		if err := tx.Model(&book).Update("stock", book.Stock-1).Error; err != nil {
			return err
		}

		borrowRecord = models.BorrowRecord{
			UserID:     userID,
			BookID:     req.ID,
			BorrowDate: time.Now(),
			Status:     "borrowed",
		}

		if err := tx.Create(&borrowRecord).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, ErrBookNotFound) {
			c.JSON(http.StatusNotFound, Response{
				Code: 404,
				Msg:  "图书不存在",
			})
			return
		}
		if errors.Is(err, ErrNoStock) {
			c.JSON(http.StatusOK, Response{
				Code: 200,
				Msg:  "图书无库存",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "系统繁忙,请稍后再试",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "借书成功",
		Data: borrowRecord,
	})
}

func ReturnBook(c *gin.Context) {
	var req models.FindBookRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code: 400,
			Msg:  "参数设定错误",
		})
		return
	}

	var borrowRecord models.BorrowRecord
	err := config.DB.Transaction(func(tx *gorm.DB) error {
		var book models.Book
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&book, req.ID).Error; err != nil {
			if errors.Is(err, ErrBookNotFound) {
				return ErrBookNotFound
			}
			return err
		}

		if book.Stock <= 0 {
			return ErrNoStock
		}

		if err := tx.Model(&book).Update("stock", book.Stock+1).Error; err != nil {
			return err
		}

		if err := tx.Where("user_id = ? AND book_id = ?").First(&borrowRecord).Error; err != nil {
			return ErrRecordNotFound
		}

		if err := tx.Model(&borrowRecord).Updates(models.BorrowRecord{
			ReturnDate: time.Now(),
			Status:     "returned"}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, ErrBookNotFound) {
			c.JSON(http.StatusNotFound, Response{
				Code: 404,
				Msg:  "图书不存在",
			})
			return
		}

		if errors.Is(err, ErrRecordNotFound) {
			c.JSON(http.StatusFound, Response{
				Code: 302,
				Msg:  "找不到图书记录",
			})
		}

		c.JSON(http.StatusInternalServerError, Response{
			Code: 500,
			Msg:  "系统错误",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "还书成功",
		Data: borrowRecord,
	})
}
