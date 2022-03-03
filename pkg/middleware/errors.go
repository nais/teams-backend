package middleware

import (
	"fmt"
)

// Structure middleware error messages the same way as GraphQL does,
// to ease handling on the client side.
func Errorf(format string, a ...interface{}) Errors {
	return Errors{
		Errors: []Error{
			{
				Message: fmt.Sprintf(format, a...),
			},
		},
	}
}

type Error struct {
	Message string `json:"message"`
}

type Errors struct {
	Errors []Error `json:"errors"`
}
