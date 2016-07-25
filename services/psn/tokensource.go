package psn

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/time/rate"

	"github.com/sprt/anna"
	"github.com/sprt/anna/services"
)

type tokenSource struct {
	*services.Client
	client *http.Client // pointer to Service.client
	mu     sync.Mutex   // guards Token
}

func newTokenSource(config *Config, client *http.Client) *tokenSource {
	ts := &tokenSource{
		Client: services.NewClient(client, rate.NewLimiter(rate.Inf, 1)),
		client: client,
	}
	return ts
}

func (ts *tokenSource) Token() (*oauth2.Token, error) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	err := ts.pullSSOCookie()
	if err != nil {
		return nil, err
	}
	grant, err := ts.grantCode()
	if err != nil {
		return nil, err
	}
	tok, date, err := ts.token(grant)
	if err != nil {
		return nil, err
	}
	err = ts.checkToken(tok)
	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		AccessToken:  tok.AccessToken,
		TokenType:    tok.TokenType,
		RefreshToken: tok.RefreshToken,
		Expiry:       date.Add(time.Duration(tok.ExpiresIn) * time.Second),
	}
	return token, nil
}

// checkToken checks the validity of a new token.
// It does not check its expiration date.
func (ts *tokenSource) checkToken(tok *token) error {
	req, err := http.NewRequest("GET", baseURL+"/2.0/oauth/token/"+tok.AccessToken, nil)
	if err != nil {
		return err
	}
	auth := []byte(anna.Config.PSNClientID + ":" + anna.Config.PSNClientSecret)
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString(auth))

	_, err = ts.Do(req)
	return err
}

func (ts *tokenSource) token(grantCode string) (*token, time.Time, error) {
	data := url.Values{}
	data.Set("client_id", anna.Config.PSNClientID)
	data.Set("client_secret", anna.Config.PSNClientSecret)
	data.Set("duid", anna.Config.PSNDuid)
	data.Set("redirect_uri", redirectURI)
	data.Set("scope", scope)
	data.Set("service_entity", serviceEntity)
	data.Set("code", grantCode)
	data.Set("grant_type", "authorization_code")
	body := strings.NewReader(data.Encode())

	req, err := http.NewRequest("POST", baseURL+"/2.0/oauth/token", body)
	if err != nil {
		return nil, time.Time{}, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := ts.Do(req)
	if err != nil {
		return nil, time.Time{}, err
	}
	defer resp.Body.Close()

	tok := new(token)
	err = json.NewDecoder(resp.Body).Decode(tok)
	if err != nil {
		return nil, time.Time{}, err
	}
	date, err := time.Parse(http.TimeFormat, resp.Header.Get("Date"))
	if err != nil {
		return nil, time.Time{}, err
	}

	return tok, date, nil
}

func (ts *tokenSource) grantCode() (string, error) {
	req, err := http.NewRequest("GET", baseURL+"/2.0/oauth/authorize", nil)
	if err != nil {
		return "", nil
	}

	q := req.URL.Query()
	q.Add("client_id", anna.Config.PSNClientID)
	q.Add("client_secret", anna.Config.PSNClientSecret)
	q.Add("duid", anna.Config.PSNDuid)
	q.Add("redirect_uri", redirectURI)
	q.Add("scope", scope)
	q.Add("service_entity", serviceEntity)
	q.Add("response_type", "code")
	req.URL.RawQuery = q.Encode()

	for _, cookie := range ts.client.Jar.Cookies(req.URL) {
		req.AddCookie(cookie)
	}

	oldCheckRedirect := ts.client.CheckRedirect
	ts.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return services.ErrUseLastResponse
	}
	defer func() {
		ts.client.CheckRedirect = oldCheckRedirect
	}()

	resp, err := ts.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return resp.Header.Get("X-NP-GRANT-CODE"), nil
}

func (ts *tokenSource) pullSSOCookie() error {
	data := url.Values{}
	data.Set("authentication_type", "password")
	data.Set("username", anna.Config.PSNEmail)
	data.Set("password", anna.Config.PSNPassword)
	data.Set("client_id", anna.Config.PSNClientID)
	body := strings.NewReader(data.Encode())

	req, err := http.NewRequest("POST", baseURL+"/2.0/ssocookie", body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := ts.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if ts.client.Jar == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return err
		}
		ts.client.Jar = jar
	}
	ts.client.Jar.SetCookies(req.URL, resp.Cookies())
	// FIXME: reset jar when done

	return nil
}

type token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}
