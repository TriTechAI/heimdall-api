# Heimdall 项目状态 (Project Status)

## ✅ **已完成的工作**

### 1. 项目架构设计
- [x] 确定微服务架构方案 (`admin-api` + `public-api` + `common`)
- [x] 制定完整的开发规范体系
- [x] 设计Go Workspace工作模式

### 2. 规范文档体系 (`docs/`)
- [x] `AI-PROMPT-GUIDELINES.md` - AI协作指令
- [x] `GO-ZERO-GUIDELINES.md` - Go-Zero开发规范
- [x] `TDD-GUIDELINES.md` - 测试驱动开发规范
- [x] `API-DESIGN-GUIDELINES.md` - RESTful API设计规范
- [x] `MONGODB-MODELING-GUIDELINES.md` - MongoDB建模规范
- [x] `MULTI-SERVICE-ARCHITECTURE.md` - 微服务架构规范
- [x] `BLOG-SYSTEM-CONSIDERATIONS.md` - 博客系统特殊考虑
- [x] `CODE-REVIEW-GUIDELINES.md` - 代码审查流程和质量标准
- [x] `CONTRIBUTING.md` - Git工作流和贡献指南

### 3. 项目基础架构
- [x] ~~创建Go Workspace (`go.work`)~~ (已废弃，改为统一模块)
- [x] ~~初始化三个Go模块 (`admin-api`, `public-api`, `common`)~~ (已重构为统一模块)
- [x] **重构为统一Go模块** (`go.mod`)
- [x] 使用goctl生成服务基础代码
- [x] 创建`common`包目录结构

### 4. 设计文档体系 (`design/`)
- [x] `SYSTEM-ARCHITECTURE-AND-MODULES.md` - 系统架构和模块设计
- [x] `DATA-MODEL-DESIGN.md` - 完整的MongoDB数据模型设计
- [x] `SECURITY-DESIGN.md` - 安全架构和防护策略设计
- [x] `API-INTERFACE-SPECIFICATION.md` - 完整的API接口规范文档

### 5. 最近重要变更
- [x] **安全架构调整** (2024-01): 
  - 将Admin API从内网/VPN部署调整为公网部署
  - 采用账号密码登录方式，强化多重安全防护
  - 新增登录失败锁定、IP限流、操作审计等安全机制
  - 更新了相关设计文档和规范文档

### 6. 文档一致性校验和优化
- [x] **文档冲突检查和修复** (2024-01):
  - [x] 统一API版本控制：所有接口添加 `/api/v1` 前缀
  - [x] 统一响应格式：分页和错误响应格式标准化
  - [x] 统一参数命名：查询参数从 `pageSize` 改为 `limit`
  - [x] 统一认证描述：JWT Token认证机制描述一致
  - [x] 统一技术栈描述：Go、go-zero、MongoDB、Redis版本一致
  - [x] 集合命名规范说明：添加兼容性考虑的说明
  - [x] 跨文档引用一致性：确保docs和design目录间的描述统一

### 7. 架构重构优化
- [x] **统一Go模块架构** (2024-01):
  - [x] 将多模块结构重构为统一模块管理
  - [x] 删除子目录中的独立 `go.mod` 文件
  - [x] 在根目录创建统一的 `go.mod` 文件
  - [x] 更新所有相关文档，反映新的模块结构
  - [x] 更新README.md，提供完整的项目说明

## 📋 **当前项目结构**

```
heimdall-api/
├── go.mod                      # 统一的模块定义文件
├── docs/                       # 规范文档
├── design/                     # 设计文档
├── admin-api/                  # 后台管理服务
│   └── admin/
│       ├── admin.go
│       ├── admin.api
│       ├── etc/
│       └── internal/
├── public-api/                 # 前台公开服务
│   └── public/
│       ├── public.go
│       ├── public.api
│       ├── etc/
│       └── internal/
├── common/                     # 共享包
│   ├── dao/
│   ├── model/
│   ├── constants/
│   ├── client/
│   ├── errors/
│   └── utils/
└── PROJECT-STATUS.md           # 本文档
```

## 🚀 下一步开发计划 (Next Development Plan)

### **任务拆分原则**
- **原子性**: 每个任务都可独立开发、测试、部署
- **明确性**: 每个任务都有清晰的验收标准和预估时间
- **可追踪**: 任务间依赖关系明确，便于项目管理
- **可测试**: 每个任务都包含相应的测试要求

### **阶段零: 基础设施准备 (Foundation)**

