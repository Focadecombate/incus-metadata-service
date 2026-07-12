package configs

import (
	"database/sql"
	"net/http"
	"testing"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserDataHandler_WithUserData(t *testing.T) {
	handler, mockDB := setupConfigsTest()

	raw := "#cloud-config\npackages:\n  - nginx\n"
	mockDB.On("GetInstanceUserDataByIP", mock.Anything, mock.Anything).
		Return(db.InstanceUserDatum{UserData: raw}, nil)

	c, w := newGetContext("*/*", "", "")
	handler.UserDataHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, raw, w.Body.String())
	mockDB.AssertExpectations(t)
}

func TestUserDataHandler_KnownInstanceNoUserData(t *testing.T) {
	handler, mockDB := setupConfigsTest()

	// No user-data row for the instance...
	mockDB.On("GetInstanceUserDataByIP", mock.Anything, mock.Anything).
		Return(db.InstanceUserDatum{}, sql.ErrNoRows)
	// ...but the instance itself exists (known IP).
	mockDB.On("GetInstanceByIP", mock.Anything, mock.Anything).
		Return(db.Instance{ID: 1}, nil)

	c, w := newGetContext("*/*", "", "")
	handler.UserDataHandler(c)

	// A 200 with an empty body keeps cloud-init's datasource alive.
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Body.String())
	mockDB.AssertExpectations(t)
}

func TestUserDataHandler_UnknownIP(t *testing.T) {
	handler, mockDB := setupConfigsTest()

	mockDB.On("GetInstanceUserDataByIP", mock.Anything, mock.Anything).
		Return(db.InstanceUserDatum{}, sql.ErrNoRows)
	mockDB.On("GetInstanceByIP", mock.Anything, mock.Anything).
		Return(db.Instance{}, sql.ErrNoRows)

	c, w := newGetContext("*/*", "", "")
	handler.UserDataHandler(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockDB.AssertExpectations(t)
}
