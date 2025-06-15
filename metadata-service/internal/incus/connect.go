package incus

import (
	"fmt"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/config"
	incus "github.com/lxc/incus/client"
)

func ConnectToIncus(config *config.Config) (incus.InstanceServer, error) {
	// Connect to the Incus server
	client, err := incus.ConnectIncus(config.Incus.ServerURL, &incus.ConnectionArgs{
		TLSClientCert: config.Incus.TLSClientCert,
		TLSClientKey:  config.Incus.TLSClientKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Incus server: %w", err)
	}

	return client, nil
}