#### **0.1: 开发环境与工具链 (2小时)**
- [x] `(#T000)` [P0][Project] **配置开发环境** *(30分钟)* ✅ DONE
  - [x] 验证Go 1.24.4+环境
  - [x] 安装MongoDB和Redis本地服务(已使用docker启动)
  - [x] 配置VSCode/GoLand开发环境(已完成)
  - **验收**: `make deps && make test` 成功执行

- [x] `(#T001)` [P0][Project] **完善项目工具链** *(90分钟)* ✅ DONE
  - [x] 配置golangci-lint和pre-commit hooks(已有.golangci.yml配置)
  - [x] 创建env.example模板文件
  - **验收**: 所有make命令正常工作，代码质量检查通过

#### **0.2: 数据库初始化与种子数据 (3小时)**
- [x] `(#T002)` [P0][common] **数据库连接基础设施** *(90分钟)* ✅ DONE
  - [x] `common/client/mongodb.go`: 实现MongoDB连接池
  - [x] `common/client/redis.go`: 实现Redis连接客户端
  - [x] `common/client/health.go`: 实现健康检查
  - **依赖**: T001
  - **验收**: 连接测试通过，健康检查接口返回正常

- [x] `(#T003)` [P0][common] **创建数据库Schema** *(90分钟)* ✅ DONE
  - [x] 编写MongoDB索引初始化脚本 `scripts/db/create_indexes.js`
  - [x] 创建数据库种子数据脚本 `scripts/db/seed_data.js`
  - [x] 创建便捷的数据库设置脚本 `scripts/setup-database.sh`
  - **依赖**: T002
  - **验收**: 数据库初始化脚本执行成功，种子数据加载正常

### **阶段一: 核心功能 (MVP - Minimum Viable Product)**

#### **1.1: 基础模块和配置 (4小时)**
- [x] `(#T010)` [P1][Config] **服务配置文件完善** *(60分钟)* ✅ DONE
  - [x] 完善 `admin-api.yaml`: 数据库、Redis、JWT、安全配置
  - [x] 完善 `public-api.yaml`: 数据库、Redis、CORS、限流配置
  - [x] 添加配置验证逻辑
  - [x] **重要经验**: 解决了go-zero内置配置字段冲突问题(Log/Timeout字段)
  - [x] **文档更新**: 在规范和设计文档中记录配置冲突解决方案，避免后续踩坑
  - **依赖**: T003
  - **验收**: ✅ 配置文件结构完整，服务启动时验证通过，配置冲突已解决

- [x] `(#T011)` [P1][common] **通用工具包** *(120分钟)* ✅ **DONE** - *2024-01-XX*
  - [x] `common/utils/password.go`: 密码加密/验证(bcrypt)
  - [x] `common/utils/jwt.go`: JWT生成/验证/解析
  - [x] `common/utils/validator.go`: 参数验证工具
  - [x] `common/utils/response.go`: 统一响应格式
  - **验收**: 单元测试覆盖率77%，所有工具函数正常工作，455个断言全部通过

- [x] `(#T012)` [P1][common] **业务常量定义** *(60分钟)* ✅ **DONE** - *2024-01-XX*
  - [x] `common/constants/user.go`: 用户角色、状态常量
  - [x] `common/constants/post.go`: 文章状态常量
  - [x] `common/constants/error.go`: 错误码常量
  - [x] `common/constants/cache.go`: 缓存键常量
  - **验收**: 常量定义完整，命名规范一致，代码编译通过

#### **1.2: 用户与认证模块 (8小时)**
- [x] `(#T020)` [P1][common] **用户数据模型** *(90分钟)* ✅ **DONE** - *2024-01-XX*
  - [x] `common/model/user.go`: 定义User结构体，包含验证规则
  - [x] `common/model/login_log.go`: 定义LoginLog结构体
  - [x] 实现模型的验证方法和转换方法
  - **依赖**: T012 ✅
  - **验收**: 用户模型和登录日志模型结构完整，验证规则正确，96个单元测试通过

