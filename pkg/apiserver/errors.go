package apiserver

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgconn"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Handle all errors from all API endpoints by parsing their contents and/or type,
// then assigning an appropriate HTTP error code and attaching the error message as JSON.
func ErrorHandler(c *gin.Context, err error) (code int, errint interface{}) {
	// Set original error message in JSON output
	errint = gin.H{
		"error": err.Error(),
	}

	// Type-checking errors
	switch t := err.(type) {
	case *pgconn.PgError:
		switch t.Code {
		case "22P02": // invalid input syntax (for type uuid?) - INVALID TEXT REPRESENTATION
			code = http.StatusBadRequest
			return
		}
	}

	// Generic errors
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		// Generic response from database backend
		code = http.StatusNotFound
	default:
		// In case no better error code exists, we fall back to "internal server error"
		code = http.StatusInternalServerError
		log.Error(err)
	}
	return
}
