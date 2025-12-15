package models

import "mime/multipart"

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CreateBookRequest struct {
	Title        string                `form:"title" binding:"required"`
	Author       string                `form:"author" binding:"required"`
	Summary      string                `form:"summary" binding:"omitempty"`
	Cover        *multipart.FileHeader `form:"cover" binding:"omitempty"`
	InitialStock int                   `form:"initial_stock" binding:"gte=0"`
}

// 需要图书ID字段
type UpdateBookRequest struct {
	ID         uint                  `form:"id" binding:"required"`
	Title      string                `form:"title" binding:"omitempty"`
	Author     string                `form:"author" binding:"omitempty"`
	Summary    string                `form:"summary" binding:"omitempty"`
	Cover      *multipart.FileHeader `form:"cover" binding:"omitempty"`
	Stock      int                   `form:"stock" binding:"omitempty"`
	TotalStock int                   `form:"total_stock" binding:"omitempty"`
}

type FindBookRequest struct {
	ID uint `json:"id" binding:"required"`
}