- [x] `(#T021)` [P1][common] **用户数据访问层** *(180分钟)* ✅ **DONE** - *2024-01-XX*
  - [x] `common/dao/user_dao.go`: 实现UserDAO接口
    - [x] Create(user *User) error
    - [x] GetByID(id string) (*User, error)
    - [x] GetByUsername(username string) (*User, error)
    - [x] GetByEmail(email string) (*User, error)
    - [x] Update(id string, updates map[string]interface{}) error
    - [x] Delete(id string) error (软删除)
    - [x] List(filter map[string]interface{}, page, limit int) ([]*User, int64, error)
    - [x] UpdateLoginInfo, IncrementLoginFailCount, LockUser, UnlockUser, GetLockedUsers方法
    - [x] CreateIndexes索引创建方法
  - [x] **规范合规修复**: 完全重写测试以符合TDD规范要求
    - [x] 使用mockey框架进行运行时打桩
    - [x] 测试覆盖正常场景和异常场景
    - [x] 64个goconvey BDD风格测试断言全部通过
    - [x] 修复Makefile添加mockey所需编译器标志
  - **依赖**: T020 ✅
  - **验收**: ✅ UserDAO功能完整，包含14个方法，符合TDD-GUIDELINES规范，64个测试断言全部通过，参数验证逻辑100%覆盖，构建测试全部成功

- [x] `(#T022)` [P1][common] **登录日志数据访问层** *(60分钟)* ✅ **DONE** - *2024-01-XX*
  - [x] `common/dao/login_log_dao.go`: 实现LoginLogDAO
    - [x] Create(log *LoginLog) error
    - [x] List(filter map[string]interface{}, page, limit int) ([]*LoginLog, int64, error)
    - [x] 额外实现：GetByUserID, GetByIPAddress, GetRecentFailedLogins, CreateIndexes等方法
  - **依赖**: T020 ✅
  - **验收**: ✅ LoginLogDAO功能完整，包含7个核心方法和2个辅助方法，符合TDD-GUIDELINES规范，47个测试断言全部通过，构建测试全部成功

- [x] `(#T023)` [P1][admin-api] **认证API接口定义** *(60分钟)* ✅ **DONE** - *2024-01-XX*
  - [x] 在`admin.api`中定义认证相关接口
    - [x] POST /api/v1/admin/auth/login
    - [x] GET /api/v1/admin/auth/profile
    - [x] POST /api/v1/admin/auth/logout
  - [x] 定义相应的请求/响应结构体
    - [x] LoginRequest/LoginResponse/LoginData - 登录相关类型
    - [x] ProfileResponse - 用户信息响应类型
    - [x] LogoutRequest/LogoutResponse - 登出相关类型
    - [x] UserInfo, BaseResponse, ErrorResponse等基础类型
  - **依赖**: T020 ✅, T021 ✅, T022 ✅
  - **验收**: ✅ API文件格式正确，goctl代码生成成功，生成了完整的handler/logic/types文件，项目编译通过，路由正确配置JWT保护

- [x] `(#T024)` [P1][admin-api] **用户登录逻辑** *(90分钟)* ✅ **DONE** - *2024-01-XX*
  - [x] 实现LoginLogic：用户名/密码验证
  - [x] 实现登录失败次数限制（Redis缓存）
  - [x] 实现JWT Token生成
  - [x] 记录登录日志
  - [x] ServiceContext依赖注入：MongoDB、Redis、UserDAO、LoginLogDAO
  - [x] 完整的参数验证和错误处理
  - [x] 账户状态检查和自动锁定机制
  - **依赖**: T021 ✅, T022 ✅, T023 ✅
  - **验收**: ✅ 登录功能完整，包含12个核心方法，安全机制完备，编译测试通过，基础验证测试通过

- [ ] `(#T025)` [P1][admin-api] **用户信息获取** *(45分钟)*
  - [ ] 实现GetProfileLogic：获取当前用户信息
  - [ ] 实现JWT中间件验证
  - **依赖**: T024
  - **验收**: 认证用户可正常获取个人信息

- [ ] `(#T026)` [P1][admin-api] **用户管理接口** *(120分钟)*
  - [ ] 在`admin.api`中定义用户管理接口
    - [ ] GET /api/v1/admin/users (分页列表)
    - [ ] GET /api/v1/admin/users/{id} (用户详情)
  - [ ] 实现GetUserListLogic和GetUserDetailLogic
  - **依赖**: T021, T024
  - **验收**: 用户列表和详情接口正常工作

- [ ] `(#T027)` [P1][admin-api] **登录日志管理** *(45分钟)*
  - [ ] 在`admin.api`中定义登录日志接口
    - [ ] GET /api/v1/admin/security/login-logs
  - [ ] 实现GetLoginLogsLogic
  - **依赖**: T022, T024
  - **验收**: 登录日志查询功能正常

