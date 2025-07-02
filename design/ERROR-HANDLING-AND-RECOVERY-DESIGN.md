# 错误处理与恢复设计 (Error Handling & Recovery Design)

本文档定义了 Heimdall 博客系统的错误分类、处理策略、恢复机制和容错设计。

## 1. 错误分类体系

### 1.1. 错误级别分类

**Critical (致命错误)**:
- 数据库连接完全失败
- 核心服务崩溃
- 数据损坏或丢失
- 安全漏洞被利用

**Error (业务错误)**:
- API请求失败
- 业务逻辑错误
- 第三方服务不可用
- 文件系统故障

**Warning (警告)**:
- 性能指标超阈值
- 缓存失效
- 配置项缺失
- 资源使用率高

**Info (信息)**:
- 正常业务流程
- 用户操作记录
- 系统状态变更
- 定期任务执行

### 1.2. 错误码设计

**错误码结构**: `EHHCCCC`
- `E`: 固定前缀
- `HH`: 服务代码 (01=Admin, 02=Public, 99=Common)
- `CCCC`: 错误序号

```go
// common/errors/codes.go
const (
    // 通用错误 (E99xxxx)
    ErrInternalServer     = "E990001"
    ErrInvalidParams      = "E990002" 
    ErrUnauthorized       = "E990003"
    ErrForbidden         = "E990004"
    ErrNotFound          = "E990005"
    ErrRateLimit         = "E990006"
    
    // Admin API错误 (E01xxxx)
    ErrUserNotFound      = "E010001"
    ErrInvalidPassword   = "E010002"
    ErrUserLocked        = "E010003"
    ErrTokenExpired      = "E010004"
    ErrInsufficientPerm  = "E010005"
    ErrPostNotFound      = "E010006"
    
    // Public API错误 (E02xxxx)
    ErrPostNotPublished  = "E020001"
    ErrCommentNotAllowed = "E020002"
    ErrSearchTimeout     = "E020003"
    ErrContentNotFound   = "E020004"
)
```

## 2. 统一错误处理架构

### 2.1. 错误类型定义

```go
// 业务错误接口
type BusinessError interface {
    error
    Code() string
    Message() string
    Details() interface{}
    StatusCode() int
}

// 业务错误实现
type BizError struct {
    code       string
    message    string
    details    interface{}
    statusCode int
    timestamp  time.Time
    cause      error
}

func (e *BizError) Error() string {
    return fmt.Sprintf("[%s] %s", e.code, e.message)
}

// 错误构造函数
func New(code string, details ...interface{}) *BizError {
    return &BizError{
        code:       code,
        message:    ErrorMessages[code],
        details:    getDetails(details),
        statusCode: getStatusCodeByErrorCode(code),
        timestamp:  time.Now(),
    }
}
```

### 2.2. 错误处理中间件

```go
// Go-Zero错误处理中间件
func ErrorHandlerMiddleware() rest.Middleware {
    return func(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if err := recover(); err != nil {
                    logrus.WithFields(logrus.Fields{
                        "error":      err,
                        "stack":      string(debug.Stack()),
                        "method":     r.Method,
                        "path":       r.URL.Path,
                        "remote_ip":  r.RemoteAddr,
                    }).Error("Panic recovered")
                    
                    httpx.Error(w, errors.New(errors.ErrInternalServer))
                }
            }()
            
            next(w, r)
        }
    }
}

// 自定义错误响应处理器
func CustomErrorHandler(err error) (int, interface{}) {
    switch e := err.(type) {
    case *errors.BizError:
        return e.StatusCode(), map[string]interface{}{
            "code":      e.Code(),
            "message":   e.Message(),
            "details":   e.Details(),
            "timestamp": e.Timestamp().Unix(),
        }
    default:
        logrus.WithField("error", err).Error("Unknown error occurred")
        return 500, map[string]interface{}{
            "code":      errors.ErrInternalServer,
            "message":   "内部服务器错误",
            "timestamp": time.Now().Unix(),
        }
    }
}
```

## 3. 层级错误处理策略

### 3.1. Handler层错误处理

```go
func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
    var req types.LoginRequest
    if err := httpx.Parse(r, &req); err != nil {
        httpx.Error(w, errors.New(errors.ErrInvalidParams, err.Error()))
        return
    }
    
    resp, err := h.svcCtx.UserLogic.Login(r.Context(), &req)
    if err != nil {
        httpx.Error(w, err)
        return
    }
    
    httpx.OkJson(w, resp)
}
```

### 3.2. Logic层错误处理

```go
func (l *UserLogic) Login(ctx context.Context, req *types.LoginRequest) (*types.LoginResponse, error) {
    // 参数验证
    if err := l.validateLoginParams(req); err != nil {
        return nil, errors.New(errors.ErrInvalidParams, err.Error())
    }
    
    // 检查登录限制
    if blocked, err := l.svcCtx.LoginLimiter.IsBlocked(ctx, req.Username); err != nil {
        logrus.WithError(err).Error("Failed to check login limiter")
        return nil, errors.Wrap(err, errors.ErrInternalServer)
    } else if blocked {
        return nil, errors.New(errors.ErrUserLocked, "账号已被锁定")
    }
    
    // 获取用户信息
    user, err := l.svcCtx.UserDAO.GetByUsername(ctx, req.Username)
    if err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            l.recordLoginFailure(ctx, req.Username, "user_not_found")
            return nil, errors.New(errors.ErrUserNotFound)
        }
        return nil, errors.Wrap(err, errors.ErrInternalServer)
    }
    
    // 验证密码
    if !l.svcCtx.PasswordEncoder.Verify(req.Password, user.PasswordHash) {
        l.recordLoginFailure(ctx, req.Username, "invalid_password")
        return nil, errors.New(errors.ErrInvalidPassword)
    }
    
    // 生成Token
    token, err := l.svcCtx.JWTGenerator.Generate(user.ID, user.Username, user.Role)
    if err != nil {
        return nil, errors.Wrap(err, errors.ErrInternalServer)
    }
    
    return &types.LoginResponse{
        Token: token,
        User: types.UserInfo{
            ID:       user.ID,
            Username: user.Username,
            Role:     user.Role,
        },
    }, nil
}
```

