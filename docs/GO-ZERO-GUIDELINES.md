# Go-Zero 开发规范

本文档旨在为基于 go-zero 框架的项目提供一套统一的开发标准和最佳实践，以确保代码的可读性、可维护性和团队协作效率。

## 1. 核心原则：API-First

**所有开发活动都必须以 API 定义为起点。**

- **唯一事实来源 (Single Source of Truth)**: 项目中所有的 API 路由、请求/响应结构体、服务定义，都**必须**在 `.api` 文件中进行描述。`.api` 文件是项目接口的唯一权威来源。
- **代码生成驱动**:
  - 严禁手动创建或修改 `handler`、`types`、`routes` 等由 `goctl` 生成的代码。
  - **标准工作流**:
    1.  **定义**: 在 `.api` 文件中添加或修改 API 定义。
    2.  **生成**: 运行 `goctl api go -api <your-api-file>.api -dir . --style=goZero` 命令重新生成代码。
    3.  **实现**: 在 `internal/logic` 目录中找到对应的 `logic` 文件，并填充业务逻辑。

## 2. 目录结构与职责划分

严格遵守 `go-zero` 的目录结构约定，并结合以下自定义规范，确保各模块职责清晰。

### 2.1. 多服务架构总览

我们采用微服务架构，将 `heimdall` 项目拆分为两个独立的服务：

```
heimdall-api/
├── admin-api/                  # 后台管理服务
│   ├── go.mod
│   └── admin/                  # goctl生成的服务代码
│       ├── admin.go
│       ├── admin.api
│       ├── etc/
│       └── internal/
│           ├── handler/
│           ├── logic/
│           ├── svc/
│           ├── types/
│           ├── config/
│           └── middleware/
├── public-api/                 # 前台公开服务  
│   ├── go.mod
│   └── public/                 # goctl生成的服务代码
│       ├── public.go
│       ├── public.api
│       ├── etc/
│       └── internal/
│           ├── handler/
│           ├── logic/
│           ├── svc/
│           ├── types/
│           ├── config/
│           └── middleware/
├── common/                     # 共享代码模块
│   ├── dao/                    # 数据访问层
│   ├── model/                  # 数据模型
│   ├── constants/              # 共享常量
│   ├── client/                 # 第三方服务客户端
│   ├── errors/                 # 业务错误定义
│   └── go.mod
├── go.work                     # Go Workspace 配置
└── [规范文档...]
```

### 2.2. 服务职责划分

- **`admin-api`**: 
  - **用途**: 博客管理后台，面向作者、编辑、管理员
  - **主要功能**: 用户认证、文章CRUD、评论管理、系统设置、数据统计
  - **安全级别**: 高，建议部署在内网或通过VPN访问
  - **访问量**: 低，主要是管理人员使用

- **`public-api`**:
  - **用途**: 博客前台，面向公众读者  
  - **主要功能**: 文章展示、评论查看、搜索、标签浏览
  - **安全级别**: 公开，需要防范各种网络攻击
  - **访问量**: 高，需要考虑性能优化和缓存

- **`common`**:
  - **用途**: 两个服务共享的核心业务代码
  - **包含**: 数据模型、数据访问层、业务常量、第三方客户端等
  - **原则**: 只包含纯业务逻辑，不包含HTTP相关代码

### 2.3. 单个服务内部结构

每个服务 (`admin-api`, `public-api`) 内部严格遵守 go-zero 的标准结构：

- `etc/`: 仅用于存放项目配置文件 (`.yaml`)。
- `internal/config/`: 由 `goctl` 生成的配置结构体。**禁止手动修改**。
- `internal/handler/`: **保持精简**。Handler 仅负责解析请求、调用 `logic` 层、并返回响应。**严禁在 Handler 中编写任何业务逻辑。**
- `internal/logic/`: **核心业务逻辑层**。负责编排和组合业务流程。它调用 `common/dao` 层进行数据持久化，调用 `common/client` 层与第三方服务交互。
- `internal/svc/`: **服务上下文 (ServiceContext)**。作为依赖注入的容器。在此初始化并持有项目所需的全部外部依赖（配置、DAO 实例、第三方服务 Client 等），并注入到 `logic` 层。
- `internal/types/`: 由 `goctl` 自动生成，存放 API 的请求和响应结构体。**禁止手动修改**。
- `internal/middleware/`: 存放服务特有的中间件。

### 2.4. 共享模块结构 (`common/`)

- `common/model/`: **（手动创建）模型定义层**。
  - **只存放**与数据表一一映射的 Go `struct` 定义。
  - 可以包含模型自身的一些简单、无依赖的方法（例如 `IsPublished()`)。
  - **严禁**在此层包含任何数据库操作（CURD）代码。
- `common/dao/`: **（手动创建）数据访问层 (Data Access Object)**。
  - **专门用于与数据库交互**。
  - 为每个 `model` 定义一个 `interface` 和它的实现。例如 `user_dao.go` 包含 `UserDao` 接口和 `NewUserDao(...)`。
  - 负责所有 SQL/ORM 操作。`logic` 层通过调用 `dao` 层的接口来操作数据。
