package slug

import (
	"context"
	"fmt"
	"io"
	"strconv"
)

type Slug string

func (s Slug) MarshalGQLContext(_ context.Context, w io.Writer) error {
	txt := strconv.Quote(s.String())
	_, err := io.WriteString(w, txt)
	return err
}

func (s *Slug) UnmarshalGQLContext(_ context.Context, v interface{}) error {
	input, ok := v.(string)
	if !ok {
		return fmt.Errorf("slug must be a string")
	}

	*s = Slug(input)
	return nil
}

func (s Slug) String() string {
	return string(s)
}
