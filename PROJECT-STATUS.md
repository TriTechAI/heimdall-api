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
- [x] 创建Go Workspace (`go.work`)
- [x] 初始化三个Go模块 (`admin-api`, `public-api`, `common`)
- [x] 使用goctl生成服务基础代码
- [x] 创建`common`模块目录结构

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

## 📋 **当前项目结构**

```
heimdall-api/
├── docs/                       # 规范文档
├── design/                     # 设计文档
├── admin-api/                  # 后台管理服务
│   ├── go.mod
│   └── admin/
│       ├── admin.go
│       ├── admin.api
│       ├── etc/
│       └── internal/
├── public-api/                 # 前台公开服务
│   ├── go.mod
│   └── public/
│       ├── public.go
│       ├── public.api
│       ├── etc/
│       └── internal/
├── common/                     # 共享模块
│   ├── go.mod
│   ├── dao/
│   ├── model/
│   ├── constants/
│   ├── client/
│   ├── errors/
│   └── utils/
├── go.work                     # Go工作区配置
└── PROJECT-STATUS.md           # 本文档
```

## 🚀 下一步开发计划 (Next Development Plan)

### **阶段一: 核心功能 (MVP - Minimum Viable Product)**

#### **1.1: 基础模块和环境配置 (P1)**
- [ ] `(#T001)` [P1][Config] `admin-api`: 在 `admin-api.yaml` 中配置数据库和Redis连接信息。
- [ ] `(#T002)` [P1][Config] `public-api`: 在 `public-api.yaml` 中配置数据库和Redis连接信息。
- [ ] `(#T003)` [P1][common] `[DB]`：在 `common` 模块中实现MongoDB和Redis的客户端初始化逻辑。
- [ ] `(#T004)` [P1][common] `[Utils]`：在 `common/utils` 中实现密码加密/验证工具 (`password.go`)。
- [ ] `(#T005)` [P1][common] `[Constants]`：创建 `common/constants` 目录，用于存放业务常量。

#### **1.2: 用户与认证模块 (P1)**
- [ ] `(#T010)` [P1][common] `[Model]`：在 `common/model/user.go` 中定义 `User` 数据模型。
- [ ] `(#T011)` [P1][common] `[Model]`：在 `common/model/login_log.go` 中定义 `LoginLog` 数据模型。
- [ ] `(#T012)` [P1][common] `[Constants]`：在 `common/constants/user.go` 中定义用户角色和状态常量。
- [ ] `(#T013)` [P1][common] `[DAO]`：在 `common/dao/user_dao.go` 中实现 `UserDAO` 的 `Create`, `GetByID`, `GetByUsername`, `Update`, `Delete`, `List` 方法。
- [ ] `(#T014)` [P1][common] `[Test]`：为 `UserDAO` 编写单元测试，确保其所有方法正确。
- [ ] `(#T015)` [P1][common] `[DAO]`：在 `common/dao/login_log_dao.go` 中实现 `LoginLogDAO` 的 `Create`, `List` 方法。
- [ ] `(#T016)` [P1][common] `[Test]`：为 `LoginLogDAO` 编写单元测试。
- [ ] `(#T017)` [P1][admin-api] `[Config]`：在 `admin-api.yaml` 中配置JWT密钥和过期时间。
- [ ] `(#T018)` [P1][admin-api] `[Svc]`：在 `ServiceContext` 中注入 `UserDAO` 和 `LoginLogDAO`。
- [ ] `(#T019)` [P1][admin-api] `[API]`：在 `admin.api` 中定义用户登录接口 `POST /api/v1/admin/auth/login`。
- [ ] `(#T020)` [P1][admin-api] `[Logic]`：实现 `LoginLogic` 登录逻辑。
- [ ] `(#T021)` [P1][admin-api] `[API]`：在 `admin.api` 中定义获取当前用户信息接口 `GET /api/v1/admin/auth/profile`。
- [ ] `(#T022)` [P1][admin-api] `[Logic]`：实现 `GetProfileLogic` 逻辑。
- [ ] `(#T023)` [P1][admin-api] `[API]`：在 `admin.api` 中定义用户列表接口 `GET /api/v1/admin/users`。
- [ ] `(#T024)` [P1][admin-api] `[Logic]`：实现 `GetUserListLogic` 逻辑。
- [ ] `(#T025)` [P1][admin-api] `[API]`：在 `admin.api` 中定义登录日志接口 `GET /api/v1/admin/security/login-logs`。
- [ ] `(#T026)` [P1][admin-api] `[Logic]`：实现 `GetLoginLogsLogic` 逻辑。

