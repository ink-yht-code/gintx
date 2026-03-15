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

// Package redis 提供 Redis 客户端初始化
package redis

import (
	"time"

	"github.com/redis/go-redis/v9"
)

// Config Redis 配置
type Config struct {
	Addr     string
	Password string
	DB       int
}

// New 创建 Redis 客户端
func New(cfg Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
}
