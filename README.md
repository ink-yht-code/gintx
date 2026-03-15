# gintx - 运行时框架

[![License](https://img.shields.io/badge/License-Proprietary-red.svg)](#license)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://golang.org)
[![Version](https://img.shields.io/badge/version-v1.0.0-blue.svg)](https://github.com/ink-yht-code/gintx/releases)

gintx 提供微服务运行时基础设施组件，包括日志、数据库、Redis、HTTP/gRPC 服务器、事务管理、健康检查等，帮助开发者快速构建生产级的微服务。

## 目录

- [特性](#特性)
- [快速开始](#快速开始)
- [核心模块](#核心模块)
  - [log - 结构化日志](#log---结构化日志)
  - [tx - 事务管理](#tx---事务管理)
  - [db - 数据库初始化](#db---数据库初始化)
  - [redis - Redis初始化](#redis---redis初始化)
  - [httpx - HTTP服务器](#httpx---http服务器)
  - [rpc - gRPC服务器](#rpc---grpc服务器)
  - [health - 健康检查](#health---健康检查)
  - [error - 错误映射](#error---错误映射)
  - [app - 应用启动器](#app---应用启动器)
- [错误码规范](#错误码规范)
- [完整示例](#完整示例)
- [贡献](#贡献)
- [许可证](#许可证)

## 特性

- **结构化日志** - 基于 zap，支持 ctx 注入 request_id，便于链路追踪
- **事务管理** - 基于 ctx 的事务传递，简化事务处理，DAO 层无感知
- **数据库初始化** - GORM 初始化，支持连接池配置和日志级别设置
- **Redis 初始化** - Redis 客户端初始化，基于 go-redis/v9
- **HTTP 服务器** - Gin 服务器，内置常用中间件，支持优雅关闭
- **gRPC 服务器** - gRPC 服务器，内置拦截器，支持优雅关闭
- **健康检查** - HTTP 和 gRPC 健康检查，支持依赖检查
- **优雅关闭** - 支持信号监听和超时关闭
- **应用启动器** - 统一初始化所有组件，简化应用启动流程

## 快速开始

### 安装

```bash
go get github.com/ink-yht-code/gintx
```

### 基本使用

```go
package main

import (
    "github.com/ink-yht-code/gintx/app"
    "github.com/ink-yht-code/gintx/httpx"
    "github.com/ink-yht-code/gintx/log"
)

func main() {
    // 创建应用
    application, err := app.New(&app.Config{
        Service: app.ServiceConfig{ID: 101, Name: "user"},
        HTTP:    httpx.Config{Enabled: true, Addr: ":8080"},
        Log:     log.Config{Level: "info", Encoding: "json"},
    })
    if err != nil {
        log.Fatal("failed to create app", zap.Error(err))
    }
    
    // 启动
    go application.Run()
    
    // 等待信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    // 优雅关闭
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    application.Shutdown(ctx)
}
```

## 核心模块

### log - 结构化日志

基于 zap 的结构化日志，支持 ctx 注入 request_id，便于链路追踪。

#### 配置

```yaml
log:
  level: "info"      # 日志级别: debug|info|warn|error
  encoding: "json"   # 输出格式: json|console
  output: "stdout"   # 输出位置: stdout|file
```

#### 使用

```go
import "github.com/ink-yht-code/gintx/log"

// 初始化
log.Init(log.Config{
    Level:    "info",
    Encoding: "json",
    Output:   "stdout",
})

// 普通日志
log.Info("service started", zap.String("name", "user-service"))
log.Error("database error", zap.Error(err))

// 带 ctx 的日志（自动注入 request_id）
func Handler(ctx *gin.Context) {
    log.Ctx(ctx).Info("handling request", zap.String("path", ctx.Request.URL.Path))
}
```

#### 在中间件中自动注入 request_id

```go
// httpx 服务器已内置 RequestID 中间件
server := httpx.NewServer(httpx.Config{Addr: ":8080"})
// 所有请求都会自动生成 request_id 并注入到 ctx
```

### tx - 事务管理

事务管理器，支持基于 ctx 的事务传递，让 DAO 层无感知事务。

#### 核心概念

```
Service 层                    DAO 层
    |                            |
    v                            |
txMgr.Do(ctx, func(ctx) {       |
    // ctx 中注入了事务 DB        |
    repo.Save(ctx, user) ------>|
                                v
                           db := tx.GetDB(ctx, d.db)
                           // 自动使用事务 DB
                           db.Create(user)
})
```

#### 使用

```go
import "github.com/ink-yht-code/gintx/tx"

// 创建事务管理器
txMgr := tx.NewManager(db)

// 在 Service 层开启事务
func (s *UserService) Register(ctx context.Context, req *RegisterReq) error {
    return s.txMgr.Do(ctx, func(ctx context.Context) error {
        // 创建用户
        if err := s.userRepo.Save(ctx, &user); err != nil {
            return err // 自动回滚
        }
        
        // 创建用户资料
        if err := s.profileRepo.Save(ctx, &profile); err != nil {
            return err // 自动回滚
        }
        
        return nil // 自动提交
    })
}

// 在 DAO 层获取 DB
func (d *UserDAO) Save(ctx context.Context, user *entity.User) error {
    // 自动使用事务 DB（如果在事务中）
    db := tx.GetDB(ctx, d.db)
    return db.Create(user).Error
}
```

#### 事务传播

```go
// 嵌套事务会复用外层事务
func (s *UserService) ComplexOperation(ctx context.Context) error {
    return s.txMgr.Do(ctx, func(ctx context.Context) error {
        // 这个 ctx 已经包含事务 DB
        
        s.otherService.Operation(ctx) // 会复用同一个事务
        
        return nil
    })
}
```

### db - 数据库初始化

GORM 数据库初始化，支持连接池配置和日志级别设置。

#### 配置

```yaml
db:
  dsn: "user:pass@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=True&loc=Local"
  max_open: 100      # 最大连接数
  max_idle: 10       # 最大空闲连接数
  log_level: "info"  # gorm 日志级别: silent|error|warn|info
```

#### 使用

```go
import "github.com/ink-yht-code/gintx/db"

db, err := db.New(db.Config{
    DSN:      "user:pass@tcp(127.0.0.1:3306)/mydb?charset=utf8mb4&parseTime=True&loc=Local",
    MaxOpen:  100,
    MaxIdle:  10,
    LogLevel: "info",
})
if err != nil {
    log.Fatal("failed to connect database", zap.Error(err))
}

// 自动迁移
db.AutoMigrate(&User{}, &Profile{})
```

### redis - Redis 初始化

Redis 客户端初始化，基于 go-redis/v9。

#### 配置

```yaml
redis:
  addr: "127.0.0.1:6379"
  password: ""
  db: 0
```

#### 使用

```go
import "github.com/ink-yht-code/gintx/redis"

client := redis.New(redis.Config{
    Addr:     "127.0.0.1:6379",
    Password: "",
    DB:       0,
})

// 基本操作
client.Set(ctx, "key", "value", time.Hour)
val, err := client.Get(ctx, "key").Result()

// Session 存储
client.Set(ctx, "session:"+sessionId, sessionData, time.Hour*2)
```

### httpx - HTTP 服务器

Gin HTTP 服务器，内置常用中间件，支持优雅关闭。

#### 配置

```yaml
http:
  enabled: true
  addr: ":8080"
```

#### 使用

```go
import "github.com/ink-yht-code/gintx/httpx"

server := httpx.NewServer(httpx.Config{
    Enabled: true,
    Addr:    ":8080",
})

// 注册路由
server.Engine.GET("/api/hello", handler)
server.Engine.POST("/api/users", createUserHandler)

// 启动（阻塞）
server.Run()

// 或异步启动
go server.Run()

// 优雅关闭
server.Shutdown(ctx)
```

#### 内置中间件

| 中间件        | 说明                  |
| ------------- | --------------------- |
| `RequestID()` | 为每个请求生成唯一 ID |
| `Logger()`    | 记录访问日志          |
| `Recovery()`  | Panic 恢复，返回 500  |
| `CORS()`      | 跨域处理（可选）      |

#### 添加自定义中间件

```go
server.Engine.Use(cors.Default())
server.Engine.Use(rateLimiterMiddleware)
```

### rpc - gRPC 服务器

gRPC 服务器，内置拦截器，支持优雅关闭。

#### 配置

```yaml
grpc:
  enabled: true
  addr: ":9090"
```

#### 使用

```go
import (
    "github.com/ink-yht-code/gintx/rpc"
    pb "your-project/api/gen/go"
)

server := rpc.NewServer(rpc.Config{
    Enabled: true,
    Addr:    ":9090",
})

// 注册服务
pb.RegisterUserServiceServer(server.Server, &userService{})

// 启动
server.Run()

// 关闭
server.Shutdown(ctx)
```

#### 内置拦截器

| 拦截器                | 说明               |
| --------------------- | ------------------ |
| `LoggingInterceptor`  | 记录请求日志       |
| `RecoveryInterceptor` | Panic 恢复         |
| `AuthInterceptor`     | 认证拦截器（可选） |

### health - 健康检查

HTTP 和 gRPC 健康检查，支持依赖检查。

#### HTTP 健康检查

```go
import "github.com/ink-yht-code/gintx/health"

// 框架介绍（首页）
server.Engine.GET("/", health.HTTPHandler())

// 存活检查（K8s liveness）
server.Engine.GET("/health", health.LiveHandler())

// 就绪检查（K8s readiness）
server.Engine.GET("/ready", health.ReadyHandler(
    health.CheckDB(db),
    health.CheckRedis(redisClient),
))
```

#### gRPC 健康检查

```go
import (
    "github.com/ink-yht-code/gintx/health"
    grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
)

// 创建健康检查服务器
healthServer := health.NewGRPCHealthServer()

// 注册检查器
healthServer.Register("user-service", health.CheckDB(db))

// 注册到 gRPC 服务器
grpc_health_v1.RegisterHealthServer(server.Server, healthServer)
```

### error - 错误映射

业务错误到 HTTP/gRPC 响应映射。

#### 实现 BizError 接口

```go
import "github.com/ink-yht-code/gintx/error"

// 业务错误
type BizError struct {
    code int
    msg  string
}

func (e *BizError) BizCode() int   { return e.code }
func (e *BizError) BizMsg() string { return e.msg }
func (e *BizError) Error() string  { return e.msg }

// 使用
func (s *Service) GetUser(id int64) (*User, error) {
    user, err := s.repo.Find(id)
    if err != nil {
        return nil, &BizError{code: 4, msg: "用户不存在"}
    }
    return user, nil
}
```

#### 在 Handler 中映射

```go
func (h *Handler) GetUser(ctx *gctx.Context) (gint.Result, error) {
    user, err := h.svc.GetUser(id)
    if err != nil {
        // 自动映射业务错误码
        return gint.Result{}, err
    }
    return gint.Result{Code: 0, Data: user}, nil
}
```

### app - 应用启动器

应用生命周期管理，统一初始化所有组件。

#### 使用

```go
import "github.com/ink-yht-code/gintx/app"

func main() {
    // 创建应用
    application, err := app.New(&app.Config{
        Service: app.ServiceConfig{ID: 101, Name: "user"},
        HTTP:    httpx.Config{Enabled: true, Addr: ":8080"},
        GRPC:    rpc.Config{Enabled: false},
        Log:     log.Config{Level: "info", Encoding: "json"},
        DB:      db.Config{DSN: "..."},
        Redis:   redis.Config{Addr: "..."},
    })
    if err != nil {
        log.Fatal("failed to create app", zap.Error(err))
    }
    
    // 启动
    go application.Run()
    
    // 等待信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    // 优雅关闭
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    application.Shutdown(ctx)
}
```

## 错误码规范

业务码 = ServiceID × 10000 + BizCode

| BizCode | 含义     |
| ------- | -------- |
| 0       | 成功     |
| 1       | 参数错误 |
| 2       | 未授权   |
| 3       | 无权限   |
| 4       | 未找到   |
| 5       | 冲突     |
| 9999    | 内部错误 |

例如 user 服务（ServiceID=101）：
- 参数错误：1010001
- 未授权：1010002
- 内部错误：1019999

## 完整示例

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/ink-yht-code/gint/gint"
    "github.com/ink-yht-code/gint/gint/gctx"
    "github.com/ink-yht-code/gint/gintx/db"
    "github.com/ink-yht-code/gint/gintx/httpx"
    "github.com/ink-yht-code/gint/gintx/log"
    "github.com/ink-yht-code/gint/gintx/tx"
)

func main() {
    // 初始化日志
    log.Init(log.Config{Level: "info", Encoding: "json"})
    
    // 初始化数据库
    database, err := db.New(db.Config{
        DSN:      "user:pass@tcp(127.0.0.1:3306)/mydb",
        MaxOpen:  100,
        MaxIdle:  10,
        LogLevel: "info",
    })
    if err != nil {
        log.Fatal("failed to connect database", zap.Error(err))
    }
    
    // 创建事务管理器
    txMgr := tx.NewManager(database)
    
    // 创建 HTTP 服务器
    server := httpx.NewServer(httpx.Config{Addr: ":8080"})
    
    // 注册路由
    server.Engine.GET("/ping", gint.W(func(ctx *gctx.Context) (gint.Result, error) {
        return gint.Result{Code: 0, Msg: "pong"}, nil
    }))
    
    // 启动
    go server.Run()
    fmt.Println("Server started on :8080")
    
    // 优雅关闭
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    fmt.Println("Shutting down...")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    server.Shutdown(ctx)
    fmt.Println("Server exited")
}
```

## 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

Proprietary License

未经版权所有者书面授权，不得使用、复制、修改或分发本项目的任何部分。

详见 [LICENSE](../LICENSE)。

## 联系方式

- 项目主页: [https://github.com/ink-yht-code/gintx](https://github.com/ink-yht-code/gintx)
- 问题反馈: [https://github.com/ink-yht-code/gintx/issues](https://github.com/ink-yht-code/gintx/issues)

---

Made with ❤️ by [ink-yht-code](https://github.com/ink-yht-code)
