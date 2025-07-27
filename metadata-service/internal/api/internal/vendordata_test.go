package internal_routes

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/config"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Helper function to create a test handler with mock database
func setupTestHandler() (*Handler, *mocks.MockQuerier) {
	mockDB := &mocks.MockQuerier{}
	handler := &Handler{
		Config:   &config.Config{},
		Database: mockDB,
	}
	return handler, mockDB
}

// Helper function to create test context with gin
func setupTestContext(method, path string, body interface{}) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}

	c.Request = httptest.NewRequest(method, path, bytes.NewBuffer(bodyBytes))
	c.Request.Header.Set("Content-Type", "application/json")
	return c, w
}

func TestUpdateVendorData_Success(t *testing.T) {
	handler, mockDB := setupTestHandler()

	// Setup test data
	vendorName := "test-vendor"
	requestData := AddVendorDataKeyRequest{
		Data: map[string]any{
			"key1": "value1",
			"key2": 123,
		},
	}

	existingVendorData := db.GetVendorDataRow{
		ID:          1,
		Name:        vendorName,
		Description: nil,
		CreatedAt:   nil,
		UpdatedAt:   nil,
		Data:        []byte(`{"old_key": "old_value"}`),
	}

	updatedVendorData := db.VendorDatum{
		ID:          1,
		Name:        vendorName,
		Description: nil,
		CreatedAt:   nil,
		UpdatedAt:   nil,
		DeletedAt:   nil,
		Data:        []byte(`{"key1": "value1", "key2": 123}`),
	}

	// Setup mocks
	mockDB.On("GetVendorData", mock.Anything, vendorName).Return(existingVendorData, nil)
	mockDB.On("UpdateVendorData", mock.Anything, mock.AnythingOfType("db.UpdateVendorDataParams")).Return(updatedVendorData, nil)

	// Setup context
	c, w := setupTestContext("PUT", "/vendor/"+vendorName, requestData)
	c.Params = gin.Params{{Key: "vendor_name", Value: vendorName}}

	// Execute
	handler.UpdateVendorData(c)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Vendor data updated successfully", response["message"])

	mockDB.AssertExpectations(t)
}

