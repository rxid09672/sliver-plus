package httpclient

/*
	Sliver Implant Framework
	Copyright (C) 2019-2025  Bishop Fox

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with this program.  If not, see <https://www.gnu.org/licenses/>.

	------------------------------------------------------------------------
	uTLS Integration for TLS Fingerprinting
	
	This driver uses the uTLS library (github.com/refraction-networking/utls)
	to customize TLS ClientHello fingerprints, mimicking popular browsers
	(Chrome, Firefox, Edge, iOS Safari) or generating randomized handshakes.
	
	This helps evade TLS fingerprinting (JA3) detection by making Sliver
	implants blend in with normal HTTPS traffic.
	
	Note: uTLS is pure Go and cross-platform compatible (Linux, Windows, macOS).
	------------------------------------------------------------------------
*/

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	utls "github.com/refraction-networking/utls"
)

// getBrowserID returns the ClientHelloID for a given fingerprint string
func getBrowserID(fingerprint string) utls.ClientHelloID {
	switch fingerprint {
	case "chrome":
		return utls.HelloChrome_Auto
	case "firefox":
		return utls.HelloFirefox_Auto
	case "ios":
		return utls.HelloIOS_Auto
	case "safari":
		return utls.HelloSafari_Auto
	case "edge":
		return utls.HelloEdge_Auto
	case "randomized":
		return utls.HelloRandomized
	case "randomized-alpn":
		return utls.HelloRandomizedALPN
	case "randomized-noalpn":
		return utls.HelloRandomizedNoALPN
	default:
		// Safe default: randomized with ALPN (good for stealth)
		return utls.HelloRandomizedALPN
	}
}

// dialTLS establishes a TLS connection using utls for fingerprinting
func dialTLS(ctx context.Context, network, addr string, opts *HTTPOptions, fingerprint string) (net.Conn, error) {
	// 1. Create TCP dialer with timeout
	dialer := &net.Dialer{
		Timeout: opts.NetTimeout,
	}

	// 2. Establish raw TCP connection
	tcpConn, err := dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, fmt.Errorf("tcp dial failed: %w", err)
	}

	// 3. Extract SNI (Server Name Indication) from address
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		tcpConn.Close()
		return nil, fmt.Errorf("failed to parse host from address: %w", err)
	}

	// 4. Create uTLS config
	// Note: InsecureSkipVerify is true because Sliver does app-layer encryption
	// and typically uses self-signed or domain-fronted certificates
	tlsConfig := &utls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// 5. Create uTLS connection with chosen fingerprint
	clientHelloID := getBrowserID(fingerprint)
	uconn := utls.UClient(tcpConn, tlsConfig, clientHelloID)

	// 6. Complete the TLS handshake
	// Note: Unlike stdlib tls, uTLS requires explicit Handshake() call
	if err := uconn.Handshake(); err != nil {
		tcpConn.Close()
		return nil, fmt.Errorf("utls handshake failed: %w", err)
	}

	return uconn, nil
}

// UTLSHTTPDriver creates an HTTP driver that uses uTLS for custom TLS fingerprinting
func UTLSHTTPDriver(origin string, secure bool, opts *HTTPOptions) (HTTPDriver, error) {
	// Get the TLS fingerprint from options (default will be set by caller)
	fingerprint := opts.TLSFingerprint
	if fingerprint == "" {
		fingerprint = "randomized-alpn" // Safe default
	}

	// Create HTTP transport
	transport := &http.Transport{
		IdleConnTimeout: time.Millisecond,
		MaxIdleConns:    10,
	}

	// For HTTPS, override DialTLSContext to use uTLS
	if secure {
		transport.DialTLSContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialTLS(ctx, network, addr, opts, fingerprint)
		}
		transport.TLSHandshakeTimeout = opts.TlsTimeout
	}

	// Create HTTP client
	client := &http.Client{
		Jar:       cookieJar(),
		Timeout:   opts.NetTimeout,
		Transport: transport,
	}

	// Apply proxy configuration (reuse existing function)
	parseProxyConfig(origin, transport, opts.ProxyConfig)

	return client, nil
}

