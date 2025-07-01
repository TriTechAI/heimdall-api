# TDD (测试驱动开发) 规范

本规范旨在定义项目测试驱动开发的流程和标准，以确保代码质量、逻辑正确性，并提高代码的可维护性。

## 1. 核心理念

我们依然遵循 **红灯 -> 绿灯 -> 重构** 的 TDD 循环，但在工具选型上，我们追求更现代、更高效的方式。

## 2. 核心测试框架

我们选用以下业界领先的测试框架作为我们的技术栈：

- **核心测试框架**: `github.com/smartystreets/goconvey/convey`
  - 提供行为驱动开发（BDD）风格的语法，让测试用例的结构更清晰、更具可读性。
- **Mock 框架**: `github.com/bytedance/mockey`
  - 通过运行时打桩（Patching）的方式，允许我们直接 Mock 具体函数或方法，无需为测试而创建 `interface`。
- **MongoDB 集成测试**: `go.mongodb.org/mongo-driver/mongo/integration/mtest`
  - 用于编写需要真实数据库交互的集成测试。
- **Redis 内存测试**: `github.com/alicebob/miniredis/v2`
  - 提供一个内存中的 Redis 服务，用于快速的缓存层测试。

## 3. `Logic` 单元测试实战 (`goconvey` + `mockey`)

这是我们项目的核心测试模式。我们将彻底告别 `gomock` 和 `interface` 驱动的测试，拥抱 `mockey` 带来的灵活性。

**核心思想**: 在单元测试中，我们不关心 `DAO` 或 `Client` 的内部实现，只关心它的**输入和输出**。`mockey` 让我们可以在运行时**替换**掉一个真实的函数实现，让它返回我们指定的任何结果。

### 步骤 1: 编写 `DAO` 层（无需接口）

我们直接编写 `DAO` 的具体实现。注意，为了让 `mockey` 能够成功 Patch，**结构体方法接收者必须是指针类型**。

```go
// file: common/dao/user_dao.go
package dao

// ... imports ...

type UserDAO struct { // 可以不叫 Impl 了
    // ... 依赖
}

func NewUserDAO(...) *UserDAO { ... }

// 注意，这里的接收者是 *UserDAO
func (d *UserDAO) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
    // ... 真实的数据库查询逻辑 ...
}
```

### 步骤 2: 编写 `Logic` 测试用例

假设我们要测试 `CreateUserLogic`，它依赖 `UserDAO`。

```go
// file: admin-api/admin/internal/logic/user/create_user_logic_test.go
package user

import (
    "context"
    "errors"
    "testing"

    . "github.com/bytedance/mockey"
    . "github.com/smartystreets/goconvey/convey"

    "github.com/heimdall-api/common/dao"
    "github.com/heimdall-api/common/model"
    "github.com/heimdall-api/admin-api/admin/internal/svc"
    "github.com/heimdall-api/admin-api/admin/internal/types"
)

func TestCreateUserLogic(t *testing.T) {
    // 1. 创建一个 DAO 实例，它不需要真实的数据库连接。
    // 我们只是需要一个目标来 Patch 它的方法。
    userDAO := &dao.UserDAO{}

    // 2. 创建注入了该 DAO 实例的 ServiceContext
    serviceCtx := &svc.ServiceContext{
        UserDAO: userDAO,
    }
    l := NewCreateUserLogic(context.Background(), serviceCtx)

    // 3. 使用 Convey 来组织测试场景
    Convey("Test Create User Logic", t, func() {
        req := &types.CreateUserReq{Username: "testuser", Password: "password123"}

        Convey("Success Scenario", func() {
            // 4. 使用 mockey 对 DAO 方法进行打桩
            Patch(userDAO.GetUserByUsername, func(_ *dao.UserDAO, _ context.Context, _ string) (*model.User, error) {
                // 模拟用户不存在
                return nil, nil
            }).Build()
            // 在当前 Convey 作用域结束后，自动取消所有桩
            defer Unpatch()

            resp, err := l.CreateUser(req)

            // 5. 使用 Convey 的 So 进行断言
            So(err, ShouldBeNil)
            So(resp, ShouldNotBeNil)
        })

        Convey("Failure Scenario: Username already exists", func() {
            Patch(userDAO.GetUserByUsername, func(_ *dao.UserDAO, _ context.Context, username string) (*model.User, error) {
                // 模拟用户已存在
                return &model.User{Username: username}, nil
            }).Build()
            defer Unpatch()

            resp, err := l.CreateUser(req)

            So(err, ShouldNotBeNil)
            // So(errors.Is(err, YourCustomError), ShouldBeTrue) // 更具体的错误断言
            So(resp, ShouldBeNil)
        })
    })
}
```

## 4. 集成与内存测试

当我们需要测试真实的服务交互时（例如，测试复杂的 SQL 查询），我们会利用内存测试工具。

- **Redis**: 使用 `miniredis`，我们可以在测试开始时启动一个内存 Redis 服务，在测试结束时关闭它。测试代码将连接到这个内存服务，从而实现对缓存逻辑的快速、真实的测试。
- **MongoDB**: 类似地，`mtest` 工具提供了在测试期间编程式地设置和清理 MongoDB 测试数据的能力。

这些测试会被标记为集成测试，可以与单元测试分开运行。

## 5. 多服务架构下的测试策略

### 5.1. 测试层级划分

在我们的微服务架构中，测试分为三个层级：

- **Common 模块测试**: 测试共享的数据模型、DAO层和工具函数
- **服务Logic层测试**: 测试各个服务的业务逻辑（`admin-api`, `public-api`）
- **服务集成测试**: 测试完整的HTTP接口调用

### 5.2. Common 模块测试

**测试位置**: `common/` 目录下的各个子包
**测试重点**: 
- `dao/` 包的数据库操作逻辑
- `model/` 包的数据验证方法
- `client/` 包的第三方服务调用

**示例测试**:
```go
// file: common/dao/user_dao_test.go
package dao

func TestUserDAO_CreateUser(t *testing.T) {
    Convey("Test User DAO Create User", t, func() {
        // 使用 mtest 创建内存MongoDB
        mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
        defer mt.Close()
        
        mt.Run("success", func(mt *mtest.T) {
            // Mock MongoDB 操作
            mt.AddMockResponses(mtest.CreateSuccessResponse())
            
            userDAO := NewUserDAO(mt.Client)
            user := &model.User{Username: "test"}
            
            err := userDAO.CreateUser(context.Background(), user)
            So(err, ShouldBeNil)
        })
    })
}
```

### 5.3. 服务Logic层测试

**测试位置**: 各服务的 `internal/logic/` 目录
**Mock策略**: 使用 `mockey` Mock `common/dao` 的方法

**示例测试**:
```go
// file: admin-api/admin/internal/logic/user/create_user_logic_test.go
func TestCreateUserLogic(t *testing.T) {
    // 创建共享DAO的实例用于Mock
    userDAO := &dao.UserDAO{}
    
    Convey("Test Admin Create User Logic", t, func() {
        Convey("Success scenario", func() {
            // Mock common/dao 的方法
            Patch(userDAO.CreateUser, func(*dao.UserDAO, context.Context, *model.User) error {
                return nil
            }).Build()
            defer Unpatch()
            
            // 测试admin-api的创建用户逻辑
            // ...
        })
    })
}
```

### 5.4. 跨服务测试原则

- **服务隔离**: 每个服务的测试必须独立运行，不依赖其他服务
- **共享Mock**: 对 `common` 模块的Mock可以在不同服务的测试中复用
- **数据隔离**: 使用不同的测试数据库或清理策略，避免测试间相互影响