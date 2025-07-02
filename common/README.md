# Common 共享包

本目录包含两个API服务（admin-api和public-api）共享的代码和组件。

## 📁 目录结构

```
common/
├── dao/          # 数据访问层 (Data Access Object)
├── model/        # 数据模型定义
├── constants/    # 业务常量定义
├── client/       # 第三方服务客户端
├── errors/       # 业务错误定义
└── utils/        # 通用工具函数
```

## 📝 各目录用途

### `dao/` - 数据访问层
- 封装MongoDB和Redis的数据库操作
- 为每个model定义对应的DAO接口和实现
- 提供统一的数据访问接口

**示例**:
```go
// user_dao.go
type UserDAO interface {
    Create(ctx context.Context, user *model.User) error
    GetByID(ctx context.Context, id string) (*model.User, error)
    GetByUsername(ctx context.Context, username string) (*model.User, error)
    Update(ctx context.Context, user *model.User) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, page, limit int) ([]*model.User, int64, error)
}
```

### `model/` - 数据模型
- 定义与MongoDB集合对应的Go结构体
- 包含BSON标签和JSON标签
- 可包含模型的简单验证方法

**示例**:
```go
// user.go
type User struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Username    string             `bson:"username" json:"username"`
    Email       string             `bson:"email" json:"email"`
    CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
    UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}
```

### `constants/` - 业务常量
- 定义系统中使用的所有魔法字符串和枚举值
- 按业务领域分散到不同文件
- 避免在代码中直接使用字符串常量

**示例**:
```go
// user_constants.go
const (
    RoleOwner  = "Owner"
    RoleAdmin  = "Admin"
    RoleEditor = "Editor"
    RoleAuthor = "Author"
)
```

### `client/` - 第三方客户端
- 封装对外部服务的调用
- 如邮件服务、文件存储、支付服务等
- 每个服务对应一个子目录

**示例**:
```
client/
├── email/      # 邮件服务客户端
├── storage/    # 文件存储客户端
└── sms/        # 短信服务客户端
```

### `errors/` - 错误定义
- 定义统一的业务错误类型
- 便于在服务间传递和处理错误
- 包含错误码和错误消息

**示例**:
```go
// errors.go
var (
    ErrUserNotFound = errors.New("用户不存在")
    ErrUserExists   = errors.New("用户已存在")
    ErrInvalidPassword = errors.New("密码不正确")
)
```

### `utils/` - 工具函数
- 提供通用的辅助函数
- 如密码加密、分页计算、时间处理等
- 无业务逻辑的纯函数

**示例**:
```go
// password.go
func HashPassword(password string) (string, error)
func VerifyPassword(password, hash string) bool

// pagination.go
func CalculatePagination(page, limit int) (offset int, realLimit int)
```

## 🚫 使用规范

### 禁止事项
- ❌ 不要在common包中包含HTTP相关代码
- ❌ 不要在model中包含数据库操作代码
- ❌ 不要在common包中依赖具体的API服务

### 推荐做法
- ✅ 保持接口简单和通用
- ✅ 使用依赖注入的方式
- ✅ 为所有公共代码编写单元测试
- ✅ 保持向后兼容性

## 📚 开发指南

1. **添加新模型**: 在`model/`中定义结构体，在`dao/`中添加对应的接口和实现
2. **添加常量**: 在`constants/`中按业务领域分类添加
3. **添加工具函数**: 在`utils/`中添加纯函数，避免副作用
4. **添加第三方客户端**: 在`client/`中创建子目录，封装外部服务调用

遵循这些规范可以确保common包的代码质量和可维护性。 