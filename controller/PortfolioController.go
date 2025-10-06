package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gibranfajar/backend-codetech/config"
	"github.com/gibranfajar/backend-codetech/model"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// getAllData
func GetAllPortfolio(c *gin.Context) {
	var portfolios []model.Portfolio

	rows, err := config.DB.Query("SELECT id, title, url, image, created_at, updated_at FROM portfolios")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
		return
	}

	if rows == nil {
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
		return
	}

	for rows.Next() {
		var portfolio model.Portfolio
		if err := rows.Scan(&portfolio.Id, &portfolio.Title, &portfolio.Url, &portfolio.Image, &portfolio.CreatedAt, &portfolio.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
			return
		}
		portfolios = append(portfolios, portfolio)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": portfolios,
	})
}

// create data
func CreatePortfolio(c *gin.Context) {
	title := c.PostForm("title")
	url := c.PostForm("url")

	var req model.PortfolioRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi menggunakan validator
	if err := config.Validate.Struct(req); err != nil {
		var errors []string
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image is required"})
		return
	}

	// Buat folder uploads jika belum ada
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	filename := uuid.New().String() + filepath.Ext(file.Filename)
	savePath := "uploads/" + filename

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	// Simpan ke database PostgreSQL
	query := `
		INSERT INTO portfolios (title, url, image, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = config.DB.Exec(query, title, url, "/uploads/"+filename, time.Now(), time.Now())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Data created successfully",
	})
}

// update data
func UpdatePortfolio(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req model.PortfolioRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi menggunakan validator
	if err := config.Validate.Struct(req); err != nil {
		var errors []string
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	title := c.PostForm("title")
	url := c.PostForm("url")

	// Ambil image lama dari database
	var oldImage string
	err = config.DB.QueryRow("SELECT image FROM portfolios WHERE id = $1", id).Scan(&oldImage)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Portfolio not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing portfolio", "detail": err.Error()})
		return
	}

	imagePath := oldImage // default: gunakan image lama

	// Cek apakah ada file baru yang diupload
	file, err := c.FormFile("image")
	if err == nil {
		// Jika ada file baru, simpan
		if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
			return
		}

		filename := uuid.New().String() + filepath.Ext(file.Filename)
		savePath := "uploads/" + filename
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
			return
		}
		imagePath = "/uploads/" + filename

		// Hapus file lama jika ada
		if oldImage != "" {
			_, imageFile := filepath.Split(oldImage)
			oldFilePath := filepath.Join("uploads", imageFile)
			if _, err := os.Stat(oldFilePath); err == nil {
				_ = os.Remove(oldFilePath) // diabaikan kalau gagal hapus, agar tidak ganggu update
			}
		}
	}

	// Update data ke PostgreSQL
	query := `
		UPDATE portfolios
		SET title = $1, url = $2, image = $3, updated_at = $4
		WHERE id = $5
	`
	_, err = config.DB.Exec(query, title, url, imagePath, time.Now(), id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data updated successfully",
	})
}

// delete data
func DeletePortfolio(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Ambil data lama untuk dapatkan image lama
	var oldImage string
	err = config.DB.QueryRow("SELECT image FROM portfolios WHERE id = $1", id).Scan(&oldImage)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Portfolio not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing portfolio", "detail": err.Error()})
		return
	}

	// Hapus file lama jika ada
	if oldImage != "" {
		_, imageFile := filepath.Split(oldImage)
		imagePath := filepath.Join("uploads", imageFile)
		if _, err := os.Stat(imagePath); err == nil {
			if err := os.Remove(imagePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image file", "detail": err.Error()})
				return
			}
		}
	}

	// Hapus data dari PostgreSQL
	query := `DELETE FROM portfolios WHERE id = $1`
	_, err = config.DB.Exec(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data deleted successfully",
	})
}
