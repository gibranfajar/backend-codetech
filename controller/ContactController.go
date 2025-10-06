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
func GetAllContact(c *gin.Context) {
	var contacts []model.Contact

	rows, err := config.DB.Query(`
		SELECT id, phone, email, address, office_operation, created_at, updated_at 
		FROM contacts
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var contact model.Contact
		if err := rows.Scan(
			&contact.Id, &contact.Phone, &contact.Email, &contact.Address, &contact.OfficeOperation, &contact.CreatedAt, &contact.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse data", "detail": err.Error()})
			return
		}
		contacts = append(contacts, contact)
	}

	// Cek apakah ada data
	if len(contacts) == 0 {
		c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": contacts})
}

// create data
func CreateContact(c *gin.Context) {
	phone := c.PostForm("phone")
	email := c.PostForm("email")
	address := c.PostForm("address")
	officeOperation := c.PostForm("office_operation")

	var req model.ContactRequest
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

	// Cek apakah data dengan phone sudah ada
	var existingID int
	err := config.DB.QueryRow(`SELECT id FROM contacts WHERE phone = $1`, phone).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data already exists"})
		return
	} else if err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing data", "detail": err.Error()})
		return
	}

	// Insert ke database
	_, err = config.DB.Exec(`
		INSERT INTO contacts (phone, email, address, office_operation, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, phone, email, address, officeOperation, time.Now(), time.Now())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Data created successfully"})
}

// update data
func UpdateContact(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var req model.ContactRequest
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

	phone := c.PostForm("phone")
	email := c.PostForm("email")
	address := c.PostForm("address")
	officeOperation := c.PostForm("office_operation")

	// Cek apakah data ada
	var existingID int
	err = config.DB.QueryRow(`SELECT id FROM contacts WHERE id = $1`, id).Scan(&existingID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing contact", "detail": err.Error()})
		return
	}

	// Update data
	_, err = config.DB.Exec(`
		UPDATE contacts 
		SET phone = $1, email = $2, address = $3, office_operation = $4, updated_at = $5
		WHERE id = $6
	`, phone, email, address, officeOperation, time.Now(), id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data updated successfully"})
}

// delete data
func DeleteContact(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// Cek apakah data dengan ID tersebut ada
	var existingID int
	err = config.DB.QueryRow(`SELECT id FROM contacts WHERE id = $1`, id).Scan(&existingID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "Failed to fetch contact",
			"detail": err.Error(),
		})
		return
	}

	// Hapus data
	_, err = config.DB.Exec(`DELETE FROM contacts WHERE id = $1`, id)
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
