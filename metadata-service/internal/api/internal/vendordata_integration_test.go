package internal_routes

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Integration-style tests that test the handlers with actual router setup

func setupRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Setup routes similar to the actual API
	api := router.Group("/api/v1")
	{
		vendor := api.Group("/vendor")
		{
			vendor.POST("", handler.CreateVendorData)
			vendor.GET("/:vendor_name", handler.GetVendorData)
			vendor.PUT("/:vendor_name", handler.UpdateVendorData)
		}
	}
	return router
}

func TestVendorDataRoutes_Integration(t *testing.T) {
	handler, mockDB := setupTestHandler()
	router := setupRouter(handler)

	t.Run("Complete workflow - Create, Get, Update", func(t *testing.T) {
		vendorName := "integration-vendor"
		description := "Integration test vendor"
		
		// Test Create
		createReq := CreateVendorDataRequest{
			VendorName:  vendorName,
			Description: &description,
			Data: map[string]any{
				"version": "1.0.0",
				"features": []string{"feature1", "feature2"},
			},
		}
		
		createdData := db.VendorDatum{
			ID:          1,
			Name:        vendorName,
			Description: &description,
		}
		
		mockDB.On("GetVendorData", mock.Anything, vendorName).Return(db.GetVendorDataRow{}, sql.ErrNoRows).Once()
		mockDB.On("CreateVendorData", mock.Anything, mock.AnythingOfType("db.CreateVendorDataParams")).Return(createdData, nil).Once()
		
		reqBody, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/v1/vendor", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
		
		// Test Get
		getData := map[string]any{
			"version": "1.0.0",
			"features": []any{"feature1", "feature2"},
		}
		dataBytes, _ := json.Marshal(getData)
		
		getVendorData := db.GetVendorDataRow{
			ID:   1,
			Name: vendorName,
			Data: dataBytes,
		}
		
		mockDB.On("GetVendorData", mock.Anything, vendorName).Return(getVendorData, nil).Once()
		
		req = httptest.NewRequest("GET", "/api/v1/vendor/"+vendorName, nil)
		w = httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		
		var getResponse map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &getResponse)
		assert.NoError(t, err)
		assert.Contains(t, getResponse, "data")
		
		// Test Update
		updateReq := AddVendorDataKeyRequest{
			Data: map[string]any{
				"version": "2.0.0",
				"features": []string{"feature1", "feature2", "feature3"},
				"new_field": "new_value",
			},
		}
		
		mockDB.On("GetVendorData", mock.Anything, vendorName).Return(getVendorData, nil).Once()
		mockDB.On("UpdateVendorData", mock.Anything, mock.AnythingOfType("db.UpdateVendorDataParams")).Return(createdData, nil).Once()
		
		reqBody, _ = json.Marshal(updateReq)
		req = httptest.NewRequest("PUT", "/api/v1/vendor/"+vendorName, bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		
		mockDB.AssertExpectations(t)
	})
}

