package otter

import (
	"net/http"
)

type Otter struct {
	client    http.Client
	Endpoint  string
	Username  string
	Password  string
	cookieStr string
}

func NewOtter(endpoint, username, password string) *Otter {
	return &Otter{
		client: http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		Endpoint: endpoint,
		Username: username,
		Password: password,
	}
}
