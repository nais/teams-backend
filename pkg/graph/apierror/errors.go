package apierror

import (
	"context"
	"errors"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/jackc/pgconn"
	log "github.com/sirupsen/logrus"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

var (
	ErrTeamSlug    = Errorf("Your team identifier does not fit our requirements. Team identifiers must contain only lowercase alphanumeric characters or hyphens, contain at least 3 characters and at most 30 characters, start with an alphabetic character, end with an alphanumeric character, and not contain two hyphens in a row.")
	ErrInternal    = Errorf("The server errored out while processing your request, and we didn't write a suitable error message. You might consider that a bug on our side. Please try again, and if the error persists, contact the NAIS team.")
	ErrDatabase    = Errorf("The database system encountered an error while processing your request. This is probably a transient error, please try again. If the error persists, contact the NAIS team.")
	ErrTeamPurpose = Errorf("You must specify the purpose for your team. This is a human-readable string which is used in external systems, and is important because other people might need to to understand what your team is all about.")
)

type Error struct {
	err error
}

func (e Error) Error() string {
	return e.err.Error()
}

// Format an error message for API users.
// This message will probably be read by a human being, format it accordingly and don't leak information!
func Errorf(format string, args ...any) Error {
	return Error{
		err: fmt.Errorf(format, args...),
	}
}

// Make sure that errors that are not formatted for API users will not leak through.
// All unspecified errors are logged.
func GetErrorPresenter() graphql.ErrorPresenterFunc {
	return func(ctx context.Context, e error) *gqlerror.Error {
		err := graphql.DefaultErrorPresenter(ctx, e)
		unwrappedError := errors.Unwrap(e)

		switch originalError := unwrappedError.(type) {
		case Error:
			// Error is already formatted for end-user consumption.
		case *pgconn.PgError:
			err.Message = ErrDatabase.Error()
			// Log error?
			log.Errorf("database error %s: %s (%s)", originalError.Code, originalError.Message, originalError.Detail)
			//err.Message = pgErr.Detail
			return err
		default:
			log.Errorf("unhandled error: %s", originalError.Error())
			err.Message = ErrInternal.Error()
		}

		return err
	}
}
