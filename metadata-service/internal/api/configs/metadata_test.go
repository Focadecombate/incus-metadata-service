package configs

import (
	"database/sql"
	"net/http"
	"strings"
	"testing"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAllMetadataHandler(t *testing.T) {
	metadataJSON := []byte(`{"instance-id":"i-abc123","local-hostname":"web-01"}`)

	tests := []struct {
		name       string
		accept     string
		row        db.InstanceMetadatum
		rowErr     error
		wantStatus int
		wantBody   []string // substrings expected in the body
	}{
		{
			name:       "known instance, Accept */* -> 200 YAML with instance-id",
			accept:     "*/*",
			row:        db.InstanceMetadatum{Metadata: metadataJSON},
			wantStatus: http.StatusOK,
			wantBody:   []string{"instance-id: i-abc123", "local-hostname: web-01"},
		},
		{
			name:       "known instance, missing Accept -> 200 YAML",
			accept:     "",
			row:        db.InstanceMetadatum{Metadata: metadataJSON},
			wantStatus: http.StatusOK,
			wantBody:   []string{"instance-id: i-abc123"},
		},
		{
			name:       "known instance, compound Accept with */* -> 200 YAML (no 406)",
			accept:     "text/html,application/xhtml+xml,*/*;q=0.8",
			row:        db.InstanceMetadatum{Metadata: metadataJSON},
			wantStatus: http.StatusOK,
			wantBody:   []string{"instance-id: i-abc123"},
		},
		{
			name:       "unknown IP -> 404",
			accept:     "*/*",
			rowErr:     sql.ErrNoRows,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, mockDB := setupConfigsTest()
			mockDB.On("GetInstanceMetadataByIP", mock.Anything, mock.Anything).
				Return(tt.row, tt.rowErr)

			c, w := newGetContext(tt.accept, "", "")
			handler.AllMetadataHandler(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			for _, want := range tt.wantBody {
				assert.Contains(t, w.Body.String(), want)
			}
			mockDB.AssertExpectations(t)
		})
	}
}

func TestAllMetadataHandler_ExplicitJSON(t *testing.T) {
	handler, mockDB := setupConfigsTest()
	mockDB.On("GetInstanceMetadataByIP", mock.Anything, mock.Anything).
		Return(db.InstanceMetadatum{Metadata: []byte(`{"instance-id":"i-json"}`)}, nil)

	c, w := newGetContext("application/json", "", "")
	handler.AllMetadataHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"instance-id":"i-json"`)
	mockDB.AssertExpectations(t)
}

// TestAllMetadataHandler_ReadsClientIP documents that the handler resolves the
// instance from the request ClientIP. The trusted-proxy wiring that decides
// whether X-Forwarded-For is honoured is fixed by another agent; here we only
// assert the handler passes a non-empty ClientIP to the datastore.
func TestAllMetadataHandler_ReadsClientIP(t *testing.T) {
	handler, mockDB := setupConfigsTest()

	var capturedIP string
	mockDB.On("GetInstanceMetadataByIP", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			if p, ok := args.Get(1).(*string); ok && p != nil {
				capturedIP = *p
			}
		}).
		Return(db.InstanceMetadatum{Metadata: []byte(`{"instance-id":"i-ip"}`)}, nil)

	c, w := newGetContext("*/*", "198.51.100.7:5555", "203.0.113.9")
	handler.AllMetadataHandler(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, capturedIP, "handler should query the datastore with the resolved ClientIP")
	// ClientIP is either the proxied XFF value or the direct RemoteAddr host,
	// depending on the (separately-fixed) trusted-proxy config. Both are valid
	// here; we only require that the handler used one of them.
	assert.True(t,
		strings.Contains(capturedIP, "203.0.113.9") || strings.Contains(capturedIP, "198.51.100.7"),
		"unexpected client IP: %s", capturedIP)
	mockDB.AssertExpectations(t)
}
