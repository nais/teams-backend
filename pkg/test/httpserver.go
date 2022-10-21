package test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func HttpServerWithHandlers(t *testing.T, handlers []http.HandlerFunc) *httptest.Server {
	idx := 0
	t.Cleanup(func() { assert.Equal(t, len(handlers), idx) })
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(handlers) < idx+1 {
			t.Fatalf("unexpected request, add missing handler func: %v", r)
		}
		handlers[idx](w, r)
		idx += 1
	}))
}
