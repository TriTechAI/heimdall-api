# Heimdall API 代码风格指南

本文档定义了Heimdall API项目的统一代码风格规范。所有贡献者必须遵循这些规范以保持代码的一致性和可维护性。

## 1. 文件和包命名

### 1.1 文件命名
- 使用全小写字母
- 单词间不使用分隔符
- 测试文件以`_test.go`结尾

```go
// ✅ 正确
userlogic.go
postdao.go
jwt_test.go

// ❌ 错误
user_logic.go
post-dao.go
UserLogic.go
```

### 1.2 包命名
- 使用简短、有意义的小写单词
- 避免使用下划线或混合大小写
- 包名应该与目录名一致

```go
// ✅ 正确
package user
package dao
package constants

// ❌ 错误
package user_service
package userService
package CONSTANTS
```

## 2. 命名规范

### 2.1 变量命名
- 使用驼峰式命名（camelCase）
- 首字母小写（除非需要导出）
- 避免使用单字母变量（除了循环计数器）

```go
// ✅ 正确
var userCount int
var isActive bool
var maxRetryTimes int

// ❌ 错误
var user_count int
var IsActive bool  // 非导出变量
var n int          // 含义不明
```

### 2.2 常量命名
- 导出常量使用大写字母和下划线分隔
- 未导出常量使用驼峰式命名

```go
// ✅ 正确
const MAX_RETRY_COUNT = 3
const DEFAULT_PAGE_SIZE = 10
const defaultTimeout = 30 * time.Second

// ❌ 错误
const maxRetryCount = 3  // 导出常量应该全大写
const TIMEOUT = 30       // 缺少单位说明
```

### 2.3 函数命名
- 使用驼峰式命名
- 动词或动词短语
- 导出函数首字母大写

```go
// ✅ 正确
func getUserByID(id string) (*User, error)
func ValidateEmail(email string) error
func (s *UserService) CreateUser(user *User) error

// ❌ 错误
func get_user_by_id(id string) (*User, error)
func validate_email(email string) error
func User(id string) (*User, error)  // 名词作函数名
```

### 2.4 接口命名
- 单方法接口使用"方法名+er"
- 多方法接口使用描述性名称

```go
// ✅ 正确
type Reader interface {
    Read([]byte) (int, error)
}

type UserRepository interface {
    Create(user *User) error
    GetByID(id string) (*User, error)
    Update(user *User) error
}

// ❌ 错误
type UserInterface interface {}  // 避免Interface后缀
type IUser interface {}          // 避免I前缀
```

## 3. 注释规范

### 3.1 注释语言
- **统一使用中文注释**
- 注释应该清晰、简洁、有价值

### 3.2 包注释
```go
// Package user 提供用户管理相关功能
// 包括用户的创建、查询、更新和删除操作
package user
```

### 3.3 函数注释
```go
// GetUserByID 根据用户ID获取用户信息
// 参数:
//   - id: 用户唯一标识
// 返回值:
//   - *User: 用户信息，未找到时返回nil
//   - error: 错误信息
func GetUserByID(id string) (*User, error) {
    // 实现代码
}
```

### 3.4 结构体和字段注释
```go
// User 用户信息结构体
type User struct {
    // ID 用户唯一标识
    ID string `json:"id" bson:"_id"`
    
    // Username 用户名，唯一
    Username string `json:"username" bson:"username"`
    
    // Email 用户邮箱，唯一
    Email string `json:"email" bson:"email"`
    
    // CreatedAt 创建时间
    CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
}
```

## 4. 代码组织

### 4.1 导入分组
按以下顺序组织导入，每组之间用空行分隔：
1. 标准库
2. 第三方库
3. 项目内部包

```go
import (
    "context"
    "fmt"
    "time"
    
    "github.com/zeromicro/go-zero/core/logx"
    "go.mongodb.org/mongo-driver/bson"
    
    "github.com/heimdall-api/common/constants"
    "github.com/heimdall-api/common/errors"
    "github.com/heimdall-api/common/model"
)
```

### 4.2 函数长度
- 函数长度不超过25行
- 复杂逻辑拆分为多个子函数
- 每个函数只做一件事

```go
// ✅ 正确：函数职责单一，长度适中
func (l *CreateUserLogic) CreateUser(req *types.CreateUserRequest) (*types.CreateUserResponse, error) {
    // 1. 验证参数
    if err := l.validateRequest(req); err != nil {
        return nil, err
    }
    
    // 2. 检查用户是否存在
    if exists, err := l.checkUserExists(req.Username, req.Email); err != nil {
        return nil, err
    } else if exists {
        return nil, errors.UserAlreadyExists()
    }
    
    // 3. 创建用户
    user, err := l.createUser(req)
    if err != nil {
        return nil, err
    }
    
    // 4. 构建响应
    return l.buildResponse(user), nil
}
```

