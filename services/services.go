package services

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/time/rate"
)

const (
	defaultTimeout   = 30 * time.Second
	defaultRateLimit = 10 * time.Second
)

var ErrUseLastResponse = errors.New("use last response")

type Client struct {
	rl     *rate.Limiter // limits HTTP requests
	client *http.Client
}

func NewClient(client *http.Client, rl *rate.Limiter) *Client {
	if client.Timeout == 0 {
		client.Timeout = defaultTimeout
	}
	if rl == nil {
		rl = rate.NewLimiter(rate.Every(defaultRateLimit), 1)
	}
	return &Client{
		client: client,
		rl:     rl,
	}
}

func (s *Client) Do(req *http.Request) (*http.Response, error) {
	err := s.rl.Wait(context.TODO())
	if err != nil {
		return nil, err
	}

	log.Printf("%s %s", req.Method, req.URL)
	reqDump := reqDump(req)

	resp, err := s.client.Do(req)
	if err, ok := err.(*url.Error); ok && err.Err == ErrUseLastResponse {
		return resp, nil
	}
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 400 {
		return resp, nil
	}

	log.Print(reqDump)
	log.Print(respDump(resp))

	respErr := &ResponseError{Response: resp}
	if resp.StatusCode >= 500 {
		err = (*ServerError)(respErr)
	} else { // >= 400
		err = (*ClientError)(respErr)
	}
	return resp, err
}

type ClientError ResponseError

func (e *ClientError) Error() string {
	return (*ResponseError)(e).Error()
}

type ServerError ResponseError

func (e *ServerError) Error() string {
	return (*ResponseError)(e).Error()
}

type ResponseError struct {
	Response *http.Response
}

func (e *ResponseError) Error() string {
	req := e.Response.Request
	return fmt.Sprintf("%s %s: %s", req.Method, req.URL, e.Response.Status)
}

func reqDump(req *http.Request) string {
	reqDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return fmt.Sprintf("ERROR dumping request: %s", err)
	}
	return string(reqDump)
}

func respDump(resp *http.Response) string {
	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return fmt.Sprintf("ERROR dumping response: %s", err)
	}
	return string(respDump)
}
