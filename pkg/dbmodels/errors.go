package dbmodels

import (
	"errors"
	"net/http"

	"github.com/jackc/pgconn"
	"gorm.io/gorm"
)

// Parse a database-related error and try to return a matching HTTP error code.
func ErrorCode(err error) int {
	// Type-checking errors
	switch t := err.(type) {
	case *pgconn.PgError:
		switch t.Code {
		case "22P02": // invalid input syntax (for type uuid?) - INVALID TEXT REPRESENTATION
			return http.StatusBadRequest
		}
	}

	// Generic errors
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		// Generic response from database backend
		return http.StatusNotFound
	default:
		// In case no better error code exists, we fall back to "internal server error"
		return http.StatusInternalServerError
	}
}
