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
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
		return
	}

	if rows == nil {
		c.JSON(http.StatusOK, gin.H{"message": "No data found"})
	}

	for rows.Next() {
		var faq model.FaqResponse
		if err := rows.Scan(&faq.Id, &faq.Question, &faq.Answer, &faq.Category, &faq.CreatedAt, &faq.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
			return
		}
		faqs = append(faqs, faq)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": faqs,
	})
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
	err := config.Validate.Struct(req)
	if err != nil {
		errors := []string{}
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	_, err = config.DB.Exec(`INSERT INTO faqs (question, answer, category_id, created_at, updated_at)
	VALUES (@p1, @p2, @p3, @p4, @p5)`,
		question, answer, categoryId, time.Now(), time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
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
	err = config.Validate.Struct(req)
	if err != nil {
		errors := []string{}
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	// check database
	var faq model.Faq
	err = config.DB.QueryRow("SELECT id FROM faqs WHERE id = @p1", sql.Named("p1", id)).Scan(&faq.Id)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data not found"})
		return
	}

	question := c.PostForm("question")
	answer := c.PostForm("answer")
	categoryId := c.PostForm("category_id")

	_, err = config.DB.Exec(`UPDATE faqs SET question = @p1, answer = @p2, category_id = @p3, updated_at = @p4 WHERE id = @p5`,
		question, answer, categoryId, time.Now(), id)

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

	// check database
	var faq model.Faq
	err = config.DB.QueryRow("SELECT id FROM faqs WHERE id = @p1", sql.Named("p1", id)).Scan(&faq.Id)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data not found"})
		return
	}

	_, err = config.DB.Exec(`DELETE FROM faqs WHERE id = @p1`, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data deleted successfully",
	})
}
