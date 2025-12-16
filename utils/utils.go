package utils

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	var err error
	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPasswordBytes), nil
}

func ComparePassword(dbPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(password))
	if err != nil {
		return err
	}
	return nil
}

func SaveImages(c *gin.Context, file *multipart.FileHeader) (string, error) {
	uplaodDir := "uploads"

	if _, err := os.Stat(uplaodDir); os.IsNotExist(err) {
		os.Mkdir(uplaodDir, 0755)
	}

	ext := filepath.Ext(file.Filename)
	timestamp := time.Now().Unix()
	randomStr := uuid.New().String()[:8]
	newFileName := fmt.Sprintf("cover_%d_%s%s", timestamp, randomStr, ext)

	dst := filepath.Join(uplaodDir, newFileName)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		return "", err
	}

	return filepath.ToSlash(dst), nil
}

func RemoveFile(filePath string) error {
	if filePath != "" {
		err := os.Remove(filePath)
		return err
	}
	return nil
}
