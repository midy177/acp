package pnet

import (
	"context"
	"crypto/tls"
	"flag"
	"net"
	"net/http"
)

type HTTPClient struct {
	*http.Client
	chLaddr chan net.Addr
}

var insecureSkipVerify = flag.Bool("k", false, "Http client Skip TLS Verify (default: false)")

func NewHTTPClient(useIpv6 bool, laddr string) *HTTPClient {
	client := &HTTPClient{
		Client:  &http.Client{},
		chLaddr: make(chan net.Addr, 1),
	}

	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.DialContext = func(ctx context.Context, _network, addr string) (net.Conn, error) {
		network := "tcp4"
		if useIpv6 {
			network = "tcp6"
		}
		c, err := DialContext(ctx, network, laddr, addr)
		if err == nil {
			client.chLaddr <- c.LocalAddr()
		} else {
			defaultLogger.Debugf("pnet.HTTPClient.Transport.DialContext: %v", err)
		}
		close(client.chLaddr)
		return c, err
	}
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: *insecureSkipVerify}

	client.Client.Transport = tr
	return client
}

func (cl *HTTPClient) GetLAddr() <-chan net.Addr {
	return cl.chLaddr
}
