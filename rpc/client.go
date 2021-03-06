package rpc

import (
	"net/url"

	"github.com/powerman/rpc-codec/jsonrpc2"
)

type Client struct {
	*jsonrpc2.Client
}

func NewClient(addr string) (*Client, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	client := jsonrpc2.NewHTTPClient(httpHeader + u.Host + Path)
	return &Client{client}, nil
}