#### **1.3: 内容管理模块 - 文章 (10小时)**
- [ ] `(#T030)` [P1][common] **文章数据模型** *(90分钟)*
  - [ ] `common/model/post.go`: 定义Post结构体
    - [ ] 基础字段: ID, Title, Content, Summary, Slug
    - [ ] 元数据: AuthorID, Tags, Categories, Status
    - [ ] 时间字段: CreatedAt, UpdatedAt, PublishedAt
    - [ ] SEO字段: MetaTitle, MetaDescription
  - [ ] 实现slug自动生成和验证
  - **依赖**: T012
  - **验收**: 模型完整，验证规则正确，单元测试通过

- [ ] `(#T031)` [P1][common] **文章数据访问层** *(150分钟)*
  - [ ] `common/dao/post_dao.go`: 实现PostDAO接口
    - [ ] Create(post *Post) error
    - [ ] GetByID(id string) (*Post, error)
    - [ ] GetBySlug(slug string) (*Post, error)
    - [ ] Update(id string, updates map[string]interface{}) error
    - [ ] Delete(id string) error (软删除)
    - [ ] List(filter PostFilter, page, limit int) ([]*Post, int64, error)
    - [ ] GetPublishedList(filter PostFilter, page, limit int) ([]*Post, int64, error)
  - [ ] 实现复杂查询：按标签、分类、作者、状态过滤
  - **依赖**: T030
  - **验收**: 所有方法正确实现，查询性能良好，单元测试覆盖率>90%

- [ ] `(#T032)` [P1][admin-api] **文章管理API接口定义** *(60分钟)*
  - [ ] 在`admin.api`中定义文章管理接口
    - [ ] GET /api/v1/admin/posts (列表，支持过滤)
    - [ ] POST /api/v1/admin/posts (创建)
    - [ ] GET /api/v1/admin/posts/{id} (详情)
    - [ ] PUT /api/v1/admin/posts/{id} (更新)
    - [ ] DELETE /api/v1/admin/posts/{id} (删除)
    - [ ] POST /api/v1/admin/posts/{id}/publish (发布)
    - [ ] POST /api/v1/admin/posts/{id}/unpublish (取消发布)
  - **验收**: API文件正确，goctl生成成功

- [ ] `(#T033)` [P1][admin-api] **文章创建功能** *(90分钟)*
  - [ ] 实现CreatePostLogic：文章创建逻辑
  - [ ] 实现slug重复检查和自动生成
  - [ ] 实现输入验证和安全过滤
  - **依赖**: T031, T032
  - **验收**: 文章创建功能完整，输入验证生效

- [ ] `(#T034)` [P1][admin-api] **文章查询功能** *(75分钟)*
  - [ ] 实现GetPostListLogic：支持多条件过滤、排序
  - [ ] 实现GetPostDetailLogic：获取文章详情
  - [ ] 实现分页和性能优化
  - **依赖**: T031, T032
  - **验收**: 查询功能完整，性能良好

- [ ] `(#T035)` [P1][admin-api] **文章更新功能** *(90分钟)*
  - [ ] 实现UpdatePostLogic：文章更新逻辑
  - [ ] 实现部分更新和版本控制
  - [ ] 实现更新日志记录
  - **依赖**: T031, T032
  - **验收**: 更新功能正常，数据一致性保证

- [ ] `(#T036)` [P1][admin-api] **文章发布管理** *(75分钟)*
  - [ ] 实现PublishPostLogic：文章发布逻辑
  - [ ] 实现UnpublishPostLogic：取消发布逻辑
  - [ ] 实现发布状态验证和时间更新
  - **依赖**: T031, T032
  - **验收**: 发布功能正常，状态管理正确

- [ ] `(#T037)` [P1][admin-api] **文章删除功能** *(45分钟)*
  - [ ] 实现DeletePostLogic：软删除逻辑
  - [ ] 实现删除权限验证
  - **依赖**: T031, T032
  - **验收**: 删除功能安全，可恢复

- [ ] `(#T038)` [P1][public-api] **公开文章API接口定义** *(45分钟)*
  - [ ] 在`public.api`中定义公开文章接口
    - [ ] GET /api/v1/public/posts (公开文章列表)
    - [ ] GET /api/v1/public/posts/{slug} (文章详情)
  - [ ] 定义请求/响应结构体
  - **验收**: API定义正确，无敏感信息泄露

- [ ] `(#T039)` [P1][public-api] **公开文章功能** *(90分钟)*
  - [ ] 实现GetPublicPostListLogic：仅返回已发布文章
  - [ ] 实现GetPublicPostDetailLogic：文章详情，增加浏览计数
  - [ ] 实现缓存机制提升性能
  - **依赖**: T031, T038
  - **验收**: 公开接口正常，性能良好，安全性保证

