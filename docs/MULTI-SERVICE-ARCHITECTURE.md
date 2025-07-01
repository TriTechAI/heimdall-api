# 微服务架构规范 (Multi-Service Architecture Guidelines)

本文档定义了 `heimdall` 项目的微服务架构设计原则、服务划分标准和开发协作规范。

## 1. 架构决策理由 (Architecture Decision Record)

### 1.1. 为什么选择微服务架构？

对于博客系统这种天然具有"前后台分离"特性的应用，拆分服务具有以下关键优势：

- **🔒 安全隔离**: 管理后台可部署在内网，公开API暴露在公网，物理隔离降低安全风险
- **🚀 独立部署**: 更新管理功能不影响博客正常访问，提升系统稳定性  
- **📈 精准扩容**: 公开API(读多)和管理API(写少)有不同资源需求，可独立扩容
- **🛠️ 故障隔离**: 单个服务异常不会影响其他服务运行
- **👥 团队协作**: 不同团队可以独立开发和维护不同的服务

### 1.2. 服务拆分原则

我们按照**业务功能**和**用户群体**进行服务拆分：

- **用户群体**: 管理人员 vs 公众访客
- **业务性质**: 内容管理 vs 内容展示  
- **安全级别**: 高权限操作 vs 公开访问
- **访问模式**: 低频写操作 vs 高频读操作

## 2. 服务详细定义

### 2.1. Admin API 服务

**服务名**: `admin-api`  
**端口**: `8080` (可配置)  
**目标用户**: 博客管理员、编辑、作者

**核心功能**:
- **用户管理**: 用户注册、登录、角色管理、权限控制
- **内容管理**: 文章CRUD、草稿管理、发布/撤回、批量操作
- **评论管理**: 评论审核、删除、垃圾评论过滤
- **系统设置**: 博客配置、主题设置、SEO配置
- **数据统计**: 访问量统计、内容分析、用户行为分析
- **媒体管理**: 图片上传、文件管理、存储配置

**安全要求**:
- 强制JWT认证
- 角色权限验证  
- 操作日志记录
- 建议内网部署或VPN访问

### 2.2. Public API 服务

**服务名**: `public-api`  
**端口**: `8081` (可配置)  
**目标用户**: 博客访问者、搜索引擎、第三方应用

**核心功能**:
- **内容展示**: 文章列表、文章详情、分页浏览
- **分类浏览**: 按标签、分类、作者、时间过滤  
- **搜索功能**: 全文搜索、高级搜索、搜索结果排序
- **评论功能**: 评论展示、评论提交(游客+注册用户)
- **RSS/Feed**: RSS订阅、Atom feed、Sitemap生成
- **SEO优化**: meta标签、结构化数据、友好URL

**性能要求**:
- 支持高并发访问
- 缓存友好设计
- CDN兼容
- 响应时间优化

### 2.3. Common 共享模块

**模块名**: `common`  
**性质**: Go模块，被两个服务引用

**包含内容**:
- **数据模型** (`model/`): 数据库实体定义
- **数据访问** (`dao/`): 数据库操作封装
- **业务常量** (`constants/`): 状态、角色、配置常量
- **第三方客户端** (`client/`): 邮件、存储、支付等外部服务
- **错误定义** (`errors/`): 统一的业务错误类型
- **工具函数** (`utils/`): 通用辅助函数

## 3. Go Workspace 管理

### 3.1. Workspace 结构

```
heimdall-api/
├── go.work                     # Workspace 配置文件
├── admin-api/                  # 管理服务
│   ├── go.mod                  # 模块定义
│   └── admin/                  # goctl生成的服务代码
│       ├── admin.go            # 服务入口
│       ├── admin.api           # API定义
│       ├── etc/                # 配置文件
│       └── internal/           # 内部代码
├── public-api/                 # 公开服务  
│   ├── go.mod                  # 模块定义
│   └── public/                 # goctl生成的服务代码
│       ├── public.go           # 服务入口
│       ├── public.api          # API定义
│       ├── etc/                # 配置文件
│       └── internal/           # 内部代码
└── common/                     # 共享模块
    ├── go.mod                  # 模块定义
    ├── dao/                    # 数据访问层
    ├── model/                  # 数据模型
    ├── constants/              # 业务常量
    ├── client/                 # 第三方客户端
    ├── errors/                 # 错误定义
    └── utils/                  # 工具函数
```

### 3.2. 模块依赖关系

```
admin-api  ──┐
              ├──► common
public-api ──┘
```

- `admin-api` 和 `public-api` 都依赖 `common`
- 两个API服务之间**无直接依赖**
- `common` 模块**不依赖**任何API服务

### 3.3. 包导入规范

**在 admin-api 中**:
```go
import (
    "github.com/heimdall-api/common/model"
    "github.com/heimdall-api/common/dao"
    "github.com/heimdall-api/common/constants"
    
    "github.com/heimdall-api/admin-api/admin/internal/svc"
    "github.com/heimdall-api/admin-api/admin/internal/types"
)
```

**在 public-api 中**:
```go
import (
    "github.com/heimdall-api/common/model"
    "github.com/heimdall-api/common/dao"
    "github.com/heimdall-api/common/constants"
    
    "github.com/heimdall-api/public-api/public/internal/svc"
    "github.com/heimdall-api/public-api/public/internal/types"
)
```

## 4. 开发协作规范

### 4.1. 功能开发流程

1. **需求分析**: 确定功能属于哪个服务
2. **API设计**: 在对应的 `.api` 文件中定义接口
3. **共享代码**: 先完善 `common` 模块的相关代码
4. **服务实现**: 在具体服务中实现业务逻辑
5. **测试验证**: 编写单元测试和集成测试
6. **部署验证**: 独立验证每个服务的功能

### 4.2. 代码审查要点

- **服务边界**: 确保功能放在正确的服务中
- **依赖方向**: 检查模块依赖是否符合架构设计
- **共享合理性**: 评估代码是否适合放在 `common` 模块
- **接口设计**: 确保API设计符合RESTful规范
- **安全考虑**: 验证权限控制和数据验证

### 4.3. 部署策略

**开发环境**:
- 两个服务可以在同一台机器上运行
- 使用不同端口进行区分
- 共享同一个数据库实例

**生产环境**:
- `admin-api` 部署在内网或通过VPN访问
- `public-api` 部署在公网，前置负载均衡器
- 可以独立扩容和版本更新
- 使用相同的数据库集群但不同的连接池配置

## 5. 监控和运维

### 5.1. 日志规范

- 每个服务维护独立的日志
- 使用统一的日志格式和级别
- 包含服务名标识便于日志聚合
- 敏感信息脱敏处理

### 5.2. 健康检查

- 每个服务提供 `/health` 端点
- 检查数据库连接状态
- 检查关键依赖服务状态
- 支持优雅关闭和重启

### 5.3. 性能监控

- 接口响应时间监控
- 数据库查询性能监控  
- 内存和CPU使用率监控
- 错误率和可用性监控 