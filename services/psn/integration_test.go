// +build integration

package psn

import (
	"net/http"
	"testing"

	"github.com/sprt/anna"
)

var integrationConfig = &Config{
	Username:     anna.Config.PSNUsername,
	Email:        anna.Config.PSNEmail,
	Password:     anna.Config.PSNPassword,
	ClientID:     anna.Config.PSNClientID,
	ClientSecret: anna.Config.PSNClientSecret,
	DUID:         anna.Config.PSNDuid,
}

func TestIntegrationTokenSource(t *testing.T) {
	ts := newTokenSource(integrationConfig, http.DefaultClient)
	if _, err := ts.Token(); err != nil {
		t.Error(err)
	}
}
