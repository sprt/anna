// +build integration

package psn

import (
	"net/http"
	"testing"
)

func TestIntegrationTokenSource(t *testing.T) {
	ts := newTokenSource(http.DefaultClient)
	if _, err := ts.Token(); err != nil {
		t.Error(err)
	}
}
