// +build ignore

package main

import (
	"fmt"
	
	tls "github.com/refraction-networking/utls"
)

func main() {
	// Verify we can reference utls types
	var _ *tls.UConn
	var _ tls.ClientHelloID
	
	// Check ClientHello IDs are available
	browsers := []string{
		tls.HelloChrome_Auto.Client,
		tls.HelloFirefox_Auto.Client,
		tls.HelloIOS_Auto.Client,
		tls.HelloEdge_Auto.Client,
		tls.HelloSafari_Auto.Client,
	}
	
	fmt.Println("✅ utls import successful!")
	fmt.Printf("✅ Browser fingerprints available: %v\n", browsers)
	fmt.Printf("✅ TLS versions: 1.0=0x%04x, 1.1=0x%04x, 1.2=0x%04x, 1.3=0x%04x\n",
		tls.VersionTLS10, tls.VersionTLS11, tls.VersionTLS12, tls.VersionTLS13)
}

