package configs

import (
	"database/sql"
	"net/http"
	"testing"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVendorDataHandler_PerInstance(t *testing.T) {
	handler, mockDB := setupConfigsTest()

	raw := "#cloud-config\nruncmd:\n  - echo hi\n"
	mockDB.On("GetInstanceVendorDataByIP", mock.Anything, mock.Anything).
		Return(db.InstanceVendorDatum{VendorData: raw}, nil)

	c, w := newGetContext("*/*", "", "")
	handler.VendorDataHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, raw, w.Body.String())
	mockDB.AssertExpectations(t)
}

func TestVendorDataHandler_GlobalFallback(t *testing.T) {
	handler, mockDB := setupConfigsTest()

	mockDB.On("GetInstanceVendorDataByIP", mock.Anything, mock.Anything).
		Return(db.InstanceVendorDatum{}, sql.ErrNoRows)
	mockDB.On("GetVendorData", mock.Anything, "default").
		Return(db.GetVendorDataRow{Data: []byte(`{"foo":"bar"}`)}, nil)

	c, w := newGetContext("*/*", "", "")
	handler.VendorDataHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "bar")
	mockDB.AssertExpectations(t)
}

func TestVendorDataHandler_NoDataAtAll(t *testing.T) {
	handler, mockDB := setupConfigsTest()

	mockDB.On("GetInstanceVendorDataByIP", mock.Anything, mock.Anything).
		Return(db.InstanceVendorDatum{}, sql.ErrNoRows)
	mockDB.On("GetVendorData", mock.Anything, "default").
		Return(db.GetVendorDataRow{}, sql.ErrNoRows)

	c, w := newGetContext("*/*", "", "")
	handler.VendorDataHandler(c)

	// cloud-init tolerates a vendor-data 404 cleanly; a bare "{}" warns.
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockDB.AssertExpectations(t)
}
