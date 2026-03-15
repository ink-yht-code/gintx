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

package rpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		cfg := Config{Enabled: false, Addr: ":9090"}
		s := NewServer(cfg)
		assert.Nil(t, s)
	})

	t.Run("enabled", func(t *testing.T) {
		cfg := Config{Enabled: true, Addr: ":0"}
		s := NewServer(cfg)
		assert.NotNil(t, s)
		assert.NotNil(t, s.Server)
	})
}

func TestServer_Shutdown(t *testing.T) {
	cfg := Config{Enabled: true, Addr: ":0"}
	s := NewServer(cfg)

	ctx := context.Background()
	err := s.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestUnaryInterceptor(t *testing.T) {
	// Test that the interceptor function exists
	assert.NotNil(t, unaryInterceptor)
}

func TestStreamInterceptor(t *testing.T) {
	// Test that the interceptor function exists
	assert.NotNil(t, streamInterceptor)
}
