package client

import (
	"net"
	"net/http"
	"time"
)

const (
	userAgentHeader = "User-Agent"
)

type predefinedHeaderTransport struct {
	useragentName string
	http.Transport
}

func (pht *predefinedHeaderTransport) RoundTrip(req *http.Request) (res *http.Response, err error) {
	req.Close = pht.DisableKeepAlives
	req.Header.Set(userAgentHeader, pht.useragentName)
	res, err = pht.Transport.RoundTrip(req)
	return
}

// NewRoundTripper returns a http.RoundTripper that has some tunable parameters.
func NewRoundTripper(
	keepaliveDuration time.Duration,
	connectTimeout time.Duration,
	responseHeaderTimeout time.Duration,
	idleConnectionTimeout time.Duration,
	maxIdleConnections int,
	serverName string,
) http.RoundTripper {

	keepaliveDisabled := keepaliveDuration == 0
	dialer := &net.Dialer{
		Timeout:   connectTimeout,
		KeepAlive: keepaliveDuration,
	}

	return &predefinedHeaderTransport{
		useragentName: serverName,
		Transport: http.Transport{
			DisableKeepAlives:     keepaliveDisabled,
			DisableCompression:    true,
			MaxIdleConnsPerHost:   maxIdleConnections,
			DialContext:           dialer.DialContext,
			MaxIdleConns:          maxIdleConnections,
			IdleConnTimeout:       idleConnectionTimeout,
			ResponseHeaderTimeout: responseHeaderTimeout,
		},
	}
}
