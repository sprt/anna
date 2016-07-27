package psn

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/sprt/anna/services/internal/testutil"
	"golang.org/x/oauth2"
)

func TestTokenSource(t *testing.T) {
	env.Setup()
	defer env.Teardown()

	env.Mux.HandleFunc("/2.0/ssocookie", func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
			t.Fatalf("Content-Type is %q, want %q", got, "application/x-www-form-urlencoded")
		}
		w.Header().Set("Set-Cookie", "npsso=foo; expires=Tue, 21-Jul-2116 15:27:01 GMT; path=/")
	})

	env.Mux.HandleFunc("/2.0/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-NP-GRANT-CODE", "bar")
	})

	env.Mux.HandleFunc("/2.0/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Date", "Thu, 21 Jul 2016 16:51:34 GMT")
		fmt.Fprint(w, `{
			"access_token": "foo",
			"token_type": "bearer",
			"refresh_token": "bar",
			"expires_in": 3599,
			"scope": "psn:sceapp user:account.get user:account.realName.get kamaji:ugc:distributor user:account.settings.privacy.update user:account.realName.update kamaji:get_account_hash user:account.settings.privacy.get oauth:manage_device_usercodes"
		}`)
	})

	env.Mux.HandleFunc("/2.0/oauth/token/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	want := &oauth2.Token{
		AccessToken:  "foo",
		TokenType:    "bearer",
		RefreshToken: "bar",
		Expiry:       time.Date(2016, 7, 21, 17, 51, 33, 0, time.UTC),
	}

	s := NewTokenSource(&Config{}, env.Client)
	got, err := s.Token()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got:\n%s\nwant:\n%s", testutil.Pretty(got), testutil.Pretty(want))
	}
}
