package client

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"
)

var (
	httpCannotRedirectError = errors.New("this client cannot redirect")
	disableRedirect         = func(_ *http.Request, _ []*http.Request) error {
		return httpCannotRedirectError
	}
	limitedRedirect = func(_ *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		return nil
	}
)

// Client is an http.Client with some tunable parameters.
type Client interface {
	http.RoundTripper
	HttpClient() *http.Client
	roundTripper() http.RoundTripper
	Do(req *http.Request) (*http.Response, error)
	Get(url string) (resp *http.Response, err error)
	Head(url string) (resp *http.Response, err error)
	Post(url string, contentType string, body io.Reader) (resp *http.Response, err error)
	PostForm(url string, data url.Values) (resp *http.Response, err error)
}

type wrappedClient struct {
	http.Client
}

func (cli *wrappedClient) HttpClient() *http.Client {
	return &cli.Client
}
func (cli *wrappedClient) roundTripper() http.RoundTripper {
	return cli.Transport
}

func (cli *wrappedClient) RoundTrip(req *http.Request) (*http.Response, error) {
	return cli.Client.Transport.RoundTrip(req)
}

// NewClient returns a Client interface that has some tunable parameters.
func NewClient(
	keepaliveDuration time.Duration,
	connectTimeout time.Duration,
	responseHeaderTimeout time.Duration,
	idleConnectionTimeout time.Duration,
	maxIdleConnections int,
	redirectSupport bool,
	serverName string,
) Client {

	var redirectChecker func(*http.Request, []*http.Request) error
	if redirectSupport {
		redirectChecker = limitedRedirect
	} else {
		redirectChecker = disableRedirect
	}

	return &wrappedClient{
		Client: http.Client{
			Transport:     NewRoundTripper(keepaliveDuration, connectTimeout, responseHeaderTimeout, idleConnectionTimeout, maxIdleConnections, serverName),
			CheckRedirect: redirectChecker,
			Jar:           nil,
		},
	}
}
