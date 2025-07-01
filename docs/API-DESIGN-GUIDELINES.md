# API 设计规范

本规范旨在为项目提供一套统一的、符合业界最佳实践的 RESTful API 设计标准。所有由 `go-zero` 的 `.api` 文件定义的接口都应遵循此规范。

## 1. RESTful 设计原则

我们遵循 REST (Representational State Transfer) 的核心原则来设计我们的 API。

- **面向资源 (Resource-Oriented)**: API 的核心是"资源"。每个 URL 都代表一种资源。
  - **示例**: `/users`, `/posts`, `/posts/{postId}/comments`
- **使用标准 HTTP 方法**: 使用 HTTP 方法 (动词) 来描述对资源的操作。
  - `GET`: 读取资源。
  - `POST`: 创建新资源。
  - `PUT`: **完整**替换、更新一个现有资源。
  - `PATCH`: **部分**更新一个现有资源。
  - `DELETE`: 删除一个资源。
- **无状态 (Stateless)**: 服务器不应保存客户端的会话状态。每个从客户端发来的请求都应包含所有必要的信息，以便服务器能够理解和处理它。我们将通过 JWT Token 来实现这一点。

## 2. URL 结构与版本控制

- **URL 结构**:
  - **使用复数名词**: 集合资源应使用复数名词，例如 `/users`, `/posts`。
  - **路径变量**: 使用路径变量来标识单个资源，例如 `/users/{userId}`。
  - **全小写**: URL 路径全部使用小写字母。
  - **连接符**: 使用连字符 `-` 来连接路径中的单词，例如 `/blog-posts` (如果需要)。但我们优先推荐 `camelCase` 风格的单个词，如 `/blogPosts`。

- **版本控制**:
  - 我们采用在 URL 中添加版本号的方式。所有 API 路径都应以 `/api/v1` 作为前缀。
  - **示例**: `/api/v1/users`, `/api/v1/posts`
  - 这将在 `.api` 文件的 `service` 定义中通过 `@server(prefix=...)` 来实现。

## 3. 请求与响应

### 3.1. 数据格式

- 所有请求体 (Request Body) 和响应体 (Response Body) 都必须使用 **JSON** 格式。
- `Content-Type` header 应设置为 `application/json`。

### 3.2. 命名约定

- **JSON 字段**: 请求和响应体中的所有 JSON 字段名都必须使用 `camelCase` 命名法。
  - **示例**: `"userId"`, `"firstName"`, `"createdAt"`

### 3.3. 成功响应 (`2xx`)

- `GET`:
  - 查询单个资源: 返回 `200 OK` 和资源对象。
  - 查询资源集合: 返回 `200 OK` 和一个包含资源对象的数组。如果集合为空，返回一个空数组 `[]`。
- `POST`: 创建成功后，返回 `201 Created` 和新创建的资源对象。
- `PUT / PATCH`: 更新成功后，返回 `200 OK` 和更新后的完整资源对象。
- `DELETE`: 删除成功后，返回 `204 No Content`，响应体应为空。

### 3.4. 统一错误响应 (`4xx` & `5xx`)

为了提供一致的错误处理体验，所有客户端错误 (`4xx`) 和服务器错误 (`5xx`) 都应返回一个遵循相同结构的 JSON 对象。

- **错误响应结构**:
  ```json
  {
    "code": "unique_error_code",
    "msg": "A human-readable error message.",
    "details": {
      "field": "The specific field that caused the error, if applicable."
    }
  }
  ```
- **字段说明**:
  - `code`: 一个机器可读的、唯一的错误码字符串，方便客户端进行程序化处理。例如 `resource_not_found`, `validation_failed`。
  - `msg`: 一段人类可读的、清晰的错误描述信息。
  - `details` (可选): 一个包含更具体错误信息的对象。例如，在参数验证失败时，它可以指出是哪个字段不符合要求。

- **常用 HTTP 状态码**:
  - `400 Bad Request`: 请求无效。最常见的错误，例如请求参数格式错误、验证失败等。
  - `401 Unauthorized`: 未认证。请求需要用户认证，但客户端未提供有效的凭证 (Token)。
  - `403 Forbidden`: 已认证，但无权限。用户已登录，但其角色无权执行此操作。
  - `404 Not Found`: 请求的资源不存在。
  - `409 Conflict`: 资源冲突。例如，尝试创建一个用户名已存在的用户。
  - `500 Internal Server Error`: 服务器内部错误。这是一个通用的服务器端错误，表示服务器遇到了一个它不知道如何处理的意外情况。

## 4. 认证 (Authentication)

- 我们使用 **JWT (JSON Web Token)** 进行认证。
- 客户端在登录后获取 Token，并在后续所有需要认证的请求中，通过 `Authorization` HTTP header 来携带它。
- **Header 格式**: `Authorization: Bearer <your-jwt-token>`
- `go-zero` 提供了 `@server(jwt: ...)` 注解来方便地为接口开启 JWT 认证。

## 5. 分页规范

博客系统中的文章列表、评论列表等资源经常需要分页处理。我们采用基于页码的分页策略。

### 5.1. 请求参数

- **查询参数**:
  - `page`: 页码，从 1 开始（可选，默认为 1）
  - `limit`: 每页记录数（可选，默认为 10，最大为 100）

**示例**: `GET /api/v1/posts?page=2&limit=20`

### 5.2. 响应格式

```json
{
  "data": [...], // 实际的数据数组
  "pagination": {
    "page": 2,           // 当前页码
    "limit": 20,         // 每页记录数
    "total": 156,        // 总记录数
    "totalPages": 8,     // 总页数
    "hasNext": true,     // 是否有下一页
    "hasPrev": true      // 是否有上一页
  }
}
``` 