## 4. 容错与恢复机制

### 4.1. 熔断器模式

```go
type CircuitBreaker struct {
    mu               sync.Mutex
    state            State
    failureCount     int
    successCount     int
    lastFailureTime  time.Time
    failureThreshold int
    timeout          time.Duration
}

func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
    if !cb.canExecute() {
        return errors.New("circuit breaker is open")
    }
    
    err := fn()
    cb.onResult(err == nil)
    return err
}
```

### 4.2. 重试机制

```go
func WithExponentialBackoff(config Config, fn func() error) error {
    delay := config.InitialDelay
    
    for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
        err := fn()
        if err == nil {
            return nil
        }
        
        if attempt == config.MaxAttempts {
            return err
        }
        
        if !isRetryableError(err) {
            return err
        }
        
        time.Sleep(delay)
        delay = time.Duration(float64(delay) * config.Multiplier)
        if delay > config.MaxDelay {
            delay = config.MaxDelay
        }
    }
    
    return nil
}
```

### 4.3. 优雅降级

```go
func (d *PostDAO) GetPopularPosts(ctx context.Context, limit int) ([]*model.Post, error) {
    // 主路径：从数据库获取热门文章
    posts, err := d.getPopularPostsFromDB(ctx, limit)
    if err == nil {
        return posts, nil
    }
    
    // 降级路径1：从缓存获取
    if posts, err := d.getPopularPostsFromCache(ctx); err == nil {
        logrus.Warn("Database unavailable, serving from cache")
        return posts[:min(len(posts), limit)], nil
    }
    
    // 降级路径2：返回最新文章
    if posts, err := d.getLatestPosts(ctx, limit); err == nil {
        logrus.Warn("Popular posts unavailable, serving latest posts")
        return posts, nil
    }
    
    // 降级路径3：返回空列表
    logrus.Error("All data sources unavailable, returning empty list")
    return []*model.Post{}, nil
}
```

## 5. 监控和告警

### 5.1. 错误监控指标

```go
var (
    errorCountTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "errors_total",
            Help: "Total number of errors",
        },
        []string{"service", "error_code", "error_type"},
    )
    
    circuitBreakerStateGauge = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "circuit_breaker_state",
            Help: "Circuit breaker state (0=closed, 1=half-open, 2=open)",
        },
        []string{"service", "name"},
    )
)
```

### 5.2. 告警规则

```yaml
groups:
- name: error-alerts
  rules:
  - alert: HighErrorRate
    expr: |
      (
        rate(errors_total[5m]) / 
        rate(http_requests_total[5m])
      ) > 0.05
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High error rate detected"
      description: "Error rate is {{ $value | humanizePercentage }}"

  - alert: CircuitBreakerOpen
    expr: circuit_breaker_state == 2
    for: 0s
    labels:
      severity: critical
    annotations:
      summary: "Circuit breaker is open"
      description: "Circuit breaker {{ $labels.name }} is open"
```

## 6. 灾难恢复计划

### 6.1. 故障分级

**P0 - 致命故障**:
- 影响: 整个系统不可用
- 响应时间: 15分钟内
- 恢复目标: 30分钟内

**P1 - 严重故障**:
- 影响: 核心功能不可用
- 响应时间: 30分钟内
- 恢复目标: 2小时内

**P2 - 一般故障**:
- 影响: 部分功能异常
- 响应时间: 1小时内
- 恢复目标: 4小时内

### 6.2. 应急预案

**数据库故障应急预案**:
```bash
#!/bin/bash
# 检查主数据库状态
if ! mongo --eval "db.adminCommand('ping')" > /dev/null 2>&1; then
    echo "主数据库不可用，启动故障转移..."
    
    # 切换到备用数据库
    kubectl patch configmap db-config -p '{"data":{"primary_host":"mongodb-secondary"}}'
    
    # 重启应用服务
    kubectl rollout restart deployment/admin-api
    kubectl rollout restart deployment/public-api
    
    echo "故障转移完成"
fi
```

**服务降级脚本**:
```bash
#!/bin/bash
SERVICE_NAME=$1
DEGRADATION_LEVEL=$2

case $DEGRADATION_LEVEL in
    "level1")
        # 关闭非核心功能
        kubectl patch configmap feature-flags -p '{"data":{"enable_search":"false"}}'
        ;;
    "level2")
        # 启用缓存模式
        kubectl patch configmap app-config -p '{"data":{"cache_only":"true"}}'
        ;;
    "level3")
        # 只读模式
        kubectl patch configmap app-config -p '{"data":{"read_only":"true"}}'
        ;;
esac

kubectl rollout restart deployment/$SERVICE_NAME
echo "服务 $SERVICE_NAME 已降级到 $DEGRADATION_LEVEL"
```

---

**注意**: 错误处理和恢复是系统稳定性的基石，需要在开发初期就进行规划和设计。建议定期进行故障演练，验证恢复流程的有效性。 