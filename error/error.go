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

// Package error 提供错误映射功能
package error

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// BizError 业务错误接口
type BizError interface {
	BizCode() int
	BizMsg() string
	Error() string
}

// MapToHTTP 将错误映射到 HTTP 响应
func MapToHTTP(c *gin.Context, err error) {
	if err == nil {
		return
	}

	var biz BizError
	if errors.As(err, &biz) {
		status, resp := mapBizError(biz)
		c.JSON(status, resp)
		return
	}

	// 非业务错误
	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    0,
		"message": "internal error",
	})
}

// mapBizError 映射业务错误
func mapBizError(biz BizError) (int, gin.H) {
	bizCode := biz.BizCode()
	suffix := bizCode % 10000

	var status int
	switch suffix {
	case 1: // InvalidParam
		status = http.StatusBadRequest
	case 2: // Unauthorized
		status = http.StatusUnauthorized
	case 3: // Forbidden
		status = http.StatusForbidden
	case 4: // NotFound
		status = http.StatusNotFound
	case 5: // Conflict
		status = http.StatusConflict
	default:
		status = http.StatusInternalServerError
	}

	return status, gin.H{
		"code":    bizCode,
		"message": biz.BizMsg(),
	}
}

// Handler 错误处理中间件
func Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			err := c.Errors[0].Err
			MapToHTTP(c, err)
		}
	}
}
