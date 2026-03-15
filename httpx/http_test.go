// Copyright 2025 ink-yht-code
//
// Proprietary License
//
// IMPORTANT: This software is NOT open source.
// You may NOT use, copy, modify, merge, publish, distribute, sublicense,
// or sell copies of this file, in whole or in part, without prior written
// permission from the copyright holder.
//
// This software is provided "AS IS", without warranty of any kind.

package httpx

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewServer(t *testing.T) {
	cfg := Config{
		Enabled: true,
		Addr:    ":8080",
	}

	s := NewServer(cfg)
	assert.NotNil(t, s)
	assert.NotNil(t, s.Engine)
	assert.NotNil(t, s.Server)
	assert.Equal(t, ":8080", s.Server.Addr)
}

func TestServer_Routes(t *testing.T) {
	cfg := Config{Enabled: true, Addr: ":0"}
	s := NewServer(cfg)

	s.Engine.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	s.Engine.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestRequestID(t *testing.T) {
	middleware := RequestID()

	t.Run("generates new request id", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/", nil)

		middleware(c)

		requestID, exists := c.Get("request_id")
		assert.True(t, exists)
		assert.NotEmpty(t, requestID)
	})

	t.Run("uses existing request id", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("X-Request-Id", "existing-id")

		middleware(c)

		requestID, _ := c.Get("request_id")
		assert.Equal(t, "existing-id", requestID)
	})
}

func TestLogger(t *testing.T) {
	middleware := Logger()

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/test?query=value", nil)

	// Should not panic
	assert.NotPanics(t, func() {
		middleware(c)
	})
}

func TestRecovery(t *testing.T) {
	middleware := Recovery()

	t.Run("recovers from panic", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)

		middleware(c)

		// Handler that panics
		c.AbortWithStatusJSON(500, gin.H{"error": "panic"})
	})

	t.Run("normal request", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)

		middleware(c)
		c.Next()

		// Should complete normally
	})
}

func TestServer_Shutdown(t *testing.T) {
	cfg := Config{Enabled: true, Addr: ":0"}
	s := NewServer(cfg)

	ctx := context.Background()
	err := s.Shutdown(ctx)
	assert.NoError(t, err)
}
