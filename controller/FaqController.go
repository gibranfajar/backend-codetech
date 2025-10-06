package controller

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gibranfajar/backend-codetech/config"
	"github.com/gibranfajar/backend-codetech/model"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// get all data
func GetAllFaq(c *gin.Context) {
	var faqs []model.FaqResponse

	rows, err := config.DB.Query(`
		SELECT
			f.id,
			f.question,
			f.answer,
			c.category,
			f.created_at,
			f.updated_at
		FROM faqs f
		JOIN category_faqs c ON f.category_id = c.id
		ORDER BY f.created_at DESC
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
		var faq model.FaqResponse
		if err := rows.Scan(
			&faq.Id,
			&faq.Question,
			&faq.Answer,
			&faq.Category,
			&faq.CreatedAt,
			&faq.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":  "Failed to parse data",
				"detail": err.Error(),
			})
			return
		}
		faqs = append(faqs, faq)
	}

	// Cek jika tidak ada data
	if len(faqs) == 0 {
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": faqs})
}

// create data
func CreateFaq(c *gin.Context) {
	question := c.PostForm("question")
	answer := c.PostForm("answer")
	categoryId := c.PostForm("category_id")

	var req model.FaqRequest
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

	// Insert ke database
	_, err := config.DB.Exec(`
		INSERT INTO faqs (question, answer, category_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, question, answer, categoryId, time.Now(), time.Now())

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
func UpdateFaq(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req model.FaqRequest
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

	// Check data exist
	var faq model.Faq
	err = config.DB.QueryRow("SELECT id FROM faqs WHERE id = $1", id).Scan(&faq.Id)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "detail": err.Error()})
		return
	}

	question := c.PostForm("question")
	answer := c.PostForm("answer")
	categoryId := c.PostForm("category_id")

	// Update database
	_, err = config.DB.Exec(`
		UPDATE faqs
		SET question = $1, answer = $2, category_id = $3, updated_at = $4
		WHERE id = $5
	`, question, answer, categoryId, time.Now(), id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data updated successfully",
	})
}

// delete data
func DeleteFaq(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Cek apakah data ada
	var faq model.Faq
	err = config.DB.QueryRow("SELECT id FROM faqs WHERE id = $1", id).Scan(&faq.Id)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "detail": err.Error()})
		return
	}

	// Hapus data
	_, err = config.DB.Exec("DELETE FROM faqs WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data deleted successfully",
	})
}
