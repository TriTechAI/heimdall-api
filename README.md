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

## 🚀 快速开始

### 环境要求
- Go 1.24.4+
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

## 📚 文档

- [系统架构设计](./design/SYSTEM-ARCHITECTURE-AND-MODULES.md)
- [微服务架构规范](./docs/MULTI-SERVICE-ARCHITECTURE.md)
- [API 接口规范](./design/API-INTERFACE-SPECIFICATION.md)
- [数据模型设计](./design/DATA-MODEL-DESIGN.md)
- [安全设计](./design/SECURITY-DESIGN.md)

## 🛠️ 技术栈

- **后端框架**: Go-Zero
- **数据库**: MongoDB
- **缓存**: Redis  
- **认证**: JWT
- **测试**: GoConvey + Mockey

## 📄 许可证

本项目采用 [MIT](./LICENSE) 许可证。