package realtor

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

// browserTransport returns an http.RoundTripper that performs TLS handshakes
// using utls with a Chrome ClientHello so realtor.ca's Imperva fingerprint
// check is satisfied. Chrome's preset advertises h2 in ALPN, and Imperva
// negotiates h2, so the round tripper is an http2.Transport. Each connection
// dials raw TCP, completes a utls handshake, asserts h2, and returns the
// connection wrapped so http2.Transport can read its TLS state.
func browserTransport() http.RoundTripper {
	dialer := &net.Dialer{Timeout: 15 * time.Second, KeepAlive: 30 * time.Second}

	return &http2.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
			host, _, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			raw, err := dialer.DialContext(ctx, network, addr)
			if err != nil {
				return nil, err
			}
			uconn := utls.UClient(raw, &utls.Config{ServerName: host}, utls.HelloChrome_Auto)
			if err := uconn.HandshakeContext(ctx); err != nil {
				raw.Close()
				return nil, err
			}
			if proto := uconn.ConnectionState().NegotiatedProtocol; proto != http2.NextProtoTLS {
				uconn.Close()
				return nil, fmt.Errorf("expected h2, got %q", proto)
			}
			return &h2Conn{UConn: uconn}, nil
		},
		ReadIdleTimeout: 30 * time.Second,
		PingTimeout:     15 * time.Second,
	}
}

// h2Conn adapts a utls.UConn so it satisfies http2's connectionStater
// interface (which expects ConnectionState() to return crypto/tls.ConnectionState,
// not utls.ConnectionState).
type h2Conn struct {
	*utls.UConn
}

func (c *h2Conn) ConnectionState() tls.ConnectionState {
	s := c.UConn.ConnectionState()
	return tls.ConnectionState{
		Version:                     s.Version,
		HandshakeComplete:           s.HandshakeComplete,
		DidResume:                   s.DidResume,
		CipherSuite:                 s.CipherSuite,
		NegotiatedProtocol:          s.NegotiatedProtocol,
		NegotiatedProtocolIsMutual:  true,
		ServerName:                  s.ServerName,
		PeerCertificates:            s.PeerCertificates,
		VerifiedChains:              s.VerifiedChains,
		SignedCertificateTimestamps: s.SignedCertificateTimestamps,
		OCSPResponse:                s.OCSPResponse,
		TLSUnique:                   s.TLSUnique,
	}
}
