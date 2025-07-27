package incus

import (
	"fmt"
	"os"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/config"
	incus "github.com/lxc/incus/client"
)

func ConnectToIncus(config *config.Config) (incus.InstanceServer, error) {
	// Read the certificate files
	certData, err := os.ReadFile(config.Incus.TLSClientCert)
	if err != nil {
		return nil, fmt.Errorf("failed to read client certificate file: %w", err)
	}

	keyData, err := os.ReadFile(config.Incus.TLSClientKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read client key file: %w", err)
	}

	serverCertData, err := os.ReadFile(config.Incus.TLSServerCert)
	if err != nil {
		return nil, fmt.Errorf("failed to read server certificate file: %w", err)
	}

	// Connect to the Incus server with TLS configuration
	client, err := incus.ConnectIncus(config.Incus.ServerURL, &incus.ConnectionArgs{
		TLSClientCert:      string(certData),
		TLSClientKey:       string(keyData),
		InsecureSkipVerify: config.Incus.TLSInsecureSkipVerify,
		TLSServerCert:      string(serverCertData),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Incus server: %w", err)
	}

	return client, nil
}