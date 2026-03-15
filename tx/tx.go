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

// Package tx 提供事务管理，支持 ctx 注入事务 DB
package tx

import (
	"context"

	"gorm.io/gorm"
)

type ctxKey struct{}

// Manager 事务管理器
type Manager struct {
	db *gorm.DB
}

// NewManager 创建事务管理器
func NewManager(db *gorm.DB) *Manager {
	return &Manager{db: db}
}

// Do 在事务中执行函数
func (m *Manager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 将事务 DB 注入 ctx
		ctx = context.WithValue(ctx, ctxKey{}, tx)
		return fn(ctx)
	})
}

// FromContext 从 ctx 获取 DB（优先返回事务 DB）
func FromContext(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if ctx == nil {
		return defaultDB
	}
	if tx, ok := ctx.Value(ctxKey{}).(*gorm.DB); ok {
		return tx
	}
	return defaultDB
}

// GetDB 获取 DB（用于 dao 层）
func GetDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	return FromContext(ctx, defaultDB)
}
