package db

import (
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
)

func MarshalUUID(uid *uuid.UUID) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		txt := strconv.Quote(uid.String())
		io.WriteString(w, txt)
	})
}

func UnmarshalUUID(v interface{}) (*uuid.UUID, error) {
	id, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("UUID must be a string")
	}

	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	return &uid, nil
}
