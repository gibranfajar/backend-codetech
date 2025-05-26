package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gibranfajar/backend-codetech/config"
	"github.com/gibranfajar/backend-codetech/model"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// get all category
func GetAllCategoryArticle(c *gin.Context) {
	var categoryArticles []model.CategoryArticle

	rows, err := config.DB.Query("SELECT id, category, created_at, updated_at FROM category_articles")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
		return
	}

	if rows == nil {
		c.JSON(http.StatusOK, gin.H{"message": "No data found"})
	}

	for rows.Next() {
		var categoryArticle model.CategoryArticle
		if err := rows.Scan(&categoryArticle.Id, &categoryArticle.Category, &categoryArticle.CreatedAt, &categoryArticle.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
			return
		}
		categoryArticles = append(categoryArticles, categoryArticle)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": categoryArticles,
	})
}

// create category article
func CreateCategoryArticle(c *gin.Context) {
	category := c.PostForm("category")

	var req model.CategoryArticleRequest
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

	_, err = config.DB.Exec(`
		INSERT INTO category_articles (category, created_at, updated_at)
		VALUES (@p1, @p2, @p3)
	`, category, time.Now(), time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Data created successfully",
	})
}

// update category article
func UpdateCategoryArticle(c *gin.Context) {
	id := c.Param("id")
	category := c.PostForm("category")

	var req model.CategoryArticleRequest
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

	// check database
	var categoryArticle model.CategoryArticle
	err = config.DB.QueryRow("SELECT id FROM category_articles WHERE id = @p1", sql.Named("p1", id)).Scan(&categoryArticle.Id)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data not found"})
		return
	}

	_, err = config.DB.Exec(`
		UPDATE category_articles
		SET category = @p1, updated_at = @p2
		WHERE id = @p3
	`, category, time.Now(), id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data updated successfully",
	})
}

// delete category article
func DeleteCategoryArticle(c *gin.Context) {
	id := c.Param("id")

	// check database
	var categoryArticle model.CategoryArticle
	err := config.DB.QueryRow("SELECT id FROM category_articles WHERE id = @p1", sql.Named("p1", id)).Scan(&categoryArticle.Id)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data not found"})
		return
	}

	_, err = config.DB.Exec(`
		DELETE FROM category_articles
		WHERE id = @p1
	`, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data deleted successfully",
	})
}
