# Swagger API 文档

本目录包含Heimdall项目的完整API文档，由goctl工具从.api文件自动生成。

## 📁 文件说明

- **`admin-api.yaml`** - 管理后台API文档 (端口: 8080)
- **`public-api.yaml`** - 公开前台API文档 (端口: 8081)
- **`README.md`** - 本说明文档

## 🚀 如何查看文档

### 方法一：使用Swagger Editor (推荐)

1. 访问 [Swagger Editor](https://editor.swagger.io/)
2. 将对应的yaml文件内容复制到编辑器中
3. 即可查看格式化的API文档和在线测试接口

### 方法二：使用本地Swagger UI

```bash
# 安装swagger-ui-dist
npm install -g swagger-ui-dist

# 启动本地swagger-ui (以admin-api为例)
swagger-ui-serve -f admin-api.yaml -p 3001
```

### 方法三：使用Docker运行Swagger UI

```bash
# 运行admin-api文档
docker run -p 3001:8080 -e SWAGGER_JSON=/docs/admin-api.yaml \
  -v $(pwd):/docs swaggerapi/swagger-ui

# 运行public-api文档  
docker run -p 3002:8080 -e SWAGGER_JSON=/docs/public-api.yaml \
  -v $(pwd):/docs swaggerapi/swagger-ui
```

然后访问:
- Admin API 文档: http://localhost:3001
- Public API 文档: http://localhost:3002

## 📋 API 接口概览

### Admin API (管理后台)

**认证接口:**
- `POST /api/v1/admin/auth/login` - 用户登录
- `GET /api/v1/admin/auth/profile` - 获取当前用户信息
- `POST /api/v1/admin/auth/logout` - 用户登出

**用户管理:**
- `GET /api/v1/admin/users` - 获取用户列表 (支持分页、过滤、排序)
- `GET /api/v1/admin/users/{id}` - 获取用户详情

**安全管理:**
- `GET /api/v1/admin/security/login-logs` - 获取登录日志 (支持多维度过滤)

### Public API (公开前台)

**测试接口:**
- `GET /api/v1/public/test/{name}` - 测试接口 (临时)

> **注意**: Public API当前只有测试接口，正式的博客文章、评论等接口将在后续任务中实现。

## 🔧 重新生成文档

当.api文件发生变更时，使用以下命令重新生成swagger文档：

```bash
# 重新生成admin-api文档
goctl api swagger --api admin-api/admin/admin.api --dir docs/swagger --filename admin-api --yaml

# 重新生成public-api文档
goctl api swagger --api public-api/public/public.api --dir docs/swagger --filename public-api --yaml
```

或者使用Makefile快捷命令：

```bash
# 生成所有swagger文档
make swagger

# 只生成admin-api文档
make swagger-admin

# 只生成public-api文档  
make swagger-public
```

## 🧪 接口测试

### 使用curl测试

```bash
# 测试登录接口
curl -X POST http://localhost:8080/api/v1/admin/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123",
    "rememberMe": false
  }'

# 使用token访问受保护接口
curl -X GET http://localhost:8080/api/v1/admin/auth/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 使用Postman测试

1. 导入swagger yaml文件到Postman
2. 配置环境变量：
   - `admin_base_url`: http://localhost:8080
   - `public_base_url`: http://localhost:8081
   - `jwt_token`: (登录后获取的token)

## 📚 相关文档

- [API设计规范](../API-DESIGN-GUIDELINES.md)
- [Go-Zero开发规范](../GO-ZERO-GUIDELINES.md)  
- [项目架构设计](../../design/SYSTEM-ARCHITECTURE-AND-MODULES.md)
- [API接口规范](../../design/API-INTERFACE-SPECIFICATION.md)

## 🔄 文档更新记录

- **2025-07-04**: 初始版本，包含认证、用户管理、登录日志等接口
- **待更新**: 文章管理、评论系统、标签管理等接口将在后续任务中添加

---

**注意**: 本文档由goctl工具自动生成，请不要手动修改yaml文件。如需更新API文档，请修改对应的.api文件后重新生成。 