#### **1.4: 内容管理模块 - 页面 (6小时)**
- [ ] `(#T040)` [P1][common] **页面数据模型** *(60分钟)*
  - [ ] `common/model/page.go`: 定义Page结构体
    - [ ] 基础字段: ID, Title, Content, Slug
    - [ ] 元数据: AuthorID, Status, Template
    - [ ] 时间字段: CreatedAt, UpdatedAt, PublishedAt
    - [ ] SEO字段: MetaTitle, MetaDescription
  - **依赖**: T012
  - **验收**: 页面模型完整，适应静态页面需求

- [ ] `(#T041)` [P1][common] **页面数据访问层** *(90分钟)*
  - [ ] `common/dao/page_dao.go`: 实现PageDAO接口
  - [ ] 基础CRUD操作和查询方法
  - **依赖**: T040
  - **验收**: DAO功能完整，单元测试通过

- [ ] `(#T042)` [P1][admin-api] **页面管理API** *(120分钟)*
  - [ ] 在`admin.api`中定义页面管理接口
  - [ ] 实现页面的CRUD逻辑
  - **依赖**: T041
  - **验收**: 页面管理功能完整

- [ ] `(#T043)` [P1][public-api] **公开页面API** *(90分钟)*
  - [ ] 在`public.api`中定义公开页面接口
  - [ ] 实现GetPublicPageDetailLogic
  - **依赖**: T041
  - **验收**: 公开页面访问正常

#### **1.5: MVP集成测试与验收 (4小时)**
- [ ] `(#T050)` [P1][Test] **端到端测试** *(120分钟)*
  - [ ] 编写用户认证流程的E2E测试
  - [ ] 编写文章CRUD操作的E2E测试  
  - [ ] 编写公开API访问的E2E测试
  - **依赖**: T027, T039, T043
  - **验收**: 核心用户场景测试通过

- [ ] `(#T051)` [P1][Security] **安全测试** *(60分钟)*
  - [ ] 验证JWT认证和授权
  - [ ] 验证输入验证和SQL注入防护
  - [ ] 验证CORS和限流配置
  - **依赖**: T050
  - **验收**: 安全检查项目全部通过

- [ ] `(#T052)` [P1][Performance] **性能基准测试** *(60分钟)*
  - [ ] 设置性能基准指标
  - [ ] 执行负载测试和性能分析
  - [ ] 优化查询和缓存配置
  - **依赖**: T051
  - **验收**: 性能指标达到预期标准

#### **1.6: MVP部署准备 (3小时)**
- [ ] `(#T060)` [P1][Docker] **容器化配置** *(90分钟)*
  - [ ] 编写Dockerfile for admin-api
  - [ ] 编写Dockerfile for public-api
  - [ ] 编写docker-compose.yml for完整环境
  - **验收**: 容器环境正常启动和运行

- [ ] `(#T061)` [P1][CI/CD] **基础CI/CD** *(90分钟)*
  - [ ] 配置GitHub Actions workflow
  - [ ] 实现自动化测试和构建
  - [ ] 配置代码质量检查
  - **验收**: CI/CD流水线正常工作

---

### **阶段二: 高级功能与完善**

#### **2.1: 认证与用户管理完善 (8小时)**
- [ ] `(#T100)` [P2][common] **缓存服务** *(90分钟)*
  - [ ] `common/cache/jwt_blacklist.go`: JWT黑名单缓存
  - [ ] `common/cache/rate_limiter.go`: 限流缓存
  - [ ] `common/cache/session.go`: 会话缓存
  - **验收**: 缓存服务稳定，性能良好

- [ ] `(#T101)` [P2][admin-api] **完整认证功能** *(180分钟)*
  - [ ] 实现用户登出和Token刷新
  - [ ] 实现账户锁定和解锁机制
  - [ ] 实现密码强度验证和修改
  - **依赖**: T100
  - **验收**: 认证安全机制完整

- [ ] `(#T102)` [P2][admin-api] **用户管理功能** *(150分钟)*
  - [ ] 实现用户CRUD操作
  - [ ] 实现角色权限管理
  - [ ] 实现用户状态管理
  - **依赖**: T101
  - **验收**: 用户管理功能完整

- [ ] `(#T103)` [P2][admin-api] **安全中间件** *(120分钟)*
  - [ ] 实现JWT黑名单检查中间件
  - [ ] 实现IP限流中间件
  - [ ] 实现操作审计中间件
  - **依赖**: T100
  - **验收**: 安全防护机制生效

