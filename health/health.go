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

// Package health 提供健康检查功能
package health

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// HTTPHandler HTTP 健康检查处理器
func HTTPHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"name":        "gint",
			"description": "基于 Gin 构建的 Go 微服务框架，提供路由、认证、校验等开箱即用的功能",
			"version":     "1.0.0",
		})
	}
}

// ReadyHandler 就绪检查处理器
func ReadyHandler(checks ...func() bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, check := range checks {
			if !check() {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"status": "not ready",
				})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "ready",
		})
	}
}

// Checker 健康检查器接口
type Checker interface {
	Check(ctx context.Context) error
}

// HealthServer gRPC 健康检查服务
type HealthServer struct {
	grpc_health_v1.UnimplementedHealthServer
	checkers map[string]Checker
}

// NewHealthServer 创建健康检查服务
func NewHealthServer() *HealthServer {
	return &HealthServer{
		checkers: make(map[string]Checker),
	}
}

// Register 注册检查器
func (s *HealthServer) Register(service string, checker Checker) {
	s.checkers[service] = checker
}

// Check 实现 gRPC 健康检查
func (s *HealthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	if req.Service == "" {
		// 整体健康检查
		return &grpc_health_v1.HealthCheckResponse{
			Status: grpc_health_v1.HealthCheckResponse_SERVING,
		}, nil
	}

	checker, ok := s.checkers[req.Service]
	if !ok {
		return &grpc_health_v1.HealthCheckResponse{
			Status: grpc_health_v1.HealthCheckResponse_UNKNOWN,
		}, nil
	}

	if err := checker.Check(ctx); err != nil {
		return &grpc_health_v1.HealthCheckResponse{
			Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
		}, nil
	}

	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

// Watch 实现 gRPC 健康检查（流式）
func (s *HealthServer) Watch(req *grpc_health_v1.HealthCheckRequest, stream grpc_health_v1.Health_WatchServer) error {
	// 简单实现：只返回一次状态
	resp, err := s.Check(stream.Context(), req)
	if err != nil {
		return err
	}
	return stream.Send(resp)
}

// DBChecker 数据库检查器
type DBChecker struct {
	ping func() error
}

// NewDBChecker 创建数据库检查器
func NewDBChecker(ping func() error) *DBChecker {
	return &DBChecker{ping: ping}
}

// Check 检查数据库连接
func (c *DBChecker) Check(ctx context.Context) error {
	return c.ping()
}

// RedisChecker Redis 检查器
type RedisChecker struct {
	ping func() error
}

// NewRedisChecker 创建 Redis 检查器
func NewRedisChecker(ping func() error) *RedisChecker {
	return &RedisChecker{ping: ping}
}

// Check 检查 Redis 连接
func (c *RedisChecker) Check(ctx context.Context) error {
	return c.ping()
}
