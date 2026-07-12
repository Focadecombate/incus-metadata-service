package configs

import (
	"database/sql"
	"net/http"
	"testing"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNetworkConfigHandler_StoredConfig(t *testing.T) {
	handler, mockDB := setupConfigsTest()

	// Stored as JSONB; served opaquely as YAML.
	stored := []byte(`{"version":2,"ethernets":{"eth0":{"dhcp4":true}}}`)
	mockDB.On("GetInstanceNetworkConfigByIP", mock.Anything, mock.Anything).
		Return(db.InstanceNetworkConfig{NetworkConfig: stored}, nil)

	// cloud-init sends Accept: */* and must NOT get a 406.
	c, w := newGetContext("*/*", "", "")
	handler.NetworkConfigHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "version: 2")
	assert.Contains(t, body, "dhcp4: true")
	mockDB.AssertExpectations(t)
}

func TestNetworkConfigHandler_None(t *testing.T) {
	handler, mockDB := setupConfigsTest()

	mockDB.On("GetInstanceNetworkConfigByIP", mock.Anything, mock.Anything).
		Return(db.InstanceNetworkConfig{}, sql.ErrNoRows)

	c, w := newGetContext("*/*", "", "")
	handler.NetworkConfigHandler(c)

	// 404 (tolerated by cloud-init), NOT 406.
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockDB.AssertExpectations(t)
}
