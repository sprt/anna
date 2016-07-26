// +build integration

package psn

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

var integrationconfig *Config

func init() {
	configPath := os.Getenv("ANNA_CONFIG")
	if configPath == "" {
		fmt.Fprintln(os.Stderr, "no config")
		os.Exit(1)
	}
	configPath, err := filepath.Abs(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	configDir, configFile := path.Split(configPath)
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName(strings.TrimSuffix(configFile, path.Ext(configFile)))
	v.AddConfigPath(configDir)
	if err := v.ReadInConfig(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	integrationconfig = &Config{
		Email:        v.GetString("psn.email"),
		Username:     v.GetString("psn.username"),
		Password:     v.GetString("psn.password"),
		ClientID:     v.GetString("psn.client_id"),
		ClientSecret: v.GetString("psn.client_secret"),
		DUID:         v.GetString("psn.duid"),
	}
}

func TestIntegrationTokenSource(t *testing.T) {
	ts := newTokenSource(integrationconfig, http.DefaultClient)
	if _, err := ts.Token(); err != nil {
		t.Error(err)
	}
}