#### **2.2: 标签管理 (P2)**
- [ ] `(#T120)` [P2][common] `[Model]`：在 `common/model/tag.go` 中定义独立的 `Tag` 数据模型。
- [ ] `(#T121)` [P2][common] `[DAO]`：在 `common/dao/tag_dao.go` 中实现 `TagDAO` 的 `Create`, `GetByID`, `GetBySlug`, `Update`, `Delete`, `List` 方法。
- [ ] `(#T122)` [P2][common] `[Test]`：为 `TagDAO` 编写单元测试。
- [ ] `(#T123)` [P2][admin-api] `[Svc]`：在 `ServiceContext` 中注入 `TagDAO`。
- [ ] `(#T124)` [P2][admin-api] `[API]`：在 `admin.api` 中定义标签管理的全部CRUD接口。
- [ ] `(#T125)` [P2][admin-api] `[Logic]`：实现标签管理的全部CRUD逻辑。
- [ ] `(#T126)` [P2][public-api] `[Svc]`：在 `ServiceContext` 中注入 `TagDAO`。
- [ ] `(#T127)` [P2][public-api] `[API]`：在 `public.api` 中定义获取标签列表 `GET /public/tags` 和详情 `GET /public/tags/{slug}` 的接口。
- [ ] `(#T128)` [P2][public-api] `[Logic]`：实现获取公开标签列表和详情的逻辑。

#### **2.3: 评论系统 (P2)**
- [ ] `(#T130)` [P2][common] `[Model]`：在 `common/model/comment.go` 中定义 `Comment` 数据模型。
- [ ] `(#T131)` [P2][common] `[DAO]`：在 `common/dao/comment_dao.go` 中实现 `CommentDAO`。
- [ ] `(#T132)` [P2][common] `[Test]`：为 `CommentDAO` 编写单元测试。
- [ ] `(#T133)` [P2][admin-api] `[Svc]`：在 `ServiceContext` 中注入 `CommentDAO`。
- [ ] `(#T134)` [P2][admin-api] `[API]`：在 `admin.api` 中定义获取评论列表接口 `GET /admin/comments`。
- [ ] `(#T135)` [P2][admin-api] `[Logic]`：实现 `GetCommentListLogic`。
- [ ] `(#T136)` [P2][admin-api] `[API]`：在 `admin.api` 中定义审核评论接口 `PUT /admin/comments/{id}/approve`。
- [ ] `(#T137)` [P2][admin-api] `[Logic]`：实现 `ApproveCommentLogic`。
- [ ] `(#T138)` [P2][admin-api] `[API]`：在 `admin.api` 中定义拒绝评论接口 `PUT /admin/comments/{id}/reject`。
- [ ] `(#T139)` [P2][admin-api] `[Logic]`：实现 `RejectCommentLogic`。
- [ ] `(#T140)` [P2][admin-api] `[API]`：在 `admin.api` 中定义删除评论接口 `DELETE /admin/comments/{id}`。
- [ ] `(#T141)` [P2][admin-api] `[Logic]`：实现 `DeleteCommentLogic`。
- [ ] `(#T142)` [P2][admin-api] `[API]`：在 `admin.api` 中定义批量操作评论接口 `POST /admin/comments/batch`。
- [ ] `(#T143)` [P2][admin-api] `[Logic]`：实现 `BatchCommentLogic`。
- [ ] `(#T144)` [P2][public-api] `[Svc]`：在 `ServiceContext` 中注入 `CommentDAO`。
- [ ] `(#T145)` [P2][public-api] `[API]`：在 `public.api` 中定义获取评论接口 `GET /public/posts/{postId}/comments`。
- [ ] `(#T146)` [P2][public-api] `[Logic]`：实现 `GetPublicCommentsLogic`。
- [ ] `(#T147)` [P2][public-api] `[API]`：在 `public.api` 中定义提交评论接口 `POST /public/posts/{postId}/comments`。
- [ ] `(#T148)` [P2][public-api] `[Logic]`：实现 `SubmitCommentLogic`。
- [ ] `(#T149)` [P2][public-api] `[API]` & `[Logic]`：实现评论点赞相关接口。

