package authz

import (
	"fmt"
)

type ErrNotAuthorized struct {
	role string
}

func (e ErrNotAuthorized) Error() string {
	return fmt.Sprintf("required role: %q", e.role)
}

func (e ErrNotAuthorized) Role() string {
	return e.role
}