#### **1.3: 内容管理模块 - 文章 (P1)**
- [ ] `(#T030)` [P1][common] `[Model]`：在 `common/model/post.go` 中定义 `Post` 数据模型 (包含内嵌的`Tag`结构)。
- [ ] `(#T031)` [P1][common] `[Constants]`：在 `common/constants/post.go` 中定义文章状态常量。
- [ ] `(#T032)` [P1][common] `[DAO]`：在 `common/dao/post_dao.go` 中实现 `PostDAO` 的 `Create`, `GetByID`, `GetBySlug`, `Update`, `Delete`, `List` 方法。
- [ ] `(#T033)` [P1][common] `[Test]`：为 `PostDAO` 编写单元测试。
- [ ] `(#T034)` [P1][admin-api] `[Svc]`：在 `ServiceContext` 中注入 `PostDAO`。
- [ ] `(#T035)` [P1][admin-api] `[API]`：在 `admin.api` 中定义获取文章列表接口 `GET /api/v1/admin/posts`。
- [ ] `(#T036)` [P1][admin-api] `[Logic]`：实现 `GetPostListLogic` 逻辑。
- [ ] `(#T037)` [P1][admin-api] `[API]`：在 `admin.api` 中定义创建文章接口 `POST /api/v1/admin/posts`。
- [ ] `(#T038)` [P1][admin-api] `[Logic]`：实现 `CreatePostLogic` 逻辑。
- [ ] `(#T039)` [P1][admin-api] `[API]`：在 `admin.api` 中定义获取文章详情接口 `GET /api/v1/admin/posts/{id}`。
- [ ] `(#T040)` [P1][admin-api] `[Logic]`：实现 `GetPostDetailLogic` 逻辑。
- [ ] `(#T041)` [P1][admin-api] `[API]`：在 `admin.api` 中定义更新文章接口 `PUT /api/v1/admin/posts/{id}`。
- [ ] `(#T042)` [P1][admin-api] `[Logic]`：实现 `UpdatePostLogic` 逻辑。
- [ ] `(#T043)` [P1][admin-api] `[API]`：在 `admin.api` 中定义删除文章接口 `DELETE /api/v1/admin/posts/{id}`。
- [ ] `(#T044)` [P1][admin-api] `[Logic]`：实现 `DeletePostLogic` 逻辑。
- [ ] `(#T045)` [P1][public-api] `[Svc]`：在 `ServiceContext` 中注入 `PostDAO`。
- [ ] `(#T046)` [P1][public-api] `[API]`：在 `public.api` 中定义获取公开文章列表接口 `GET /api/v1/public/posts`。
- [ ] `(#T047)` [P1][public-api] `[Logic]`：实现 `GetPublicPostListLogic` 逻辑。
- [ ] `(#T048)` [P1][public-api] `[API]`：在 `public.api` 中定义获取公开文章详情接口 `GET /api/v1/public/posts/{slug}`。
- [ ] `(#T049)` [P1][public-api] `[Logic]`：实现 `GetPublicPostDetailLogic` 逻辑。

#### **1.4: 内容管理模块 - 页面 (P1)**
- [ ] `(#T050)` [P1][common] `[Model]`：在 `common/model/page.go` 中定义 `Page` 数据模型 (结构类似Post)。
- [ ] `(#T051)` [P1][common] `[DAO]`：在 `common/dao/page_dao.go` 中实现 `PageDAO`。
- [ ] `(#T052)` [P1][common] `[Test]`：为 `PageDAO` 编写单元测试。
- [ ] `(#T053)` [P1][admin-api] `[Svc]`：在 `ServiceContext` 中注入 `PageDAO`。
- [ ] `(#T054)` [P1][admin-api] `[API]`：在 `admin.api` 中定义获取页面列表接口。
- [ ] `(#T055)` [P1][admin-api] `[Logic]`：实现获取页面列表逻辑。
- [ ] `(#T056)` [P1][admin-api] `[API]`：在 `admin.api` 中定义页面CRUD的其他接口 (创建/获取详情/更新/删除)。
- [ ] `(#T057)` [P1][admin-api] `[Logic]`：实现页面CRUD的其他逻辑。
- [ ] `(#T058)` [P1][public-api] `[Svc]`：在 `ServiceContext` 中注入 `PageDAO`。
- [ ] `(#T059)` [P1][public-api] `[API]`：在 `public.api` 中定义获取公开页面详情接口 `GET /api/v1/public/pages/{slug}`。
- [ ] `(#T060)` [P1][public-api] `[Logic]`：实现 `GetPublicPageDetailLogic` 逻辑。

