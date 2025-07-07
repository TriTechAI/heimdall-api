# E2E (端到端) 测试

本目录包含Heimdall API的端到端测试套件，用于验证整个系统的功能正确性。

## 📋 测试覆盖范围

### 1. 用户认证流程测试 (`auth_test.go`)
- ✅ 用户登录流程
- ✅ 获取用户信息
- ✅ Token失效测试
- ✅ 登录失败场景
- ✅ 认证安全性测试
- ✅ 登录尝试限制

### 2. 文章CRUD操作测试 (`post_test.go`)
- ✅ 创建文章
- ✅ 获取文章详情
- ✅ 获取文章列表
- ✅ 更新文章
- ✅ 发布状态管理
- ✅ 文章搜索和过滤
- ✅ 删除文章
- ✅ 文章权限验证
- ✅ 文章数据验证

### 3. 公开API访问测试 (`public_api_test.go`)
- ✅ 获取公开文章列表
- ✅ 分页功能测试
- ✅ 排序功能测试
- ✅ 过滤功能测试
- ✅ 获取文章详情
- ✅ 浏览计数功能
- ✅ 访问控制测试
- ✅ API性能和安全测试

## 🚀 运行测试

### 前置条件
1. **启动数据库服务**：
   ```bash
   # MongoDB
   mongod --dbpath /path/to/data/db
   
   # Redis
   redis-server
   ```

2. **环境配置**：
   ```bash
   # 复制环境变量配置文件
   cp test/e2e/.env.test .env.test
   
   # 根据需要修改配置
   nano .env.test
   ```

### 运行测试

#### 方式一：使用Makefile（推荐）
```bash
# 运行所有E2E测试
make test-e2e

# 运行单元测试
make test-unit

# 运行所有测试
make test
```

#### 方式二：直接运行
```bash
# 进入E2E测试目录
cd test/e2e

# 运行所有E2E测试
go test -v -timeout=10m .

# 运行特定测试
go test -v -run TestAuthFlow .
go test -v -run TestPostCRUD .
go test -v -run TestPublicAPI .
```

#### 方式三：使用独立入口
```bash
# 使用main.go运行测试
cd test/e2e
go run main.go -test.v -test.timeout=10m
```

## 🔧 配置说明

### 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `TEST_ADMIN_API_URL` | `http://localhost:8080` | 管理API地址 |
| `TEST_PUBLIC_API_URL` | `http://localhost:8081` | 公开API地址 |
| `TEST_USER_USERNAME` | `testuser` | 测试用户名 |
| `TEST_USER_PASSWORD` | `testpass123` | 测试用户密码 |
| `TEST_USER_EMAIL` | `test@example.com` | 测试用户邮箱 |
| `TEST_USER_ROLE` | `admin` | 测试用户角色 |
| `TEST_MONGO_URL` | `mongodb://localhost:27017` | MongoDB连接地址 |
| `TEST_REDIS_URL` | `redis://localhost:6379` | Redis连接地址 |
| `TEST_DATABASE` | `heimdall_test` | 测试数据库名 |

### 测试配置

测试套件会自动：
1. **启动服务**：启动admin-api和public-api服务
2. **创建测试数据**：创建测试用户和基础数据
3. **运行测试**：执行所有测试用例
4. **清理环境**：停止服务并清理测试数据

## 📊 测试结果

### 成功示例
```
✅ E2E测试环境设置完成
🧪 开始运行E2E测试...
=== RUN   TestAuthFlow
=== RUN   TestPostCRUD
=== RUN   TestPublicAPI
--- PASS: TestAuthFlow (2.34s)
--- PASS: TestPostCRUD (5.67s)
--- PASS: TestPublicAPI (3.21s)
PASS
✅ 所有E2E测试通过
```

### 失败示例
```
❌ 测试环境设置失败: failed to connect to MongoDB
```

## 🛠️ 故障排除

### 1. 服务启动失败
**问题**：`services failed to start`
**解决**：
- 检查端口8080/8081是否被占用
- 确认配置文件路径正确
- 检查环境变量设置

### 2. 数据库连接失败
**问题**：`failed to connect to MongoDB`
**解决**：
- 确认MongoDB服务已启动
- 检查MongoDB连接URL
- 确认数据库权限设置

### 3. 测试数据问题
**问题**：`test user not found`
**解决**：
- 检查测试用户配置
- 手动清理数据库后重试
- 确认密码哈希正确

### 4. 网络连接问题
**问题**：`connection refused`
**解决**：
- 确认服务已正常启动
- 检查防火墙设置
- 增加服务启动等待时间

## 📝 添加新测试

### 1. 创建测试文件
```go
package e2e

import (
    "testing"
    . "github.com/smartystreets/goconvey/convey"
)

func TestNewFeature(t *testing.T) {
    Convey("新功能测试", t, func() {
        config := GetTestConfig()
        client := NewTestClient()
        
        // 测试逻辑
    })
}
```

### 2. 测试模式
- 使用GoConvey的BDD风格
- 遵循AAA模式（Arrange-Act-Assert）
- 包含正常和异常场景
- 验证安全性和性能

### 3. 最佳实践
- **独立性**：每个测试用例独立运行
- **清理**：测试后清理创建的数据
- **断言**：使用明确的断言检查结果
- **日志**：添加适当的测试日志

## 🔍 调试测试

### 启用详细日志
```bash
export TEST_LOG_LEVEL=debug
go test -v -run TestAuthFlow .
```

### 保留测试数据
```bash
export TEST_CLEANUP=false
go test -v .
```

### 单独测试组件
```bash
# 只测试认证
go test -v -run TestAuth .

# 只测试文章管理
go test -v -run TestPost .

# 只测试公开API
go test -v -run TestPublic .
```

## 📈 持续集成

E2E测试可以集成到CI/CD流水线中：

```yaml
# GitHub Actions示例
- name: Run E2E Tests
  run: |
    docker-compose up -d mongodb redis
    make test-e2e
  env:
    TEST_MONGO_URL: mongodb://localhost:27017
    TEST_REDIS_URL: redis://localhost:6379
```

## 🎯 测试目标

- **功能完整性**：验证所有核心功能正常工作
- **API兼容性**：确保API接口向后兼容
- **安全性**：验证认证授权机制有效
- **性能基准**：检查响应时间在可接受范围内
- **错误处理**：验证错误场景的正确处理

通过完善的E2E测试，我们可以确保Heimdall API系统的质量和稳定性。