- `common/client/`: **（手动创建）第三方服务客户端层**。
  - 封装对外部第三方 API (如邮件服务、支付网关等) 的调用。
  - 每个第三方服务对应一个子目录，例如 `common/client/mailgun/`。
- `common/constants/`: 存放所有共享常量（魔法字符串、数字等）。
- `common/errors/`: **推荐创建**，用于定义业务错误类型。

## 3. API 定义规范 (`.api` 文件)

- **命名规范**:
  - `type` 名称使用 `PascalCase`。
  - `type` 中的字段名使用 `PascalCase`。对应的 `json` 标签使用 `camelCase`。
  - 路由 Handler 名称应清晰描述其功能，例如 `GetUserById`。
- **注释**:
  - 必须为每个 `service`、`route`、和 `type` 中的关键字段添加清晰的注释，解释其用途。
- **版本管理**: 推荐在 API 路由中包含版本标识，例如 `@server(prefix=/api/v1)`。
- **路由规则**:
  - 命名采用 RESTful 风格。
  - 使用 `@doc("...")` 描述接口功能。
  - 使用 `@handler ...` 指定 handler 名称。

**示例:**
```api
type (
    // 用户登录请求
    LoginReq {
        Username string `json:"username"` // 用户名
        Password string `json:"password"` // 密码
    }

    // 用户登录响应
    LoginReply {
        AccessToken  string `json:"accessToken"`
        AccessExpire int64  `json:"accessExpire"`
    }
)

@server(
    jwt: Auth
    prefix: /api/v1/users
)
service user {
    @doc("用户登录")
    @handler login
    post /login (LoginReq) returns (LoginReply)
}
```

## 4. 错误处理

- **Logic 层**:
  - 业务逻辑中遇到错误时，应立即 `return nil, err`，将错误向上传递。
  - 推荐调用 `internal/errors/` 中预定义的业务错误，而不是直接返回 `errors.New("...")`。
- **Handler 层**:
  - `go-zero` 默认的 `httpx.OkJson` 和 `httpx.Error` 会处理大部分情况。
  - 可通过自定义中间件或改写 `httpx.SetErrorHandler` 来实现更复杂的全局错误处理逻辑（例如，将不同类型的 `error` 映射为不同的 HTTP 状态码和响应格式）。

## 5. 配置与数据库

- **配置**:
  - 严禁在代码中硬编码任何配置项（端口、数据库地址、密钥等）。所有配置都必须在 `etc/*.yaml` 文件中定义。
  - 对于敏感信息（如密码、API Key），推荐通过环境变量加载，并在配置文件中使用 `env()` 语法。
- **数据库**:
  - 我们使用 MongoDB 作为主数据库，通过官方的 `go.mongodb.org/mongo-driver` 与数据库交互。
  - 数据库连接对象在 `svc.ServiceContext` 中初始化并持有。
  - `logic` 层**不应**直接操作 MongoDB 驱动，而应调用 `dao` 层提供的方法来完成数据操作。

## 6. 常量管理

为了提高代码的可维护性、可读性并避免魔法字符串 (Magic Strings)，项目中所有可复用的、具有业务含义的字符串或数字都必须定义为常量。

### 6.1. 常量定义位置

- **共享常量**: 在 `common/constants/` 目录下定义两个服务都会使用的常量
- **服务特有常量**: 在各自服务的 `internal/constants/` 目录下定义只有该服务使用的常量
- **按领域分类**: 根据常量的业务领域将其分散到不同的文件中

### 6.2. 常量定义示例

**共享常量** (`common/constants/`):

- `common/constants/post_constants.go`:
  ```go
  package constants

  const (
      // 文章状态
      PostStatusPublished = "published"
      PostStatusDraft     = "draft"
      PostStatusArchived  = "archived"
      PostStatusScheduled = "scheduled"
  )
  ```

- `common/constants/user_constants.go`:
  ```go
  package constants

  const (
      // 用户角色
      RoleOwner   = "Owner"        // 博客所有者
      RoleAdmin   = "Admin"        // 管理员
      RoleEditor  = "Editor"       // 编辑
      RoleAuthor  = "Author"       // 作者
      RoleViewer  = "Viewer"       // 访客/会员
  )
  ```

**服务特有常量** (如 `admin-api/admin/internal/constants/`):

- `admin-api/admin/internal/constants/context_constants.go`:
  ```go
  package constants

  type ContextKey string

  const (
      CtxKeyUserID   ContextKey = "userID"
      CtxKeyUserRole ContextKey = "userRole"
  )
  ```

### 6.3. 引用规范

- **引用共享常量**: `import "github.com/heimdall-api/common/constants"`
- **引用服务常量**: `import "github.com/heimdall-api/admin-api/admin/internal/constants"` 