package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"

	"golang.org/x/time/rate"
)

var InfRateLimiter = rate.NewLimiter(rate.Inf, 1)

type Env struct {
	Server *httptest.Server
	Mux    *http.ServeMux
	Client *http.Client
}

func (env *Env) Setup() {
	env.Mux = http.NewServeMux()
	env.Server = httptest.NewServer(env.Mux)
	p, err := url.Parse(env.Server.URL)
	if err != nil {
		panic(err)
	}
	transport := &http.Transport{Proxy: http.ProxyURL(p)}
	env.Client = &http.Client{Transport: &testTransport{transport, p}}
}

func Pretty(a interface{}) string {
	b, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (env *Env) Teardown() {
	env.Server.Close()
	env.Server = nil
	env.Mux = nil
	env.Client = nil
}

type testTransport struct {
	*http.Transport
	serverURL *url.URL
}

func (tr testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = tr.serverURL.Scheme
	return tr.Transport.RoundTrip(req)
}
