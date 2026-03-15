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

package tx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestNewManager(t *testing.T) {
	// Test that NewManager returns a non-nil Manager
	m := NewManager(nil)
	assert.NotNil(t, m)
}

func TestFromContext(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		expectNil bool
	}{
		{
			name:      "nil context returns default",
			ctx:       nil,
			expectNil: true,
		},
		{
			name:      "empty context returns default",
			ctx:       context.Background(),
			expectNil: true,
		},
		{
			name:      "context with tx value",
			ctx:       context.WithValue(context.Background(), ctxKey{}, &gorm.DB{}),
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromContext(tt.ctx, nil)
			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestGetDB(t *testing.T) {
	t.Run("returns from context when available", func(t *testing.T) {
		tx := &gorm.DB{}
		ctx := context.WithValue(context.Background(), ctxKey{}, tx)
		result := GetDB(ctx, nil)
		assert.Equal(t, tx, result)
	})

	t.Run("returns default when not in context", func(t *testing.T) {
		ctx := context.Background()
		result := GetDB(ctx, nil)
		assert.Nil(t, result)
	})
}

func TestCtxKey(t *testing.T) {
	// Verify ctxKey is a valid context key type
	key := ctxKey{}
	_ = key
}
