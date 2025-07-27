package internal_routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

// Benchmark tests for vendor data operations

func BenchmarkCreateVendorData(b *testing.B) {
	handler, mockDB := setupTestHandler()

	vendorName := "benchmark-vendor"
	createReq := CreateVendorDataRequest{
		VendorName: vendorName,
		Data: map[string]any{
			"benchmark": true,
			"iteration": 0,
		},
	}

	createdData := db.VendorDatum{
		ID:   1,
		Name: vendorName,
	}

	// Setup expectations once - they'll be used for all iterations
	mockDB.On("GetVendorData", mock.Anything, vendorName).Return(db.GetVendorDataRow{}, sql.ErrNoRows)
	mockDB.On("CreateVendorData", mock.Anything, mock.AnythingOfType("db.CreateVendorDataParams")).Return(createdData, nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		createReq.Data["iteration"] = i

		c, _ := setupTestContext("POST", "/vendor", createReq)
		handler.CreateVendorData(c)
	}
}

func BenchmarkGetVendorData(b *testing.B) {
	handler, mockDB := setupTestHandler()

	vendorName := "benchmark-vendor"
	testData := map[string]any{
		"benchmark": true,
		"data_size": "medium",
		"nested": map[string]any{
			"level1": map[string]any{
				"level2": "value",
			},
		},
	}
	dataBytes, _ := json.Marshal(testData)

	vendorData := db.GetVendorDataRow{
		ID:   1,
		Name: vendorName,
		Data: dataBytes,
	}

	mockDB.On("GetVendorData", mock.Anything, vendorName).Return(vendorData, nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c, _ := setupTestContext("GET", "/vendor/"+vendorName, nil)
		c.Params = gin.Params{{Key: "vendor_name", Value: vendorName}}
		handler.GetVendorData(c)
	}
}

func BenchmarkUpdateVendorData(b *testing.B) {
	handler, mockDB := setupTestHandler()

	vendorName := "benchmark-vendor"
	updateReq := AddVendorDataKeyRequest{
		Data: map[string]any{
			"benchmark":  true,
			"updated_at": "2024-01-01T00:00:00Z",
			"iteration":  0,
		},
	}

	existingVendorData := db.GetVendorDataRow{
		ID:   1,
		Name: vendorName,
		Data: []byte(`{"existing": "data"}`),
	}

	updatedVendorData := db.VendorDatum{
		ID:   1,
		Name: vendorName,
	}

	mockDB.On("GetVendorData", mock.Anything, vendorName).Return(existingVendorData, nil)
	mockDB.On("UpdateVendorData", mock.Anything, mock.AnythingOfType("db.UpdateVendorDataParams")).Return(updatedVendorData, nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		updateReq.Data["iteration"] = i

		c, _ := setupTestContext("PUT", "/vendor/"+vendorName, updateReq)
		c.Params = gin.Params{{Key: "vendor_name", Value: vendorName}}
		handler.UpdateVendorData(c)
	}
}

// Benchmark with large JSON payloads
func BenchmarkCreateVendorDataLargePayload(b *testing.B) {
	handler, mockDB := setupTestHandler()

	vendorName := "benchmark-large-vendor"

	// Create a large data structure
	largeData := make(map[string]any)
	for i := 0; i < 1000; i++ {
		largeData[fmt.Sprintf("key_%d", i)] = map[string]any{
			"value":       fmt.Sprintf("value_%d", i),
			"metadata":    map[string]any{"type": "string", "required": true},
			"nested_data": []any{1, 2, 3, 4, 5},
		}
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

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c, _ := setupTestContext("POST", "/vendor", createReq)
		handler.CreateVendorData(c)
	}
}

func BenchmarkJSONSerialization(b *testing.B) {
	testData := map[string]any{
		"simple_string": "value",
		"number":        123,
		"boolean":       true,
		"nested": map[string]any{
			"level1": map[string]any{
				"level2": "deep_value",
			},
		},
		"array": []any{1, "two", 3.0, true},
	}

	b.ResetTimer()

	b.Run("ToBytes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := db.ToBytes(testData)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	dataBytes, _ := db.ToBytes(testData)

	b.Run("ToJSONB", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var result map[string]any
			err := db.ToJSONB(dataBytes, &result)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