func TestUpdateVendorData_MissingVendorName(t *testing.T) {
	handler, _ := setupTestHandler()

	requestData := AddVendorDataKeyRequest{
		Data: map[string]any{"key1": "value1"},
	}

	c, w := setupTestContext("PUT", "/vendor/", requestData)
	c.Params = gin.Params{{Key: "vendor_name", Value: ""}}

	handler.UpdateVendorData(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Vendor name is required", response["error"])
}

func TestUpdateVendorData_InvalidJSON(t *testing.T) {
	handler, _ := setupTestHandler()

	c, w := setupTestContext("PUT", "/vendor/test-vendor", nil)
	c.Params = gin.Params{{Key: "vendor_name", Value: "test-vendor"}}
	c.Request = httptest.NewRequest("PUT", "/vendor/test-vendor", bytes.NewBufferString("invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateVendorData(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request payload", response["error"])
}

func TestUpdateVendorData_VendorNotFound(t *testing.T) {
	handler, mockDB := setupTestHandler()

	vendorName := "nonexistent-vendor"
	requestData := AddVendorDataKeyRequest{
		Data: map[string]any{"key1": "value1"},
	}

	mockDB.On("GetVendorData", mock.Anything, vendorName).Return(db.GetVendorDataRow{}, sql.ErrNoRows)

	c, w := setupTestContext("PUT", "/vendor/"+vendorName, requestData)
	c.Params = gin.Params{{Key: "vendor_name", Value: vendorName}}

	handler.UpdateVendorData(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to retrieve vendor data", response["error"])

	mockDB.AssertExpectations(t)
}

func TestUpdateVendorData_DatabaseUpdateError(t *testing.T) {
	handler, mockDB := setupTestHandler()

	vendorName := "test-vendor"
	requestData := AddVendorDataKeyRequest{
		Data: map[string]any{"key1": "value1"},
	}

	existingVendorData := db.GetVendorDataRow{
		ID:   1,
		Name: vendorName,
		Data: []byte(`{"old_key": "old_value"}`),
	}

	mockDB.On("GetVendorData", mock.Anything, vendorName).Return(existingVendorData, nil)
	mockDB.On("UpdateVendorData", mock.Anything, mock.AnythingOfType("db.UpdateVendorDataParams")).Return(db.VendorDatum{}, assert.AnError)

	c, w := setupTestContext("PUT", "/vendor/"+vendorName, requestData)
	c.Params = gin.Params{{Key: "vendor_name", Value: vendorName}}

	handler.UpdateVendorData(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to update vendor data", response["error"])

	mockDB.AssertExpectations(t)
}

func TestCreateVendorData_Success(t *testing.T) {
	handler, mockDB := setupTestHandler()

	vendorName := "new-vendor"
	description := "Test vendor"
	requestData := CreateVendorDataRequest{
		VendorName:  vendorName,
		Description: &description,
		Data: map[string]any{
			"key1": "value1",
			"key2": 123,
		},
	}

	createdVendorData := db.VendorDatum{
		ID:          1,
		Name:        vendorName,
		Description: &description,
		CreatedAt:   nil,
		UpdatedAt:   nil,
		DeletedAt:   nil,
		Data:        []byte(`{"key1": "value1", "key2": 123}`),
	}

	// Mock that vendor doesn't exist (returns ErrNoRows)
	mockDB.On("GetVendorData", mock.Anything, vendorName).Return(db.GetVendorDataRow{}, sql.ErrNoRows)
	mockDB.On("CreateVendorData", mock.Anything, mock.AnythingOfType("db.CreateVendorDataParams")).Return(createdVendorData, nil)

	c, w := setupTestContext("POST", "/vendor", requestData)

	handler.CreateVendorData(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Vendor data created successfully", response["message"])

	mockDB.AssertExpectations(t)
}

func TestCreateVendorData_InvalidJSON(t *testing.T) {
	handler, _ := setupTestHandler()

	c, w := setupTestContext("POST", "/vendor", nil)
	c.Request = httptest.NewRequest("POST", "/vendor", bytes.NewBufferString("invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.CreateVendorData(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid request payload", response["error"])
}

func TestCreateVendorData_VendorAlreadyExists(t *testing.T) {
	handler, mockDB := setupTestHandler()

	vendorName := "existing-vendor"
	requestData := CreateVendorDataRequest{
		VendorName: vendorName,
		Data:       map[string]any{"key1": "value1"},
	}

	existingVendorData := db.GetVendorDataRow{
		ID:   1,
		Name: vendorName,
		Data: []byte(`{"existing": "data"}`),
	}

	mockDB.On("GetVendorData", mock.Anything, vendorName).Return(existingVendorData, nil)

	c, w := setupTestContext("POST", "/vendor", requestData)

	handler.CreateVendorData(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Vendor data already exists", response["error"])

	mockDB.AssertExpectations(t)
}

func TestCreateVendorData_DatabaseCheckError(t *testing.T) {
	handler, mockDB := setupTestHandler()

	vendorName := "test-vendor"
	requestData := CreateVendorDataRequest{
		VendorName: vendorName,
		Data:       map[string]any{"key1": "value1"},
	}

	mockDB.On("GetVendorData", mock.Anything, vendorName).Return(db.GetVendorDataRow{}, assert.AnError)

	c, w := setupTestContext("POST", "/vendor", requestData)

	handler.CreateVendorData(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to check existing vendor data", response["error"])

	mockDB.AssertExpectations(t)
}

func TestCreateVendorData_DatabaseCreateError(t *testing.T) {
	handler, mockDB := setupTestHandler()

	vendorName := "new-vendor"
	requestData := CreateVendorDataRequest{
		VendorName: vendorName,
		Data:       map[string]any{"key1": "value1"},
	}

	mockDB.On("GetVendorData", mock.Anything, vendorName).Return(db.GetVendorDataRow{}, sql.ErrNoRows)
	mockDB.On("CreateVendorData", mock.Anything, mock.AnythingOfType("db.CreateVendorDataParams")).Return(db.VendorDatum{}, assert.AnError)

	c, w := setupTestContext("POST", "/vendor", requestData)

	handler.CreateVendorData(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to create vendor data", response["error"])

	mockDB.AssertExpectations(t)
}

func TestGetVendorData_Success(t *testing.T) {
	handler, mockDB := setupTestHandler()

	vendorName := "test-vendor"
	testData := map[string]any{
		"key1": "value1",
		"key2": float64(123), // JSON numbers are float64 in Go
	}
	dataBytes, _ := json.Marshal(testData)

	vendorData := db.GetVendorDataRow{
		ID:          1,
		Name:        vendorName,
		Description: nil,
		CreatedAt:   &time.Time{},
		UpdatedAt:   &time.Time{},
		Data:        dataBytes,
	}

	mockDB.On("GetVendorData", mock.Anything, vendorName).Return(vendorData, nil)

	c, w := setupTestContext("GET", "/vendor/"+vendorName, nil)
	c.Params = gin.Params{{Key: "vendor_name", Value: vendorName}}

	handler.GetVendorData(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	data, ok := response["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "value1", data["key1"])
	assert.Equal(t, float64(123), data["key2"])

	mockDB.AssertExpectations(t)
}

func TestGetVendorData_MissingVendorName(t *testing.T) {
	handler, _ := setupTestHandler()

	c, w := setupTestContext("GET", "/vendor/", nil)
	c.Params = gin.Params{{Key: "vendor_name", Value: ""}}

	handler.GetVendorData(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Vendor name is required", response["error"])
}

func TestGetVendorData_VendorNotFound(t *testing.T) {
	handler, mockDB := setupTestHandler()

	vendorName := "nonexistent-vendor"

	mockDB.On("GetVendorData", mock.Anything, vendorName).Return(db.GetVendorDataRow{}, sql.ErrNoRows)

	c, w := setupTestContext("GET", "/vendor/"+vendorName, nil)
	c.Params = gin.Params{{Key: "vendor_name", Value: vendorName}}

	handler.GetVendorData(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Vendor data not found", response["error"])

	mockDB.AssertExpectations(t)
}

func TestGetVendorData_DatabaseError(t *testing.T) {
	handler, mockDB := setupTestHandler()

	vendorName := "test-vendor"

	mockDB.On("GetVendorData", mock.Anything, vendorName).Return(db.GetVendorDataRow{}, assert.AnError)

	c, w := setupTestContext("GET", "/vendor/"+vendorName, nil)
	c.Params = gin.Params{{Key: "vendor_name", Value: vendorName}}

	handler.GetVendorData(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to retrieve vendor data", response["error"])

	mockDB.AssertExpectations(t)
}

func TestGetVendorData_InvalidDataFormat(t *testing.T) {
	handler, mockDB := setupTestHandler()

	vendorName := "test-vendor"

	vendorData := db.GetVendorDataRow{
		ID:   1,
		Name: vendorName,
		Data: []byte("invalid json data"),
	}

	mockDB.On("GetVendorData", mock.Anything, vendorName).Return(vendorData, nil)

	c, w := setupTestContext("GET", "/vendor/"+vendorName, nil)
	c.Params = gin.Params{{Key: "vendor_name", Value: vendorName}}

	handler.GetVendorData(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Failed to parse vendor data", response["error"])

	mockDB.AssertExpectations(t)
}

// Test with nil data
func TestGetVendorData_NilData(t *testing.T) {
	handler, mockDB := setupTestHandler()

	vendorName := "test-vendor"

	vendorData := db.GetVendorDataRow{
		ID:   1,
		Name: vendorName,
		Data: nil,
	}

	mockDB.On("GetVendorData", mock.Anything, vendorName).Return(vendorData, nil)

	c, w := setupTestContext("GET", "/vendor/"+vendorName, nil)
	c.Params = gin.Params{{Key: "vendor_name", Value: vendorName}}

	handler.GetVendorData(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	data, ok := response["data"]
	assert.True(t, ok)
	assert.Nil(t, data)

	mockDB.AssertExpectations(t)
}

// Test edge cases for CreateVendorData with nil data
func TestCreateVendorData_NilData(t *testing.T) {
	handler, mockDB := setupTestHandler()

	vendorName := "new-vendor"
	requestData := CreateVendorDataRequest{
		VendorName: vendorName,
		Data:       nil,
	}

	createdVendorData := db.VendorDatum{
		ID:   1,
		Name: vendorName,
		Data: []byte("null"),
	}

	mockDB.On("GetVendorData", mock.Anything, vendorName).Return(db.GetVendorDataRow{}, sql.ErrNoRows)
	mockDB.On("CreateVendorData", mock.Anything, mock.AnythingOfType("db.CreateVendorDataParams")).Return(createdVendorData, nil)

	c, w := setupTestContext("POST", "/vendor", requestData)

	handler.CreateVendorData(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Vendor data created successfully", response["message"])

	mockDB.AssertExpectations(t)
}
