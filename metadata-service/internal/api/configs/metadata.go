package configs

import (
	"net/http"
	"slices"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/content_types"
	"github.com/focadecombate/incus-metadata-service/metadata-service/pkg/types"
	"github.com/gin-gonic/gin"
)


func mockMetadata() types.Metadata {
	return types.Metadata{
		InstanceID:       "1234567890",
		Hostname:         "example-host",
		LocalHostname:    "example-local-host",
		AvailabilityZone: "us-west-1a",
		Region:           "us-west-1",
		LocalIPv4:        "192.168.1.1",
		LocalIPv6:        "fe80::1",
		PublicIPv4:       "203.0.113.1",
		PublicIPv6:       "2001:db8::1",
		PublicKeys:       []string{"ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEArD1..."},
		SecurityGroups:   []string{"sg-12345678", "sg-87654321"},
		Placement: types.Placement{
			HostID:           "host-123456",
			AvailabilityZone: "us-west-1a",
			Region:           "us-west-1",
			Project:          "example-project",
		},
		Network: types.Network{
			Interfaces: types.Interfaces{
				Macs: map[string]types.Mac{
					"eth0": {
						DeviceNumber:  "0",
						LocalHostname: "example-local-host",
						LocalIPv4:     "192.168.1.1",
						LocalIPv6:     "fe80::1",
						PublicIPv4:    "203.0.113.1",
						PublicIPv6:    "2001:db8::1",
						Mac:           "00:11:22:33:44:55",
					},
				},
			},
		}}
}

func (h *Handler) AllMetadataHandler(c *gin.Context) {
	// This is a placeholder for the metadata handler logic.
	// In a real application, you would retrieve and return metadata here.

	requested_content_type := c.GetHeader("Accept")

	if !content_types.ValidateContentType(c, requested_content_type, [][]string{content_types.JsonContentTypes, content_types.YamlContentTypes}) {
		return
	}
	// If the content type is allowed, proceed with the response.

	metadata := mockMetadata()

	// Return the metadata in the requested format

	if slices.Contains(content_types.JsonContentTypes, requested_content_type) {
		c.JSON(http.StatusOK, metadata)
		return
	}

	c.YAML(http.StatusOK, metadata)
}

func (h *Handler) MetadataByKeyHandler(c *gin.Context) {
	key := c.Param("key")
	// This is a placeholder for the metadata by key handler logic.
	// In a real application, you would retrieve and return metadata for the given key here.
	requested_content_type := c.GetHeader("Accept")

	if !content_types.ValidateContentType(c, requested_content_type, [][]string{content_types.JsonContentTypes, content_types.YamlContentTypes}) {
		return
	}

	if slices.Contains(content_types.JsonContentTypes, requested_content_type) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Metadata by key endpoint hit",
			"key":     key,
			"data":    "This is where metadata for the key would be returned.",
		})
		return
	}

	c.YAML(http.StatusOK, gin.H{
		"message": "Metadata by key endpoint hit",
		"key":     key,
		"data":    "This is where metadata for the key would be returned.",
	})
}
