# Heimdall API

基于 Go-Zero 框架的高性能博客系统后端 API，采用微服务架构和统一模块管理。

## 🏗️ 项目架构

```
heimdall-api/
├── go.mod                      # 统一的模块定义文件
├── admin-api/                  # 管理服务 (端口: 8080)
│   └── admin/                  
├── public-api/                 # 公开服务 (端口: 8081)
│   └── public/                 
├── common/                     # 共享代码包
│   ├── dao/                    # 数据访问层
│   ├── model/                  # 数据模型
│   ├── constants/              # 业务常量
│   ├── client/                 # 第三方客户端
│   ├── errors/                 # 错误定义
│   └── utils/                  # 工具函数
├── design/                     # 设计文档
└── docs/                       # 开发文档
```

## 🚀 主要特性

- 🚀 **高性能**: 基于 Go-Zero 框架，支持高并发
- 🔐 **JWT 认证**: 安全的令牌认证，支持刷新令牌
- 📝 **内容管理**: 文章、页面、标签、评论的完整管理
- 👥 **用户管理**: 基于角色的访问控制 (Owner, Admin, Editor, Author)
- 🔍 **高级搜索**: MongoDB 全文搜索支持
- 📊 **数据分析**: 浏览量统计、阅读时间计算
- 🛡️ **安全防护**: 登录限制、密码加密、JWT 黑名单
- 📤 **文件上传**: 图片和文件上传，支持验证
- 🔄 **批量操作**: 文章、标签、评论的批量删除和状态更新
- 📈 **评论统计**: 实时评论数据统计和管理

## 🚀 快速开始

### 环境要求
- Go 1.19+
- MongoDB 5.0+
- Redis 6.0+

### 安装依赖
```bash
go mod tidy
```

### 启动服务

**使用Makefile (推荐)**:
```bash
# 查看所有可用命令
make help

# 构建所有服务
make build

# 启动管理服务 (端口: 8080)
make admin

# 启动公开服务 (端口: 8081)
make public

# 运行测试
make test
```

**直接使用Go命令**:
```bash
# 启动管理服务 (端口: 8080)
go run ./admin-api/admin

# 启动公开服务 (端口: 8081)  
go run ./public-api/public

# 运行测试
go test ./...
```

## 📋 服务说明

### Admin API (管理服务)
- **端口**: 8080
- **用户**: 博客管理员、编辑、作者
- **功能**: 用户管理、内容管理、评论审核、系统设置、媒体管理

### Public API (公开服务)
- **端口**: 8081  
- **用户**: 博客访问者、搜索引擎、第三方应用
- **功能**: 内容展示、内容搜索、评论系统、RSS订阅、SEO优化

### Common (共享包)
- **性质**: 被两个服务共同引用的基础代码
- **内容**: 数据模型、数据访问、缓存、安全、工具函数等

## 📚 API 端点

### Admin API 主要端点

#### 认证管理
- `POST /api/v1/admin/auth/login` - 用户登录
- `POST /api/v1/admin/auth/logout` - 用户登出
- `POST /api/v1/admin/auth/refresh` - 刷新令牌
- `GET /api/v1/admin/auth/profile` - 获取当前用户信息
- `POST /api/v1/admin/auth/change-password` - 修改密码

#### 文章管理
- `GET /api/v1/admin/posts` - 获取文章列表
- `POST /api/v1/admin/posts` - 创建文章
- `GET /api/v1/admin/posts/:id` - 获取文章详情
- `PUT /api/v1/admin/posts/:id` - 更新文章
- `DELETE /api/v1/admin/posts/:id` - 删除文章
- `POST /api/v1/admin/posts/:id/publish` - 发布文章
- `POST /api/v1/admin/posts/:id/unpublish` - 取消发布文章
- `DELETE /api/v1/admin/posts/batch` - 批量删除文章
- `PATCH /api/v1/admin/posts/batch/status` - 批量更新文章状态

#### 标签管理
- `GET /api/v1/admin/tags` - 获取标签列表
- `POST /api/v1/admin/tags` - 创建标签
- `GET /api/v1/admin/tags/:id` - 获取标签详情
- `PUT /api/v1/admin/tags/:id` - 更新标签
- `DELETE /api/v1/admin/tags/:id` - 删除标签
- `GET /api/v1/admin/tags/all` - 获取所有标签（不分页）
- `GET /api/v1/admin/tags/search` - 搜索标签
- `DELETE /api/v1/admin/tags/batch` - 批量删除标签

#### 评论管理
- `GET /api/v1/admin/comments` - 获取评论列表
- `GET /api/v1/admin/comments/:id` - 获取评论详情
- `PUT /api/v1/admin/comments/:id` - 更新评论
- `DELETE /api/v1/admin/comments/:id` - 删除评论
- `PUT /api/v1/admin/comments/:id/approve` - 审核通过评论
- `PUT /api/v1/admin/comments/:id/reject` - 拒绝评论
- `PUT /api/v1/admin/comments/:id/spam` - 标记垃圾评论
- `POST /api/v1/admin/comments/:id/reply` - 回复评论
- `GET /api/v1/admin/comments/stats` - 获取评论统计
- `PATCH /api/v1/admin/comments/batch/status` - 批量更新评论状态

#### 文件上传
- `POST /api/v1/admin/upload/image` - 上传图片
- `POST /api/v1/admin/upload/file` - 上传文件

### Public API 主要端点

- `GET /api/v1/public/posts` - 获取公开文章列表
- `GET /api/v1/public/posts/:slug` - 通过 slug 获取文章详情
- `GET /api/v1/public/tags` - 获取标签列表
- `GET /api/v1/public/pages/:slug` - 通过 slug 获取页面详情

## 📚 文档

### 🔗 API文档
- [📖 Swagger API 文档](./docs/swagger/) - 完整的REST API接口文档
- [🌐 在线浏览](./docs/swagger/index.html) - 可视化API文档界面

### 📋 设计文档  
- [系统架构设计](./design/SYSTEM-ARCHITECTURE-AND-MODULES.md)
- [微服务架构规范](./docs/MULTI-SERVICE-ARCHITECTURE.md)
- [API 接口规范](./design/API-INTERFACE-SPECIFICATION.md)
- [数据模型设计](./design/DATA-MODEL-DESIGN.md)
- [安全设计](./design/SECURITY-DESIGN.md)

### 🔧 生成API文档
```bash
# 生成所有Swagger文档
make swagger

# 只生成管理后台API文档
make swagger-admin

# 只生成公开API文档
make swagger-public
```

## 🛠️ 技术栈

- **后端框架**: Go-Zero
- **数据库**: MongoDB
- **缓存**: Redis  
- **认证**: JWT
- **测试**: GoConvey + Mockey

## 📄 许可证

本项目采用 [MIT](./LICENSE) 许可证。