#### **2.4: 内容发现 (Public API) (P2)**
- [ ] `(#T160)` [P2][public-api] `[API]` & `[Logic]`：实现获取热门文章接口 `GET /public/posts/popular`。
- [ ] `(#T161)` [P2][public-api] `[API]` & `[Logic]`：实现获取最新文章接口 `GET /public/posts/recent`。
- [ ] `(#T162)` [P2][public-api] `[API]` & `[Logic]`：实现获取作者列表接口 `GET /public/authors`。
- [ ] `(#T163)` [P2][public-api] `[API]` & `[Logic]`：实现获取作者详情接口 `GET /public/authors/{username}`。
- [ ] `(#T164)` [P2][public-api] `[API]` & `[Logic]`：实现获取归档列表接口 `GET /public/archives`。

---

### **阶段三: 站点配置与扩展**
- [ ] `(#T200)` [P2][common] `[Model]` & `[DAO]`：实现 `Setting` 模型与DAO，并编写单元测试。
- [ ] `(#T201)` [P2][admin-api] `[API]` & `[Logic]`：实现站点设置的读写接口 `GET/PUT /admin/settings`。
- [ ] `(#T202)` [P2][common] `[Model]` & `[DAO]`：实现 `Navigation` 模型与DAO，并编写单元测试。
- [ ] `(#T203)` [P2][admin-api] `[API]` & `[Logic]`：实现导航菜单的CRUD和排序接口。
- [ ] `(#T204)` [P2][public-api] `[API]` & `[Logic]`：实现获取站点信息 `GET /public/site/info` 和导航 `GET /public/navigation` 的接口。
- [ ] `(#T210)` [P2][common] `[Client]`：实现文件存储客户端接口 `Storage` (支持本地和云存储)。
- [ ] `(#T211)` [P2][common] `[Model]` & `[DAO]`：实现 `Media` 模型与DAO，并编写单元测试。
- [ ] `(#T212)` [P2][admin-api] `[API]` & `[Logic]`：实现媒体文件的上传和管理接口。

---

### **阶段四: 分析、SEO与运维**

#### **4.1: 数据分析与统计 (4小时)**
- [ ] `(#T300)` [P2][admin-api] **Dashboard统计** *(150分钟)*
  - [ ] 实现基础数据统计API
  - [ ] 用户、文章、访问量统计
  - [ ] 趋势分析和图表数据
  - **验收**: 统计数据准确，图表展示清晰

- [ ] `(#T301)` [P2][public-api] **搜索功能** *(90分钟)*
  - [ ] 实现全文搜索接口
  - [ ] 支持文章和页面搜索
  - [ ] 搜索结果排序和高亮
  - **验收**: 搜索功能准确，响应速度快

#### **4.2: SEO与内容发现 (3小时)**
- [ ] `(#T310)` [P2][public-api] **SEO支持** *(120分钟)*
  - [ ] 实现sitemap.xml生成
  - [ ] 实现RSS feeds生成
  - [ ] 实现OpenGraph和Twitter Cards
  - **验收**: SEO元数据完整，搜索引擎友好

- [ ] `(#T311)` [P2][public-api] **内容发现** *(60分钟)*
  - [ ] 实现热门文章推荐
  - [ ] 实现相关文章推荐
  - [ ] 实现标签云和归档
  - **验收**: 内容发现功能增强用户体验

#### **4.3: 监控与日志 (5小时)**
- [ ] `(#T320)` [P2][common] **日志系统** *(120分钟)*
  - [ ] `common/log/logger.go`: 结构化日志
  - [ ] `common/log/middleware.go`: 请求日志中间件
  - [ ] 分级日志和日志轮转配置
  - **验收**: 日志系统完整，便于调试和监控

- [ ] `(#T321)` [P2][common] **监控指标** *(90分钟)*
  - [ ] `common/metrics/prometheus.go`: Prometheus指标
  - [ ] API响应时间、错误率、QPS监控
  - [ ] 数据库连接池和缓存命中率监控
  - **验收**: 监控指标覆盖关键业务点

- [ ] `(#T322)` [P2][Both] **健康检查** *(60分钟)*
  - [ ] 实现服务健康检查端点
  - [ ] 数据库和Redis连接检查
  - [ ] 服务依赖检查
  - **验收**: 健康检查准确反映服务状态

- [ ] `(#T323)` [P2][Project] **告警配置** *(60分钟)*
  - [ ] 配置关键指标告警
  - [ ] 错误率和响应时间阈值告警
  - [ ] 服务不可用告警
  - **验收**: 告警及时准确，减少故障影响

#### **4.4: 文档与质量 (4小时)**
- [ ] `(#T330)` [P2][Project] **API文档生成** *(90分钟)*
  - [ ] 基于.api文件自动生成Swagger文档
  - [ ] 接口文档在线预览和测试
  - [ ] 文档版本管理
  - **验收**: API文档完整准确，便于前端对接

