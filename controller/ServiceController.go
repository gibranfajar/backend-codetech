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

// GetAllServices - get all data
func GetAllServices(c *gin.Context) {
	var services []model.Service

	rows, err := config.DB.Query(`SELECT id, title, slug, description, icon, created_at, updated_at FROM services`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var service model.Service
		if err := rows.Scan(&service.Id, &service.Title, &service.Slug, &service.Description, &service.Icon, &service.CreatedAt, &service.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan data", "detail": err.Error()})
			return
		}
		services = append(services, service)
	}

	if len(services) == 0 {
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": services})
}

// CreateService - create new data
func CreateService(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")

	// Validasi input
	var req model.ServiceRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.Validate.Struct(req); err != nil {
		var errors []string
		for _, e := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s is %s", e.Field(), e.Tag()))
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	// Upload icon wajib
	file, err := c.FormFile("icon")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Icon is required"})
		return
	}

	os.MkdirAll("uploads", os.ModePerm)
	filename := uuid.New().String() + filepath.Ext(file.Filename)
	savePath := filepath.Join("uploads", filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload icon"})
		return
	}

	icon := "/uploads/" + filename

	query := `
		INSERT INTO services (title, slug, description, icon, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err = config.DB.Exec(query, title, slug.Make(title), description, icon, time.Now(), time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Data created successfully"})
}

// UpdateService - update data
func UpdateService(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req model.ServiceRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := config.Validate.Struct(req); err != nil {
		var errors []string
		for _, e := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s is %s", e.Field(), e.Tag()))
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	title := c.PostForm("title")
	description := c.PostForm("description")

	var oldIcon string
	err = config.DB.QueryRow("SELECT icon FROM services WHERE id = $1", id).Scan(&oldIcon)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing data", "detail": err.Error()})
		return
	}

	iconPath := oldIcon

	// Handle upload icon baru
	file, err := c.FormFile("icon")
	if err == nil {
		os.MkdirAll("uploads", os.ModePerm)
		filename := uuid.New().String() + filepath.Ext(file.Filename)
		savePath := filepath.Join("uploads", filename)

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload new icon"})
			return
		}

		// Hapus icon lama
		if oldIcon != "" {
			_, oldFile := filepath.Split(oldIcon)
			oldPath := filepath.Join("uploads", oldFile)
			if _, err := os.Stat(oldPath); err == nil {
				os.Remove(oldPath)
			}
		}

		iconPath = "/uploads/" + filename
	}

	// Update ke DB
	query := `
		UPDATE services
		SET title = $1, slug = $2, description = $3, icon = $4, updated_at = $5
		WHERE id = $6
	`
	_, err = config.DB.Exec(query, title, slug.Make(title), description, iconPath, time.Now(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data updated successfully"})
}

// DeleteService - delete data
func DeleteService(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var oldIcon string
	err = config.DB.QueryRow("SELECT icon FROM services WHERE id = $1", id).Scan(&oldIcon)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
		return
	}

	_, err = config.DB.Exec("DELETE FROM services WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data", "detail": err.Error()})
		return
	}

	if oldIcon != "" {
		_, imageFile := filepath.Split(oldIcon)
		imagePath := filepath.Join("uploads", imageFile)
		if _, err := os.Stat(imagePath); err == nil {
			if err := os.Remove(imagePath); err != nil {
				log.Printf("Warning: failed to delete icon file: %v", err)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data deleted successfully"})
}
