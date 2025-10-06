package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gibranfajar/backend-codetech/config"
	"github.com/gibranfajar/backend-codetech/model"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// get all data
func GetAllCategoryFaq(c *gin.Context) {
	var categoryFaqs []model.CategoryFaq

	rows, err := config.DB.Query("SELECT id, category, description, icon, created_at, updated_at FROM category_faqs")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
		return
	}

	if rows == nil {
		c.JSON(http.StatusOK, gin.H{"message": "No data found"})
	}

	for rows.Next() {
		var categoryFaq model.CategoryFaq
		if err := rows.Scan(&categoryFaq.Id, &categoryFaq.Category, &categoryFaq.Description, &categoryFaq.Icon, &categoryFaq.CreatedAt, &categoryFaq.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
			return
		}
		categoryFaqs = append(categoryFaqs, categoryFaq)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": categoryFaqs,
	})
}

// create data
func CreateCategoryFaq(c *gin.Context) {
	category := c.PostForm("category")
	description := c.PostForm("description")

	var req model.CategoryFaqRequest
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

	// Upload icon (wajib)
	file, err := c.FormFile("icon")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Icon is required"})
		return
	}

	// Simpan file ke folder uploads
	os.MkdirAll("uploads", os.ModePerm)
	filename := uuid.New().String() + filepath.Ext(file.Filename)
	savePath := filepath.Join("uploads", filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	icon := "/uploads/" + filename

	// Simpan ke database (PostgreSQL style)
	_, err = config.DB.Exec(`
		INSERT INTO category_faqs (category, description, icon, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, category, description, icon, time.Now(), time.Now())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Data created successfully",
	})
}

// update data
func UpdateCategoryFaq(c *gin.Context) {
	id := c.Param("id")
	category := c.PostForm("category")
	description := c.PostForm("description")

	var req model.CategoryFaqRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi menggunakan validator
	err := config.Validate.Struct(req)
	if err != nil {
		var errors []string
		for _, e := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s is %s", e.Field(), e.Tag()))
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	// Ambil icon lama
	var oldIcon string
	err = config.DB.QueryRow("SELECT icon FROM category_faqs WHERE id = $1", id).Scan(&oldIcon)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing category", "detail": err.Error()})
		return
	}

	icon := oldIcon // default icon lama

	// Cek apakah user upload file baru
	file, err := c.FormFile("icon")
	if err == nil {
		os.MkdirAll("uploads", os.ModePerm)
		filename := uuid.New().String() + filepath.Ext(file.Filename)
		savePath := filepath.Join("uploads", filename)

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
			return
		}

		// Hapus icon lama jika ada
		if oldIcon != "" {
			oldFile := filepath.Join("uploads", filepath.Base(oldIcon))
			if _, err := os.Stat(oldFile); err == nil {
				_ = os.Remove(oldFile)
			}
		}

		icon = "/uploads/" + filename
	}

	// Update data
	_, err = config.DB.Exec(`
		UPDATE category_faqs
		SET category = $1, description = $2, icon = $3, updated_at = $4
		WHERE id = $5
	`, category, description, icon, time.Now(), id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data updated successfully"})
}

// delete data
func DeleteCategoryFaq(c *gin.Context) {
	id := c.Param("id")

	// Ambil data lama untuk dapatkan icon lama
	var oldIcon string
	err := config.DB.QueryRow("SELECT icon FROM category_faqs WHERE id = $1", id).Scan(&oldIcon)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to fetch existing category",
			"detail": err.Error(),
		})
		return
	}

	// Hapus file lama jika ada
	if oldIcon != "" {
		imageFile := filepath.Base(oldIcon)
		imagePath := filepath.Join("uploads", imageFile)

		if _, err := os.Stat(imagePath); err == nil {
			if err := os.Remove(imagePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":  "Failed to delete image file",
					"detail": err.Error(),
				})
				return
			}
		}
	}

	// Hapus data dari database
	_, err = config.DB.Exec("DELETE FROM category_faqs WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to delete data",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data deleted successfully",
	})
}
