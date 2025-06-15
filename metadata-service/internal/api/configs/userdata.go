package configs

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/content_types"
	"github.com/focadecombate/incus-metadata-service/metadata-service/pkg/types"
)

func mockUserData() types.UserData {
	return types.UserData{
		Hostname:       "example-host",
		ManageEtcHosts: true,
		Users: []types.User{
			{
				Name:              "example-user",
				SSHAuthorizedKeys: []string{"ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEArD1..."},
				Groups:            []string{"sudo", "docker"},
				Shell:             "/bin/bash",
				Sudo:              "ALL=(ALL) NOPASSWD:ALL",
			},
		},
		Packages:       []string{"curl", "git"},
		PackageUpdate:  true,
		PackageUpgrade: true,
		WriteFiles: []types.File{
			{
				Path:    "/etc/hosts",
				Content: "127.0.0.1 example-host",
			},
		},
		RunCommands:  []string{"echo 'Hello, World!'"},
		FinalMessage: "User data applied successfully.",
	}
}

func (h *Handler) UserDataHandler(c *gin.Context) {
	// This is a placeholder for the user data handler logic.
	// In a real application, you would retrieve and return user data here.
	requested_content_type := c.GetHeader("Accept")

	if !content_types.ValidateContentType(c, requested_content_type, [][]string{content_types.ScriptContentTypes, content_types.YamlContentTypes}) {
		return
	}

	if content_types.IsYamlContentType(requested_content_type) {
		c.YAML(http.StatusOK, mockUserData())
		return
	}

	// Need to implement the conversion to script format if requested.

	c.YAML(http.StatusOK, mockUserData())
}
