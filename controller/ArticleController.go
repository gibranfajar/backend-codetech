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
	var article []model.ResponseArticle

	rows, err := config.DB.Query(`
		SELECT a.id, a.title, a.slug, a.description, a.thumbnail, a.views, a.created_at, a.updated_at, u.name, c.category
		FROM articles a
		JOIN users u ON a.user_id = u.id
		JOIN category_articles c ON a.category_id = c.id
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
		return
	}

	if rows == nil {
		c.JSON(http.StatusOK, gin.H{"message": "No data found"})
	}

	for rows.Next() {
		var art model.ResponseArticle
		if err := rows.Scan(&art.Id, &art.Title, &art.Slug, &art.Description, &art.Thumbnail, &art.Views, &art.CreatedAt, &art.UpdatedAt, &art.User, &art.Category); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
			return
		}
		article = append(article, art)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": article,
	})
}

// create data
func CreateArticle(c *gin.Context) {
	title := c.PostForm("title")
	user := c.PostForm("user_id")
	category := c.PostForm("category_id")
	description := c.PostForm("description")

	var req model.ArticleRequest
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

	file, err := c.FormFile("thumbnail")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thumbnail is required"})
		return
	}

	os.MkdirAll("uploads", os.ModePerm)
	filename := uuid.New().String() + filepath.Ext(file.Filename)
	savePath := "uploads/" + filename
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	thumbnail := "/uploads/" + filename

	_, err = config.DB.Exec(`
		INSERT INTO articles (title, slug, user_id, category_id, description, thumbnail, created_at, updated_at)
		VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8)
	`, title, slug.Make(title), user, category, description, thumbnail, time.Now(), time.Now())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Data created successfully",
	})
}

// update data
func UpdateArticle(c *gin.Context) {
	idParam := c.Param("id")
	title := c.PostForm("title")
	user := c.PostForm("user_id")
	category := c.PostForm("category_id")
	description := c.PostForm("description")

	var req model.ArticleRequest
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

	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Ambil data artikel termasuk thumbnail
	var article model.Article
	err = config.DB.QueryRow(
		"SELECT id, thumbnail FROM articles WHERE id = @p1",
		sql.Named("p1", id),
	).Scan(&article.Id, &article.Thumbnail)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "detail": err.Error()})
		return
	}

	// Default thumbnail tetap yang lama
	thumbnail := article.Thumbnail

	// Cek apakah ada file baru di-upload
	file, err := c.FormFile("thumbnail")
	if err == nil {
		// Upload file baru
		os.MkdirAll("uploads", os.ModePerm)
		filename := uuid.New().String() + filepath.Ext(file.Filename)
		savePath := filepath.Join("uploads", filename)

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
			return
		}

		// Hapus file lama
		if article.Thumbnail != "" {
			oldFilePath := filepath.Join("uploads", filepath.Base(article.Thumbnail))
			if _, err := os.Stat(oldFilePath); err == nil {
				os.Remove(oldFilePath)
			}
		}

		// Set thumbnail baru
		thumbnail = "/uploads/" + filename
	}

	// Update database
	_, err = config.DB.Exec(`
		UPDATE articles
		SET title = @p1, slug = @p2, user_id = @p3, category_id = @p4, description = @p5, thumbnail = @p6, updated_at = @p7
		WHERE id = @p8
	`, title, slug.Make(title), user, category, description, thumbnail, time.Now(), id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update data", "detail": err.Error()})
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

	// check database
	var article model.Article
	err = config.DB.QueryRow("SELECT id FROM articles WHERE id = @p1", sql.Named("p1", id)).Scan(&article.Id)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data not found"})
		return
	}

	// delete file lama jika ada
	var oldImage string
	err = config.DB.QueryRow("SELECT thumbnail FROM articles WHERE id = @p1", sql.Named("p1", id)).Scan(&oldImage)
	if err == nil && oldImage != "" {
		_, filename := filepath.Split(oldImage)
		os.Remove("uploads/" + filename)
	}

	_, err = config.DB.Exec("DELETE FROM articles WHERE id = @p1", sql.Named("p1", id))
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

	// Update views: tambahkan 1 ke kolom views
	result, err := config.DB.Exec(`
		UPDATE articles
		SET views = views + 1
		WHERE slug = @p1
	`, sql.Named("p1", slugParam))

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
