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

// get all article
func GetAllArticle(c *gin.Context) {
	var articles []model.ResponseArticle

	rows, err := config.DB.Query(`
		SELECT 
			a.id, a.title, a.slug, a.description, a.thumbnail, a.views, 
			a.created_at, a.updated_at, 
			u.name AS user_name, 
			c.category AS category_name
		FROM articles a
		JOIN users u ON a.user_id = u.id
		JOIN category_articles c ON a.category_id = c.id
		ORDER BY a.created_at DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to fetch data",
			"detail": err.Error(),
		})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var art model.ResponseArticle
		if err := rows.Scan(
			&art.Id,
			&art.Title,
			&art.Slug,
			&art.Description,
			&art.Thumbnail,
			&art.Views,
			&art.CreatedAt,
			&art.UpdatedAt,
			&art.User,
			&art.Category,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "Failed to parse data",
				"detail": err.Error(),
			})
			return
		}
		articles = append(articles, art)
	}

	if len(articles) == 0 {
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": articles,
	})
}

// create data
func CreateArticle(c *gin.Context) {
	title := c.PostForm("title")
	userID := c.PostForm("user_id")
	categoryID := c.PostForm("category_id")
	description := c.PostForm("description")

	var req model.ArticleRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi menggunakan validator
	if err := config.Validate.Struct(req); err != nil {
		var errors []string
		for _, e := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s is %s", e.Field(), e.Tag()))
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	// Upload file thumbnail
	file, err := c.FormFile("thumbnail")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thumbnail is required"})
		return
	}

	// Pastikan folder uploads ada
	if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	filename := uuid.New().String() + filepath.Ext(file.Filename)
	savePath := filepath.Join("uploads", filename)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	thumbnail := "/uploads/" + filename

	// Simpan ke database (PostgreSQL style)
	_, err = config.DB.Exec(`
		INSERT INTO articles (title, slug, user_id, category_id, description, thumbnail, views, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 0, $7, $8)
	`, title, slug.Make(title), userID, categoryID, description, thumbnail, time.Now(), time.Now())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to insert data",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Data created successfully",
	})
}

// update data
func UpdateArticle(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	title := c.PostForm("title")
	userID := c.PostForm("user_id")
	categoryID := c.PostForm("category_id")
	description := c.PostForm("description")

	var req model.ArticleRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validasi data
	if err := config.Validate.Struct(req); err != nil {
		var errors []string
		for _, e := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s is %s", e.Field(), e.Tag()))
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	// Ambil data artikel lama
	var article model.Article
	err = config.DB.QueryRow(`SELECT id, thumbnail FROM articles WHERE id = $1`, id).Scan(&article.Id, &article.Thumbnail)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "detail": err.Error()})
		return
	}

	// Gunakan thumbnail lama secara default
	thumbnail := article.Thumbnail

	// Jika ada file baru di-upload, ganti thumbnail
	file, err := c.FormFile("thumbnail")
	if err == nil {
		if err := os.MkdirAll("uploads", os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
			return
		}

		filename := uuid.New().String() + filepath.Ext(file.Filename)
		savePath := filepath.Join("uploads", filename)
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
			return
		}

		// Hapus thumbnail lama (jika ada)
		if article.Thumbnail != "" {
			oldFilePath := filepath.Join("uploads", filepath.Base(article.Thumbnail))
			if _, err := os.Stat(oldFilePath); err == nil {
				_ = os.Remove(oldFilePath)
			}
		}

		thumbnail = "/uploads/" + filename
	}

	// Update data ke database
	_, err = config.DB.Exec(`
		UPDATE articles
		SET title = $1, slug = $2, user_id = $3, category_id = $4,
			description = $5, thumbnail = $6, updated_at = $7
		WHERE id = $8
	`, title, slug.Make(title), userID, categoryID, description, thumbnail, time.Now(), id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to update data",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data updated successfully"})
}

// delete data
func DeleteArticle(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Cek apakah data ada
	var article model.Article
	err = config.DB.QueryRow(`SELECT id, thumbnail FROM articles WHERE id = $1`, id).Scan(&article.Id, &article.Thumbnail)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "detail": err.Error()})
		return
	}

	// Hapus file thumbnail jika ada
	if article.Thumbnail != "" {
		oldFilePath := filepath.Join("uploads", filepath.Base(article.Thumbnail))
		if _, err := os.Stat(oldFilePath); err == nil {
			if err := os.Remove(oldFilePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image", "detail": err.Error()})
				return
			}
		}
	}

	// Hapus data dari database
	_, err = config.DB.Exec(`DELETE FROM articles WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data deleted successfully",
	})
}

// hitung views artikel
func IncrementArticleViews(c *gin.Context) {
	slugParam := c.Param("slug")

	// Update kolom views (+1)
	result, err := config.DB.Exec(`
		UPDATE articles
		SET views = views + 1
		WHERE slug = $1
	`, slugParam)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to update views",
			"detail": err.Error(),
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Views updated +1",
	})
}
