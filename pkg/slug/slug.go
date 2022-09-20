package slug

import (
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
)

type Slug string

func MarshalSlug(slug *Slug) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		txt := strconv.Quote(slug.String())
		io.WriteString(w, txt)
	})
}

func UnmarshalSlug(v interface{}) (*Slug, error) {
	input, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("slug must be a string")
	}

	slug := Slug(input)
	return &slug, nil
}

func (s Slug) String() string {
	return string(s)
}

func (s Slug) StringP() *string {
	strp := string(s)
	return &strp
}
