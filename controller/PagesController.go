package controller

import (
	"database/sql"
	"fmt"
	"log"
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
	"github.com/gosimple/slug"
)

func GetAllPages(c *gin.Context) {
	var pages []model.Pages

	rows, err := config.DB.Query("SELECT id, title, slug, type, description, banner, created_at, updated_at FROM pages")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
		return
	}

	if rows == nil {
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
		return
	}

	for rows.Next() {
		var page model.Pages
		if err := rows.Scan(&page.Id, &page.Title, &page.Slug, &page.Type, &page.Description, &page.Banner, &page.CreatedAt, &page.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
			return
		}
		pages = append(pages, page)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": pages,
	})
}

func CreatePage(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")
	types := c.PostForm("type")

	var req model.PageRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi menggunakan validator
	err := config.Validate.Struct(req)
	if err != nil {
		errors := []string{}
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	// Upload banner wajib
	file, err := c.FormFile("banner")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Banner image is required"})
		return
	}

	os.MkdirAll("uploads", os.ModePerm)
	filename := uuid.New().String() + filepath.Ext(file.Filename)
	savePath := "uploads/" + filename
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	// Simpan ke database PostgreSQL
	query := `
		INSERT INTO pages (title, slug, type, description, banner, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = config.DB.Exec(
		query,
		title,
		slug.Make(title),
		types,
		description,
		"/uploads/"+filename,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to insert data",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Page created successfully",
	})
}

// update
func UpdatePage(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req model.PageRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi menggunakan validator
	err = config.Validate.Struct(req)
	if err != nil {
		var errors []string
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	title := c.PostForm("title")
	description := c.PostForm("description")
	types := c.PostForm("type")

	// Ambil banner lama dari database
	var oldBanner string
	err = config.DB.QueryRow("SELECT banner FROM pages WHERE id = $1", id).Scan(&oldBanner)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to fetch existing page",
			"detail": err.Error(),
		})
		return
	}

	bannerPath := oldBanner // default: gunakan banner lama

	// Cek apakah ada file baru
	file, err := c.FormFile("banner")
	if err == nil {
		os.MkdirAll("uploads", os.ModePerm)
		filename := uuid.New().String() + filepath.Ext(file.Filename)
		savePath := filepath.Join("uploads", filename)

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload new banner"})
			return
		}

		// Hapus file lama jika ada
		if oldBanner != "" {
			_, oldFile := filepath.Split(oldBanner)
			oldFilePath := filepath.Join("uploads", oldFile)
			if _, err := os.Stat(oldFilePath); err == nil {
				os.Remove(oldFilePath)
			}
		}

		bannerPath = "/uploads/" + filename
	}

	// Update data di PostgreSQL
	query := `
		UPDATE pages
		SET title = $1,
			slug = $2,
			type = $3,
			description = $4,
			banner = $5,
			updated_at = $6
		WHERE id = $7
	`

	_, err = config.DB.Exec(
		query,
		title,
		slug.Make(title),
		types,
		description,
		bannerPath,
		time.Now(),
		id,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to update page",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Page updated successfully"})
}

// delete
func DeletePage(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var page model.Pages
	query := "SELECT id, banner FROM pages WHERE id = $1"
	err = config.DB.QueryRow(query, id).Scan(&page.Id, &page.Banner)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Database error",
			"detail": err.Error(),
		})
		return
	}

	deleteQuery := "DELETE FROM pages WHERE id = $1"
	result, err := config.DB.Exec(deleteQuery, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to delete page",
			"detail": err.Error(),
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No page was deleted"})
		return
	}

	// Hapus file banner dari folder uploads jika ada
	if page.Banner != "" {
		_, fileName := filepath.Split(page.Banner)
		filePath := filepath.Join("uploads", fileName)

		if _, err := os.Stat(filePath); err == nil {
			if err := os.Remove(filePath); err != nil {
				log.Printf("Failed to delete banner file: %s", err)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Page deleted successfully",
	})
}
