package envmap

import (
	"fmt"
	"strings"
)

type EnvironmentMapping struct {
	Virtual string
	Real    string
}

func (c *EnvironmentMapping) Decode(value string) error {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return fmt.Errorf("parse environment mapping: invalid value %q: expected 'virtual:real'", value)
	}
	(*c).Virtual = parts[0]
	(*c).Real = parts[1]

	return nil
}
