package public

import (
	"net/http"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/content_types"
	"github.com/focadecombate/incus-metadata-service/metadata-service/pkg/types"
	"github.com/gin-gonic/gin"
)

func mockNetworkConfig() types.NetworkConfig {
	return types.NetworkConfig{
		Version: 2,
		Ethernets:  map[string]types.Ethernet{
			"eth0": {
				Match: types.Match{
					Name: "eth0",
					MacAddress: "00:11:22:33:44:55",
				},
				Addresses: []string{"192.168.1.2/24"},
				WakeOnLan: true,
				DHCP4: true,
				Gateway4: "192.168.1.1",
				Gateway6: "fe80::1",
				Nameservers: types.Nameservers{
					Search:    []string{"example.com"},
					Addresses: []string{"8.8.8.8"},
				},
				Routes: []types.Route{
					{
						To:     "192.0.2.0/24",
						Via:    "11.0.0.1",
						Metric: 3,
					},
				},
			},
		},

	}
}

func NetworkConfigHandler(c *gin.Context) {
	// This is a placeholder for the network configuration handler logic.
	// In a real application, you would retrieve and return network configuration data here.
	requestedContentType := c.GetHeader("Accept")

	if !content_types.ValidateContentType(c, requestedContentType, [][]string{content_types.YamlContentTypes}) {
		return
	}

	c.YAML(http.StatusOK, mockNetworkConfig())
}
