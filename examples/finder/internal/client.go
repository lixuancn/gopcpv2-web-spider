package internal

import (
	"net/http"
	"net"
	"time"
)

func  genHttpClient() *http.Client{
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout: 30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns: 100,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout: 60 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}