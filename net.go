package main

import (
	"net"
	"time"

	"code.google.com/p/go.net/proxy"
)

type ConnHandler struct {
	ProxyNetwork  string
	ProxyAddress  string
	ProxyUsername string
	ProxyPassword string
	Timeout       time.Duration
}

func (handler *ConnHandler) HandleConnection(network, addr string) (conn net.Conn, err error) {

	forwardDialer := &net.Dialer{

		Timeout:   handler.Timeout * time.Second,
		DualStack: true,
	}

	if len(handler.ProxyAddress) > 1 {

		auth := &proxy.Auth{

			User:     handler.ProxyUsername,
			Password: handler.ProxyPassword,
		}

		// setup the socks proxy
		dialer, err := proxy.SOCKS5(handler.ProxyNetwork, handler.ProxyAddress, auth, forwardDialer)
		if err != nil {

			return nil, err
		}
		return dialer.Dial(network, addr)
	}

	return forwardDialer.Dial(network, addr)
}
