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
	var contact model.Contact

	err := config.DB.QueryRow("SELECT id, phone, email, address, office_operation, created_at, updated_at FROM contacts").Scan(
		&contact.Id, &contact.Phone, &contact.Email, &contact.Address, &contact.OfficeOperation, &contact.CreatedAt, &contact.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{"message": "No data found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": contact,
	})

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
	err := config.Validate.Struct(req)
	if err != nil {
		errors := []string{}
		for _, err := range err.(validator.ValidationErrors) {
			errors = append(errors, fmt.Sprintf("%s is %s", err.Field(), err.Tag()))
		}
		c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
		return
	}

	// check apakah data sudah ada atau tidak
	var contact model.Contact
	err = config.DB.QueryRow("SELECT id FROM contacts WHERE phone = @p1", sql.Named("p1", phone)).Scan(&contact.Id)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data already exists"})
		return
	}

	_, err = config.DB.Exec(`INSERT INTO contacts (phone, email, address, office_operation, created_at, updated_at) VALUES (@p1, @p2, @p3, @p4, @p5, @p6)`,
		phone, email, address, officeOperation, time.Now(), time.Now())

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Data created successfully",
	})
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
	err = config.Validate.Struct(req)
	if err != nil {
		errors := []string{}
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

	// check apakah data ada dengan id tersebut
	var contact model.Contact
	err = config.DB.QueryRow("SELECT id FROM contacts WHERE id = @p1", sql.Named("p1", id)).Scan(&contact.Id)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data not found"})
		return
	}

	_, err = config.DB.Exec(`UPDATE contacts SET phone = @p1, email = @p2, address = @p3, office_operation = @p4, updated_at = @p5 WHERE id = @p6`,
		phone, email, address, officeOperation, time.Now(), id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data updated successfully",
	})
}

// delete data
func DeleteContact(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	// check apakah data ada dengan id tersebut
	var contact model.Contact
	err = config.DB.QueryRow("SELECT id FROM contacts WHERE id = @p1", sql.Named("p1", id)).Scan(&contact.Id)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data not found"})
		return
	}

	_, err = config.DB.Exec(`DELETE FROM contacts WHERE id = @p1`, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete data", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data deleted successfully",
	})

}
