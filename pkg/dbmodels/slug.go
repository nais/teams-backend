package dbmodels

import (
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"io"
	"regexp"
	"strconv"
)

type Slug string

var (
	re = regexp.MustCompile("^[a-z][a-z-]{1,18}[a-z]$")
)

func MarshalSlug(slug *Slug) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		txt := strconv.Quote(slug.String())
		io.WriteString(w, txt)
	})
}

func UnmarshalSlug(v interface{}) (*Slug, error) {
	input, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("Slug must be a string")
	}

	slug := Slug(input)
	err := slug.Validate()

	if err != nil {
		return nil, err
	}

	return &slug, nil
}

func (s Slug) Validate() error {
	match := re.MatchString(s.String())

	if !match {
		return fmt.Errorf("Invalid slug")
	}

	return nil
}

func (s Slug) String() string {
	return string(s)
}

func (s Slug) StringP() *string {
	strp := string(s)
	return &strp
}
