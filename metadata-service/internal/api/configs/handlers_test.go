package configs

import (
	"net/http"
	"net/http/httptest"

	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/config"
	"github.com/focadecombate/incus-metadata-service/metadata-service/internal/storage/db/mocks"
	"github.com/gin-gonic/gin"
)

// setupConfigsTest builds a Handler backed by a mock Querier.
func setupConfigsTest() (*Handler, *mocks.MockQuerier) {
	mockDB := &mocks.MockQuerier{}
	handler := &Handler{
		Config:   &config.Config{},
		Database: mockDB,
	}
	return handler, mockDB
}

// newGetContext creates a GET gin context with the given Accept header. An
// empty accept string leaves the header unset (mirrors a client that sends no
// Accept). remoteAddr, when non-empty, overrides the default RemoteAddr so
// ClientIP resolution can be exercised.
func newGetContext(accept, remoteAddr, xForwardedFor string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if accept != "" {
		req.Header.Set("Accept", accept)
	}
	if remoteAddr != "" {
		req.RemoteAddr = remoteAddr
	}
	if xForwardedFor != "" {
		req.Header.Set("X-Forwarded-For", xForwardedFor)
	}
	c.Request = req
	return c, w
}