func TestVendorDataValidation_EdgeCases(t *testing.T) {
	t.Run("Create with empty vendor name", func(t *testing.T) {
		handler, _ := setupTestHandler()
		router := setupRouter(handler)
		
		createReq := CreateVendorDataRequest{
			VendorName: "",
			Data:       map[string]any{"key": "value"},
		}
		
		reqBody, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/v1/vendor", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("Create with missing data field (valid case)", func(t *testing.T) {
		handler, mockDB := setupTestHandler()
		router := setupRouter(handler)
		
		// Data field is actually optional for CreateVendorDataRequest, so this should succeed
		vendorName := "test-vendor"
		createReq := map[string]any{
			"vendor_name": vendorName,
			// Missing "data" field - this is valid for create requests
		}
		
		createdData := db.VendorDatum{
			ID:   1,
			Name: vendorName,
		}
		
		mockDB.On("GetVendorData", mock.Anything, vendorName).Return(db.GetVendorDataRow{}, sql.ErrNoRows)
		mockDB.On("CreateVendorData", mock.Anything, mock.AnythingOfType("db.CreateVendorDataParams")).Return(createdData, nil)
		
		reqBody, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/v1/vendor", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Vendor data created successfully", response["message"])
		
		mockDB.AssertExpectations(t)
	})
	
	t.Run("Update with missing required data field", func(t *testing.T) {
		handler, _ := setupTestHandler()
		router := setupRouter(handler)
		
		// This should fail validation because the binding:"required" tag on Data field
		updateReq := map[string]any{
			// Missing "data" field
		}
		
		reqBody, _ := json.Marshal(updateReq)
		req := httptest.NewRequest("PUT", "/api/v1/vendor/test-vendor", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	
	t.Run("Create with complex nested JSON data", func(t *testing.T) {
		handler, mockDB := setupTestHandler()
		router := setupRouter(handler)
		
		vendorName := "complex-vendor"
		createReq := CreateVendorDataRequest{
			VendorName: vendorName,
			Data: map[string]any{
				"config": map[string]any{
					"database": map[string]any{
						"host":     "localhost",
						"port":     5432,
						"ssl_mode": "require",
					},
					"features": []any{
						map[string]any{"name": "feature1", "enabled": true},
						map[string]any{"name": "feature2", "enabled": false},
					},
				},
				"metadata": map[string]any{
					"version":     "1.0.0",
					"created_by":  "integration_test",
					"tags":        []string{"test", "integration", "complex"},
					"numbers":     []int{1, 2, 3, 4, 5},
					"mixed_array": []any{"string", 123, true, nil},
				},
			},
		}
		
		createdData := db.VendorDatum{
			ID:   1,
			Name: vendorName,
		}
		
		mockDB.On("GetVendorData", mock.Anything, vendorName).Return(db.GetVendorDataRow{}, sql.ErrNoRows)
		mockDB.On("CreateVendorData", mock.Anything, mock.AnythingOfType("db.CreateVendorDataParams")).Return(createdData, nil)
		
		reqBody, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/v1/vendor", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
		
		mockDB.AssertExpectations(t)
	})
	
	t.Run("GET with special characters in vendor name", func(t *testing.T) {
		handler, mockDB := setupTestHandler()
		router := setupRouter(handler)
		
		vendorName := "vendor-with-special-chars_123"
		
		getData := map[string]any{"key": "value"}
		dataBytes, _ := json.Marshal(getData)
		
		getVendorData := db.GetVendorDataRow{
			ID:   1,
			Name: vendorName,
			Data: dataBytes,
		}
		
		mockDB.On("GetVendorData", mock.Anything, vendorName).Return(getVendorData, nil)
		
		req := httptest.NewRequest("GET", "/api/v1/vendor/"+vendorName, nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		
		mockDB.AssertExpectations(t)
	})
}

func TestVendorDataErrorHandling(t *testing.T) {
	t.Run("Handle database connection issues", func(t *testing.T) {
		handler, mockDB := setupTestHandler()
		router := setupRouter(handler)
		
		vendorName := "test-vendor"
		
		// Simulate database connection error
		mockDB.On("GetVendorData", mock.Anything, vendorName).Return(db.GetVendorDataRow{}, assert.AnError)
		
		req := httptest.NewRequest("GET", "/api/v1/vendor/"+vendorName, nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to retrieve vendor data", response["error"])
		
		mockDB.AssertExpectations(t)
	})
	
	t.Run("Handle malformed JSON in request", func(t *testing.T) {
		handler, _ := setupTestHandler()
		router := setupRouter(handler)
		
		req := httptest.NewRequest("POST", "/api/v1/vendor", bytes.NewBufferString("{invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid request payload", response["error"])
	})
	
	t.Run("Handle large JSON payload", func(t *testing.T) {
		handler, mockDB := setupTestHandler()
		router := setupRouter(handler)
		
		vendorName := "large-data-vendor"
		
		// Create a large data structure
		largeData := make(map[string]any)
		for i := 0; i < 1000; i++ {
			largeData[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d_with_some_longer_content_to_make_it_bigger", i)
		}
		
		createReq := CreateVendorDataRequest{
			VendorName: vendorName,
			Data:       largeData,
		}
		
		createdData := db.VendorDatum{
			ID:   1,
			Name: vendorName,
		}
		
		mockDB.On("GetVendorData", mock.Anything, vendorName).Return(db.GetVendorDataRow{}, sql.ErrNoRows)
		mockDB.On("CreateVendorData", mock.Anything, mock.AnythingOfType("db.CreateVendorDataParams")).Return(createdData, nil)
		
		reqBody, _ := json.Marshal(createReq)
		req := httptest.NewRequest("POST", "/api/v1/vendor", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
		
		mockDB.AssertExpectations(t)
	})
}