### 4.3 错误处理
- 立即处理错误
- 避免忽略错误
- 提供有意义的错误信息

```go
// ✅ 正确
user, err := dao.GetUserByID(id)
if err != nil {
    if errors.Is(err, mongo.ErrNoDocuments) {
        return nil, errors.UserNotFound()
    }
    return nil, errors.WrapError(err, "查询用户失败")
}

// ❌ 错误
user, _ := dao.GetUserByID(id)  // 忽略错误
```

## 5. 常量定义

### 5.1 避免魔法数字
所有的魔法数字都应该定义为常量：

```go
// ✅ 正确
const (
    MaxLoginAttempts = 5
    TokenExpireHours = 24
    DefaultPageSize  = 10
)

if attempts > MaxLoginAttempts {
    return errors.TooManyAttempts()
}

// ❌ 错误
if attempts > 5 {  // 魔法数字
    return errors.TooManyAttempts()
}
```

### 5.2 时间格式
统一使用RFC3339格式：

```go
// ✅ 正确
const DefaultTimeFormat = time.RFC3339

timeStr := time.Now().Format(constants.DefaultTimeFormat)

// ❌ 错误
timeStr := time.Now().Format("2006-01-02T15:04:05Z07:00")
```

## 6. 测试规范

### 6.1 测试文件命名
- 测试文件以`_test.go`结尾
- 与被测试文件在同一包内

### 6.2 测试函数命名
```go
// 单元测试
func TestUserService_CreateUser(t *testing.T) {
    // 测试代码
}

// 基准测试
func BenchmarkUserService_GetUser(b *testing.B) {
    // 基准测试代码
}
```

### 6.3 测试组织
使用表驱动测试：

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {
            name:    "有效邮箱",
            email:   "test@example.com",
            wantErr: false,
        },
        {
            name:    "无效邮箱",
            email:   "invalid-email",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## 7. 并发编程

### 7.1 Goroutine命名
```go
// ✅ 正确：有意义的goroutine
go func() {
    if err := processUser(user); err != nil {
        log.Errorf("处理用户失败: %v", err)
    }
}()

// ❌ 错误：忽略错误处理
go processUser(user)
```

### 7.2 Channel命名
- 使用描述性名称
- 标明方向（如果是单向的）

```go
// ✅ 正确
type Worker struct {
    taskCh   <-chan Task  // 只读
    resultCh chan<- Result // 只写
}
```

## 8. 性能考虑

### 8.1 避免不必要的内存分配
```go
// ✅ 正确：预分配切片容量
users := make([]User, 0, count)

// ❌ 错误：动态增长
var users []User
for i := 0; i < count; i++ {
    users = append(users, user)
}
```

### 8.2 使用sync.Pool复用对象
```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func processData(data []byte) {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer bufferPool.Put(buf)
    buf.Reset()
    
    // 使用buffer
}
```

## 9. 代码审查清单

提交代码前，请确保：

- [ ] 所有函数都有中文注释
- [ ] 没有魔法数字
- [ ] 函数长度不超过25行
- [ ] 错误都被正确处理
- [ ] 变量命名清晰有意义
- [ ] 测试覆盖主要逻辑
- [ ] 没有TODO或FIXME注释
- [ ] 代码通过`go fmt`格式化
- [ ] 代码通过`golangci-lint`检查

## 10. 工具配置

### 10.1 golangci-lint配置
项目根目录创建`.golangci.yml`：

```yaml
linters:
  enable:
    - gofmt
    - golint
    - govet
    - ineffassign
    - misspell
    - unconvert
    - unparam
    - nakedret
    - prealloc
    - scopelint
    - gocritic

linters-settings:
  funlen:
    lines: 25
    statements: 15
  gocyclo:
    min-complexity: 10
  govet:
    check-shadowing: true
  golint:
    min-confidence: 0.8
  maligned:
    suggest-new: true
  goconst:
    min-len: 2
    min-occurrences: 2
  misspell:
    locale: US
```

### 10.2 pre-commit钩子
```bash
#!/bin/sh
# .git/hooks/pre-commit

# 运行格式化
go fmt ./...

# 运行lint检查
golangci-lint run

# 运行测试
go test ./...
```

---

**注意**: 本规范是活文档，会根据项目发展和团队反馈持续更新。所有更改都应该经过团队讨论和同意。