- [ ] `(#T331)` [P2][Project] **代码质量** *(90分钟)*
  - [ ] 完善golangci-lint配置
  - [ ] 添加安全扫描(gosec)
  - [ ] 代码覆盖率报告
  - **验收**: 代码质量检查严格，覆盖率>80%

- [ ] `(#T332)` [P2][Project] **部署文档** *(60分钟)*
  - [ ] 编写部署指南
  - [ ] 环境配置说明
  - [ ] 故障排查手册
  - **验收**: 文档完整，便于运维部署

## 🔧 **技术债务和持续优化**

### **已解决项目**
- [x] 添加项目级别的Makefile *(已完成)*
- [x] 配置基础开发环境 *(已完成)*

### **待优化项目**
- [ ] **性能优化** (2小时)
  - [ ] 数据库查询优化和索引调优
  - [ ] Redis缓存策略优化
  - [ ] 静态资源CDN配置

- [ ] **安全加固** (3小时)
  - [ ] 完善输入验证和XSS防护
  - [ ] 实现CSRF保护
  - [ ] 强化API限流和防爬虫

- [ ] **可扩展性** (4小时)
  - [ ] 配置负载均衡
  - [ ] 实现读写分离
  - [ ] 微服务拆分规划

- [ ] **开发体验** (2小时)
  - [ ] 完善开发工具链
  - [ ] 添加调试和性能分析工具
  - [ ] 优化本地开发环境

## 📊 **任务拆分总结与分析**

### **拆分优化结果**
经过重新设计，任务拆分具备以下特点：

#### **✅ 优化成果**
1. **原子性**: 每个任务独立可测试，平均耗时60-150分钟
2. **依赖清晰**: 明确标注任务间依赖关系，便于并行开发
3. **验收标准**: 每个任务都有明确的完成标准和质量要求
4. **时间估算**: 提供详细的工作量估算，便于项目管理
5. **优先级分层**: P0基础设施 → P1核心MVP → P2高级功能

#### **📈 关键指标**
- **总任务数**: 66个独立任务
- **MVP预计工时**: 35小时 (约1周完成核心功能)
- **完整系统预计工时**: 65小时 (约2周完成所有功能)
- **测试覆盖**: 每个模块都包含单元测试和集成测试
- **质量保证**: 包含安全测试、性能测试、代码质量检查

#### **🔄 开发流程**
```
阶段零(5h) → 阶段一MVP(30h) → 阶段二扩展(15h) → 阶段三运维(10h) → 阶段四优化(5h)
基础设施    →    核心功能    →    高级功能    →    监控部署    →    持续优化
```

### **最佳实践建议**

#### **任务执行策略**
1. **并行开发**: 同一阶段内的不同模块可并行开发
2. **增量交付**: 每完成一个子模块立即进行集成测试
3. **质量优先**: 严格执行验收标准，确保代码质量
4. **文档同步**: 随代码开发同步更新API文档

#### **风险控制**
1. **技术风险**: 在阶段零完成基础设施验证
2. **集成风险**: 每个阶段结束进行完整集成测试
3. **性能风险**: 在MVP阶段就开始性能基准测试
4. **安全风险**: 从用户认证模块开始就引入安全测试

## 📚 **开发指南**

### **开发前准备**
开发任何新功能前，请务必：

1. **阅读规范**: 查阅 `docs/` 目录下的所有规范文档
2. **遵循TDD**: 先写测试，再写实现  
3. **API优先**: 从定义 `.api` 文件开始
4. **服务判断**: 确定功能属于哪个服务 (admin vs public vs common)
5. **依赖检查**: 确认前置任务已完成且通过验收
6. **代码审查**: 确保符合所有既定规范

### **任务执行流程**
1. **任务领取**: 从PROJECT-STATUS.md选择可执行任务
2. **环境准备**: 确保开发环境符合要求
3. **编码实现**: 严格按照任务描述和验收标准开发
4. **自测验证**: 执行相关测试确保功能正常
5. **代码审查**: 提交PR并通过代码审查
6. **集成测试**: 在集成环境验证功能
7. **任务完成**: 更新任务状态并记录完成情况

### **质量标准**
- **代码覆盖率**: 单元测试覆盖率 ≥ 90%
- **性能标准**: API响应时间 < 200ms (95分位)
- **安全标准**: 通过所有安全扫描检查
- **文档标准**: API文档完整准确，代码注释清晰 