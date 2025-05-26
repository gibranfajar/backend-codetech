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
	"github.com/gosimple/slug"
)

// get all data
func GetAllServices(c *gin.Context) {
	var services []model.Service

	rows, err := config.DB.Query("SELECT id, title, slug, description, icon, created_at, updated_at FROM services")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
		return
	}

	if rows == nil {
		c.JSON(http.StatusOK, gin.H{"message": "No data found"})
	}

	for rows.Next() {
		var service model.Service
		if err := rows.Scan(&service.Id, &service.Title, &service.Slug, &service.Description, &service.Icon, &service.CreatedAt, &service.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
			return
		}
		services = append(services, service)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": services,
	})
}

// create data
func CreateService(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")
	file, err := c.FormFile("icon")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Icon is required"})
		return
	}

	// validasi
	var req model.ServiceRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi menggunakan validator
	err = config.Validate.Struct(req)
	if err != nil {
		errors := []string{}
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	os.MkdirAll("uploads", os.ModePerm)
	filename := uuid.New().String() + filepath.Ext(file.Filename)
	savePath := "uploads/" + filename
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	icon := "/uploads/" + filename

	_, err = config.DB.Exec(`
		INSERT INTO services (title, slug, description, icon, created_at, updated_at)
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6)
	`, title, slug.Make(title), description, icon, time.Now(), time.Now())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Data created successfully",
	})
}

// update data
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

	// Validasi menggunakan validator
	err = config.Validate.Struct(req)
	if err != nil {
		errors := []string{}
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	title := c.PostForm("title")
	description := c.PostForm("description")

	// Ambil data lama untuk dapatkan icon lama
	var oldIcon string
	err = config.DB.QueryRow("SELECT icon FROM services WHERE id = @p1", sql.Named("p1", id)).Scan(&oldIcon)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing service", "detail": err.Error()})
		return
	}

	iconPath := oldIcon // default: gunakan icon lama

	file, err := c.FormFile("icon")
	if err == nil {
		// Jika ada file baru, upload dan ganti
		os.MkdirAll("uploads", os.ModePerm)
		filename := uuid.New().String() + filepath.Ext(file.Filename)
		savePath := "uploads/" + filename
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
			return
		}
		iconPath = "/uploads/" + filename
	}

	// hapus file lama jika ada
	if oldIcon != "" {
		_, iconFile := filepath.Split(oldIcon)
		iconPath := filepath.Join("uploads", iconFile)
		if _, err := os.Stat(iconPath); err == nil {
			if err := os.Remove(iconPath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image", "detail": err.Error()})
				return
			}
		}
	}

	_, err = config.DB.Exec(`
		UPDATE services
		SET title = @p1, slug = @p2, description = @p3, icon = @p4, updated_at = @p5
		WHERE id = @p6
	`, title, slug.Make(title), description, iconPath, time.Now(), id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data updated successfully",
	})
}

// delete data
func DeleteService(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Ambil data lama untuk dapatkan icon lama
	var oldIcon string
	err = config.DB.QueryRow("SELECT icon FROM services WHERE id = @p1", sql.Named("p1", id)).Scan(&oldIcon)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing service", "detail": err.Error()})
		return
	}

	_, err = config.DB.Exec(`
		DELETE FROM services
		WHERE id = @p1
	`, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data", "detail": err.Error()})
		return
	}

	// Hapus file gambar jika ada
	if oldIcon != "" {
		_, imageFile := filepath.Split(oldIcon)
		imagePath := filepath.Join("uploads", imageFile)
		if _, err := os.Stat(imagePath); err == nil {
			if err := os.Remove(imagePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image", "detail": err.Error()})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data deleted successfully",
	})
}