#### **1.5: 文章发布管理 (P1)**
- [ ] `(#T063)` [P1][admin-api] `[API]`：在 `admin.api` 中定义发布文章接口 `POST /api/v1/admin/posts/{id}/publish`。
- [ ] `(#T064)` [P1][admin-api] `[Logic]`：实现 `PublishPostLogic` 逻辑 (更新`status`和`publishedAt`字段)。
- [ ] `(#T065)` [P1][admin-api] `[API]`：在 `admin.api` 中定义撤销发布接口 `POST /api/v1/admin/posts/{id}/unpublish`。
- [ ] `(#T066)` [P1][admin-api] `[Logic]`：实现 `UnpublishPostLogic` 逻辑。

---

### **阶段二: 高级功能与完善**

#### **2.1: 认证与用户管理完善 (P2)**
- [ ] `(#T100)` [P2][common] `[Cache]`：在 `common/cache` 中实现基于Redis的JWT黑名单缓存服务。
- [ ] `(#T101)` [P2][admin-api] `[API]`：在 `admin.api` 中定义用户登出接口 `POST /api/v1/admin/auth/logout`。
- [ ] `(#T102)` [P2][admin-api] `[Logic]`：实现 `LogoutLogic` 逻辑，将Token加入黑名单。
- [ ] `(#T103)` [P2][admin-api] `[Middleware]`：实现JWT操作黑名单检查的中间件。
- [ ] `(#T104)` [P2][admin-api] `[API]`：在 `admin.api` 中定义刷新令牌接口 `POST /api/v1/admin/auth/refresh`。
- [ ] `(#T105)` [P2][admin-api] `[Logic]`：实现 `RefreshTokenLogic` 逻辑。
- [ ] `(#T106)` [P2][admin-api] `[API]`：在 `admin.api` 中定义更新个人资料接口 `PUT /api/v1/admin/auth/profile`。
- [ ] `(#T107)` [P2][admin-api] `[Logic]`：实现 `UpdateProfileLogic` 逻辑。
- [ ] `(#T108)` [P2][admin-api] `[API]`：在 `admin.api` 中定义修改密码接口 `PUT /api/v1/admin/auth/password`。
- [ ] `(#T109)` [P2][admin-api] `[Logic]`：实现 `ChangePasswordLogic` 逻辑。
- [ ] `(#T110)` [P2][admin-api] `[API]`：在 `admin.api` 中定义创建用户接口 `POST /api/v1/admin/users`。
- [ ] `(#T111)` [P2][admin-api] `[Logic]`：实现 `CreateUserLogic` 逻辑。
- [ ] `(#T112)` [P2][admin-api] `[API]`：在 `admin.api` 中定义更新用户接口 `PUT /api/v1/admin/users/{id}`。
- [ ] `(#T113)` [P2][admin-api] `[Logic]`：实现 `UpdateUserLogic` 逻辑。
- [ ] `(#T114)` [P2][admin-api] `[API]`：在 `admin.api` 中定义删除用户接口 `DELETE /api/v1/admin/users/{id}`。
- [ ] `(#T115)` [P2][admin-api] `[Logic]`：实现 `DeleteUserLogic` 逻辑。

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
- [ ] `(#T300)` [P2][admin-api] `[API]` & `[Logic]`：实现数据统计(Dashboard)接口。
- [ ] `(#T301)` [P2][public-api] `[API]` & `[Logic]`：实现搜索接口 `GET /api/v1/public/search`。
- [ ] `(#T302)` [P2][public-api] `[API]` & `[Logic]`：实现 `sitemap.xml` 和 `rss.xml` 的生成接口。
- [ ] `(#T310)` [P2][Project] `[CI/CD]`：配置CI/CD流水线 (例如 GitHub Actions)。
- [ ] `(#T311)` [P2][Project] `[Container]`：为服务编写 `Dockerfile` 并配置 `docker-compose`。
- [ ] `(#T312)` [P2][Project] `[Makefile]`：创建 `Makefile` 简化常用操作 (build, test, run)。

## 🔧 **技术债务和待优化项**

- [ ] 添加项目级别的Makefile
- [ ] 配置Docker开发环境
- [ ] 设置CI/CD流程
- [ ] 配置代码审查自动化工具 (golangci-lint, gosec)
- [ ] 设置GitHub PR模板和审查规则
- [ ] 完善错误处理机制
- [ ] 添加日志配置

## 📚 **开发指南**

开发任何新功能前，请务必：

1. **阅读规范**: 查阅 `docs/` 目录下的所有规范文档
2. **遵循TDD**: 先写测试，再写实现
3. **API优先**: 从定义 `.api` 文件开始
4. **服务判断**: 确定功能属于哪个服务 (admin vs public vs common)
5. **代码审查**: 确保符合所有既定规范 