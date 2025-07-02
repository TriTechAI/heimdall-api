# API 接口规范 (API Interface Specification)

本文档详细定义了 Heimdall 博客系统的完整API接口规范，包括C端（Public API）和B端（Admin API）的所有接口、参数、响应格式等。

## 1. 接口概览

### 1.1. 服务地址

- **Admin API (B端)**: `https://api.example.com/api/v1/admin`
- **Public API (C端)**: `https://api.example.com/api/v1/public`

### 1.2. 通用规范

#### 请求格式
- **协议**: HTTPS
- **方法**: GET, POST, PUT, DELETE
- **编码**: UTF-8
- **Content-Type**: `application/json`

#### 认证方式
- **Admin API**: Bearer Token (JWT)
- **Public API**: 无需认证（评论提交除外）

#### 响应格式
```json
{
  "code": 200,
  "message": "Success",
  "data": {...},
  "timestamp": "2024-01-01T12:00:00Z"
}
```

#### 错误格式
```json
{
  "code": "validation_failed",
  "msg": "请求参数验证失败",
  "details": {
    "field": "username",
    "reason": "用户名不能为空"
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

#### 分页格式
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "list": [...],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 100,
      "totalPages": 10,
      "hasNext": true,
      "hasPrev": false
    }
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

---

## 2. Admin API (B端接口)

### 2.1. 用户与权限模块

#### 2.1.1. 用户登录
```http
POST /api/v1/admin/auth/login
```

**请求参数:**
```json
{
  "username": "john-doe",
  "password": "password123",
  "rememberMe": false
}
```

**响应示例:**
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIs...",
    "expiresIn": 7200,
    "user": {
      "id": "60f7b1c9e1d2c3d4e5f6g7h8",
      "username": "john-doe",
      "displayName": "John Doe",
      "email": "john@example.com",
      "role": "admin",
      "profileImage": "https://cdn.example.com/avatar.jpg"
    }
  }
}
```

#### 2.1.2. 刷新令牌
```http
POST /api/v1/admin/auth/refresh
```

**请求参数:**
```json
{
  "refreshToken": "eyJhbGciOiJIUzI1NiIs..."
}
```

#### 2.1.3. 用户登出
```http
POST /api/v1/admin/auth/logout
```

**请求头:**
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

#### 2.1.4. 获取当前用户信息
```http
GET /api/v1/admin/auth/profile
```

#### 2.1.5. 更新个人资料
```http
PUT /api/v1/admin/auth/profile
```

**请求参数:**
```json
{
  "displayName": "John Doe Updated",
  "bio": "A passionate writer and developer",
  "location": "Beijing, China",
  "website": "https://johndoe.com",
  "twitter": "johndoe",
  "facebook": "john.doe"
}
```

#### 2.1.6. 修改密码
```http
PUT /api/v1/admin/auth/password
```

**请求参数:**
```json
{
  "currentPassword": "oldpassword",
  "newPassword": "newpassword123",
  "confirmPassword": "newpassword123"
}
```

#### 2.1.7. 获取用户列表
```http
GET /api/v1/admin/users?page=1&limit=10&role=admin&status=active
```

**响应示例:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "list": [
      {
        "id": "60f7b1c9e1d2c3d4e5f6g7h8",
        "username": "john-doe",
        "displayName": "John Doe",
        "email": "john@example.com",
        "role": "admin",
        "status": "active",
        "lastLoginAt": "2024-01-01T12:00:00Z",
        "createdAt": "2023-12-01T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 25,
      "totalPages": 3,
      "hasNext": true,
      "hasPrev": false
    }
  }
}
```

#### 2.1.8. 创建用户
```http
POST /api/v1/admin/users
```

**请求参数:**
```json
{
  "username": "new-user",
  "email": "newuser@example.com",
  "displayName": "New User",
  "role": "author",
  "password": "temppassword123"
}
```

#### 2.1.9. 更新用户信息
```http
PUT /api/v1/admin/users/{id}
```

#### 2.1.10. 删除用户
```http
DELETE /api/v1/admin/users/{id}
```

#### 2.1.11. 获取登录日志
```http
GET /api/v1/admin/security/login-logs?page=1&limit=20&userId={userId}&success=true
```

**响应示例:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "list": [
      {
        "id": "60f7b1c9e1d2c3d4e5f6g7h8",
        "userId": "60f7b1c9e1d2c3d4e5f6g7h8",
        "username": "john-doe",
        "ipAddress": "192.168.1.1",
        "userAgent": "Mozilla/5.0...",
        "success": true,
        "createdAt": "2024-01-01T12:00:00Z"
      }
    ],
    "pagination": {...}
  }
}
```

### 2.2. 内容管理模块

#### 2.2.1. 获取文章列表
```http
GET /api/v1/admin/posts?page=1&limit=10&status=published&authorId={authorId}&tag={tagSlug}
```

**响应示例:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "list": [
      {
        "id": "60f7b1c9e1d2c3d4e5f6g7h8",
        "title": "My First Post",
        "slug": "my-first-post",
        "excerpt": "This is a brief introduction...",
        "featuredImage": "https://cdn.example.com/image.jpg",
        "status": "published",
        "type": "post",
        "author": {
          "id": "60f7b1c9e1d2c3d4e5f6g7h8",
          "displayName": "John Doe",
          "profileImage": "https://cdn.example.com/avatar.jpg"
        },
        "tags": [
          {"name": "Go", "slug": "go"},
          {"name": "技术", "slug": "tech"}
        ],
        "viewCount": 1024,
        "publishedAt": "2024-01-01T10:00:00Z",
        "createdAt": "2023-12-30T15:00:00Z",
        "updatedAt": "2024-01-01T09:00:00Z"
      }
    ],
    "pagination": {...}
  }
}
```

#### 2.2.2. 获取文章详情
```http
GET /api/v1/admin/posts/{id}
```

**响应示例:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "id": "60f7b1c9e1d2c3d4e5f6g7h8",
    "title": "My First Post",
    "slug": "my-first-post",
    "excerpt": "This is a brief introduction...",
    "markdown": "## Title\n\nContent...",
    "html": "<h2>Title</h2><p>Content...</p>",
    "featuredImage": "https://cdn.example.com/image.jpg",
    "type": "post",
    "status": "published",
    "visibility": "public",
    "authorId": "60f7b1c9e1d2c3d4e5f6g7h8",
    "tags": [
      {"name": "Go", "slug": "go"},
      {"name": "技术", "slug": "tech"}
    ],
    "metaTitle": "Learn Go Programming - Blog",
    "metaDescription": "A comprehensive guide to...",
    "canonicalUrl": "https://example.com/canonical",
    "readingTime": 5,
    "wordCount": 1250,
    "viewCount": 1024,
    "publishedAt": "2024-01-01T10:00:00Z",
    "createdAt": "2023-12-30T15:00:00Z",
    "updatedAt": "2024-01-01T09:00:00Z"
  }
}
```

#### 2.2.3. 创建文章
```http
POST /api/v1/admin/posts
```

**请求参数:**
```json
{
  "title": "New Post Title",
  "slug": "new-post-title",
  "excerpt": "Post excerpt...",
  "markdown": "## Content\n\nPost content in markdown...",
  "featuredImage": "https://cdn.example.com/image.jpg",
  "type": "post",
  "status": "draft",
  "visibility": "public",
  "tags": [
    {"name": "Go", "slug": "go"},
    {"name": "编程", "slug": "programming"}
  ],
  "metaTitle": "SEO Title",
  "metaDescription": "SEO Description",
  "publishedAt": "2024-01-01T10:00:00Z"
}
```

#### 2.2.4. 更新文章
```http
PUT /api/v1/admin/posts/{id}
```

#### 2.2.5. 删除文章
```http
DELETE /api/v1/admin/posts/{id}
```

#### 2.2.6. 发布文章
```http
POST /api/v1/admin/posts/{id}/publish
```

**请求参数:**
```json
{
  "publishedAt": "2024-01-01T10:00:00Z"
}
```

#### 2.2.7. 撤销发布
```http
POST /api/v1/admin/posts/{id}/unpublish
```

### 2.3. 评论管理模块

#### 2.3.1. 获取评论列表
```http
GET /api/v1/admin/comments?page=1&limit=20&status=pending&postId={postId}
```

**响应示例:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "list": [
      {
        "id": "60f7b1c9e1d2c3d4e5f6g7h8",
        "postId": "60f7b1c9e1d2c3d4e5f6g7h8",
        "postTitle": "My First Post",
        "authorName": "John Visitor",
        "authorEmail": "visitor@example.com",
        "content": "Great post! Very informative.",
        "status": "pending",
        "ipAddress": "192.168.1.100",
        "likeCount": 5,
        "parentId": null,
        "createdAt": "2024-01-01T14:30:00Z"
      }
    ],
    "pagination": {...}
  }
}
```

#### 2.3.2. 审核评论
```http
PUT /api/v1/admin/comments/{id}/approve
```

#### 2.3.3. 拒绝评论
```http
PUT /api/v1/admin/comments/{id}/reject
```

#### 2.3.4. 标记为垃圾评论
```http
PUT /api/v1/admin/comments/{id}/spam
```

#### 2.3.5. 删除评论
```http
DELETE /api/v1/admin/comments/{id}
```

#### 2.3.6. 批量操作评论
```http
POST /api/v1/admin/comments/batch
```

**请求参数:**
```json
{
  "action": "approve",
  "commentIds": [
    "60f7b1c9e1d2c3d4e5f6g7h8",
    "60f7b1c9e1d2c3d4e5f6g7h9"
  ]
}
```

### 2.4. 媒体管理模块

#### 2.4.1. 获取媒体文件列表
```http
GET /api/v1/admin/media?page=1&limit=12&type=image&uploaderId={uploaderId}
```

**响应示例:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "list": [
      {
        "id": "60f7b1c9e1d2c3d4e5f6g7h8",
        "filename": "sunset.jpg",
        "url": "https://cdn.example.com/media/sunset.jpg",
        "type": "image",
        "mimeType": "image/jpeg",
        "size": 2048576,
        "width": 1920,
        "height": 1080,
        "alt": "Beautiful sunset",
        "title": "Sunset Photo",
        "uploader": {
          "id": "60f7b1c9e1d2c3d4e5f6g7h8",
          "displayName": "John Doe"
        },
        "createdAt": "2024-01-01T09:00:00Z"
      }
    ],
    "pagination": {...}
  }
}
```

#### 2.4.2. 上传媒体文件
```http
POST /api/v1/admin/media/upload
Content-Type: multipart/form-data
```

**请求参数:**
```
file: [文件]
alt: "Alternative text"
title: "File title"
```

**响应示例:**
```json
{
  "code": 200,
  "message": "上传成功",
  "data": {
    "id": "60f7b1c9e1d2c3d4e5f6g7h8",
    "filename": "uploaded-image.jpg",
    "url": "https://cdn.example.com/media/uploaded-image.jpg",
    "type": "image",
    "mimeType": "image/jpeg",
    "size": 1024768,
    "width": 1280,
    "height": 720
  }
}
```

#### 2.4.3. 更新媒体信息
```http
PUT /api/v1/admin/media/{id}
```

**请求参数:**
```json
{
  "alt": "Updated alternative text",
  "title": "Updated title"
}
```

#### 2.4.4. 删除媒体文件
```http
DELETE /api/v1/admin/media/{id}
```

### 2.5. 站点设置模块

#### 2.5.1. 获取站点设置
```http
GET /api/v1/admin/settings
```

**响应示例:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "general": {
      "title": "My Awesome Blog",
      "description": "A blog about technology and life",
      "logo": "https://cdn.example.com/logo.png",
      "timezone": "Asia/Shanghai",
      "language": "zh-CN"
    },
    "design": {
      "theme": "default",
      "accentColor": "#007d9c",
      "postsPerPage": 10
    },
    "social": {
      "twitter": "myblog",
      "facebook": "myblog",
      "github": "myblog"
    }
  }
}
```

#### 2.5.2. 更新站点设置
```http
PUT /api/v1/admin/settings
```

**请求参数:**
```json
{
  "general": {
    "title": "Updated Blog Title",
    "description": "Updated description"
  },
  "design": {
    "accentColor": "#ff6b6b"
  }
}
```

### 2.6. 导航菜单管理

#### 2.6.1. 获取导航菜单
```http
GET /api/v1/admin/navigation
```

**响应示例:**
```json
{
  "code": 200,
  "message": "Success",
  "data": [
    {
      "id": "60f7b1c9e1d2c3d4e5f6g7h8",
      "label": "首页",
      "url": "/",
      "order": 1,
      "target": "_self",
      "isActive": true
    },
    {
      "id": "60f7b1c9e1d2c3d4e5f6g7h9",
      "label": "关于",
      "url": "/about",
      "order": 2,
      "target": "_self",
      "isActive": true
    }
  ]
}
```

#### 2.6.2. 创建导航菜单项
```http
POST /api/v1/admin/navigation
```

**请求参数:**
```json
{
  "label": "新菜单",
  "url": "/new-page",
  "order": 3,
  "target": "_self",
  "isActive": true
}
```

#### 2.6.3. 更新导航菜单项
```http
PUT /api/v1/admin/navigation/{id}
```

#### 2.6.4. 删除导航菜单项
```http
DELETE /api/v1/admin/navigation/{id}
```

#### 2.6.5. 调整菜单顺序
```http
PUT /api/v1/admin/navigation/reorder
```

**请求参数:**
```json
{
  "items": [
    {"id": "60f7b1c9e1d2c3d4e5f6g7h8", "order": 1},
    {"id": "60f7b1c9e1d2c3d4e5f6g7h9", "order": 2}
  ]
}
```

### 2.7. 标签管理模块

#### 2.7.1. 获取标签列表
```http
GET /api/v1/admin/tags?page=1&limit=20&visibility=public
```

**响应示例:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "list": [
      {
        "id": "60f7b1c9e1d2c3d4e5f6g7h8",
        "name": "Go 语言",
        "slug": "go-language",
        "description": "Go编程语言相关文章",
        "color": "#007d9c",
        "featuredImage": "https://cdn.example.com/tag-go.png",
        "postCount": 25,
        "visibility": "public",
        "createdAt": "2023-12-01T10:00:00Z"
      }
    ],
    "pagination": {...}
  }
}
```

#### 2.7.2. 创建标签
```http
POST /api/v1/admin/tags
```

**请求参数:**
```json
{
  "name": "新标签",
  "slug": "new-tag",
  "description": "标签描述",
  "color": "#ff6b6b",
  "featuredImage": "https://cdn.example.com/tag.png",
  "metaTitle": "SEO标题",
  "metaDescription": "SEO描述",
  "visibility": "public"
}
```

#### 2.7.3. 更新标签
```http
PUT /api/v1/admin/tags/{id}
```

#### 2.7.4. 删除标签
```http
DELETE /api/v1/admin/tags/{id}
```

### 2.8. 数据统计模块

#### 2.8.1. 获取仪表盘统计
```http
GET /api/v1/admin/analytics/dashboard
```

**响应示例:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "overview": {
      "totalPosts": 156,
      "totalComments": 423,
      "totalUsers": 8,
      "totalViews": 15420
    },
    "recentPosts": [
      {
        "id": "60f7b1c9e1d2c3d4e5f6g7h8",
        "title": "Latest Post",
        "viewCount": 150,
        "publishedAt": "2024-01-01T10:00:00Z"
      }
    ],
    "recentComments": [
      {
        "id": "60f7b1c9e1d2c3d4e5f6g7h8",
        "content": "Great post!",
        "authorName": "John",
        "postTitle": "My Post",
        "createdAt": "2024-01-01T14:30:00Z"
      }
    ],
    "poplarPosts": [
      {
        "id": "60f7b1c9e1d2c3d4e5f6g7h8",
        "title": "Popular Post",
        "viewCount": 2150,
        "publishedAt": "2023-12-15T10:00:00Z"
      }
    ]
  }
}
```

#### 2.8.2. 获取内容统计
```http
GET /api/v1/admin/analytics/content?startDate=2024-01-01&endDate=2024-01-31
```

#### 2.8.3. 获取用户活动统计
```http
GET /api/v1/admin/analytics/users?period=30d
```

---

## 3. Public API (C端接口)

### 3.1. 内容展示模块

#### 3.1.1. 获取文章列表
```http
GET /api/v1/public/posts?page=1&limit=10&tag={tagSlug}&author={authorSlug}
```

**响应示例:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "list": [
      {
        "id": "60f7b1c9e1d2c3d4e5f6g7h8",
        "title": "My First Post",
        "slug": "my-first-post",
        "excerpt": "This is a brief introduction...",
        "featuredImage": "https://cdn.example.com/image.jpg",
        "author": {
          "id": "60f7b1c9e1d2c3d4e5f6g7h8",
          "username": "john-doe",
          "displayName": "John Doe",
          "profileImage": "https://cdn.example.com/avatar.jpg",
          "bio": "A passionate writer"
        },
        "tags": [
          {"name": "Go", "slug": "go"},
          {"name": "技术", "slug": "tech"}
        ],
        "readingTime": 5,
        "viewCount": 1024,
        "publishedAt": "2024-01-01T10:00:00Z"
      }
    ],
    "pagination": {...}
  }
}
```

#### 3.1.2. 获取文章详情
```http
GET /api/v1/public/posts/{slug}
```

**响应示例:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "id": "60f7b1c9e1d2c3d4e5f6g7h8",
    "title": "My First Post",
    "slug": "my-first-post",
    "excerpt": "This is a brief introduction...",
    "html": "<h2>Title</h2><p>Content...</p>",
    "featuredImage": "https://cdn.example.com/image.jpg",
    "author": {
      "id": "60f7b1c9e1d2c3d4e5f6g7h8",
      "username": "john-doe",
      "displayName": "John Doe",
      "profileImage": "https://cdn.example.com/avatar.jpg",
      "bio": "A passionate writer and developer",
      "location": "Beijing, China",
      "website": "https://johndoe.com",
      "twitter": "johndoe"
    },
    "tags": [
      {"name": "Go", "slug": "go"},
      {"name": "技术", "slug": "tech"}
    ],
    "metaTitle": "Learn Go Programming - Blog",
    "metaDescription": "A comprehensive guide to...",
    "canonicalUrl": "https://example.com/posts/my-first-post",
    "readingTime": 5,
    "wordCount": 1250,
    "viewCount": 1024,
    "publishedAt": "2024-01-01T10:00:00Z",
    "updatedAt": "2024-01-01T09:00:00Z"
  }
}
```

#### 3.1.3. 获取页面详情
```http
GET /api/v1/public/pages/{slug}
```

#### 3.1.4. 获取热门文章
```http
GET /api/v1/public/posts/popular?limit=5
```

**响应示例:**
```json
{
  "code": 200,
  "message": "Success",
  "data": [
    {
      "id": "60f7b1c9e1d2c3d4e5f6g7h8",
      "title": "Popular Post Title",
      "slug": "popular-post",
      "featuredImage": "https://cdn.example.com/image.jpg",
      "viewCount": 5420,
      "publishedAt": "2023-12-15T10:00:00Z"
    }
  ]
}
```

#### 3.1.5. 获取最新文章
```http
GET /api/v1/public/posts/recent?limit=5
```

#### 3.1.6. 获取相关文章
```http
GET /api/v1/public/posts/{postId}/related?limit=3
```

### 3.2. 内容发现模块

#### 3.2.1. 搜索文章
```http
GET /api/v1/public/search?q={keyword}&page=1&limit=10
```

**响应示例:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "query": "Go语言",
    "list": [
      {
        "id": "60f7b1c9e1d2c3d4e5f6g7h8",
        "title": "Go语言入门教程",
        "slug": "go-tutorial",
        "excerpt": "这是一篇关于Go语言的入门教程...",
        "featuredImage": "https://cdn.example.com/go.jpg",
        "author": {
          "displayName": "John Doe"
        },
        "tags": [{"name": "Go", "slug": "go"}],
        "publishedAt": "2024-01-01T10:00:00Z",
        "relevance": 0.95
      }
    ],
    "pagination": {...}
  }
}
```

#### 3.2.2. 获取标签列表
```http
GET /api/v1/public/tags
```

**响应示例:**
```json
{
  "code": 200,
  "message": "Success",
  "data": [
    {
      "id": "60f7b1c9e1d2c3d4e5f6g7h8",
      "name": "Go 语言",
      "slug": "go-language",
      "description": "Go编程语言相关文章",
      "color": "#007d9c",
      "featuredImage": "https://cdn.example.com/tag-go.png",
      "postCount": 25
    }
  ]
}
```

#### 3.2.3. 获取标签详情
```http
GET /api/v1/public/tags/{slug}
```

#### 3.2.4. 获取作者列表
```http
GET /api/v1/public/authors
```

**响应示例:**
```json
{
  "code": 200,
  "message": "Success",
  "data": [
    {
      "id": "60f7b1c9e1d2c3d4e5f6g7h8",
      "username": "john-doe",
      "displayName": "John Doe",
      "profileImage": "https://cdn.example.com/avatar.jpg",
      "bio": "A passionate writer and developer",
      "location": "Beijing, China",
      "website": "https://johndoe.com",
      "twitter": "johndoe",
      "postCount": 15
    }
  ]
}
```

#### 3.2.5. 获取作者详情
```http
GET /api/v1/public/authors/{username}
```

#### 3.2.6. 获取归档列表
```http
GET /api/v1/public/archives
```

**响应示例:**
```json
{
  "code": 200,
  "message": "Success",
  "data": [
    {
      "year": 2024,
      "months": [
        {
          "month": 1,
          "postCount": 5,
          "posts": [
            {
              "id": "60f7b1c9e1d2c3d4e5f6g7h8",
              "title": "January Post",
              "slug": "january-post",
              "publishedAt": "2024-01-15T10:00:00Z"
            }
          ]
        }
      ]
    }
  ]
}
```

### 3.3. 互动系统模块

#### 3.3.1. 获取文章评论
```http
GET /api/v1/public/posts/{postId}/comments?page=1&limit=10
```

**响应示例:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "list": [
      {
        "id": "60f7b1c9e1d2c3d4e5f6g7h8",
        "authorName": "John Visitor",
        "authorUrl": "https://johnvisitor.com",
        "content": "<p>Great post! Very informative.</p>",
        "likeCount": 5,
        "createdAt": "2024-01-01T14:30:00Z",
        "replies": [
          {
            "id": "60f7b1c9e1d2c3d4e5f6g7h9",
            "authorName": "Author",
            "content": "<p>Thank you for your feedback!</p>",
            "createdAt": "2024-01-01T15:00:00Z"
          }
        ]
      }
    ],
    "pagination": {...}
  }
}
```

#### 3.3.2. 提交评论
```http
POST /api/v1/public/posts/{postId}/comments
```

**请求参数:**
```json
{
  "authorName": "John Visitor",
  "authorEmail": "visitor@example.com",
  "authorUrl": "https://johnvisitor.com",
  "content": "Great post! Very informative.",
  "parentId": null
}
```

**响应示例:**
```json
{
  "code": 200,
  "message": "评论已提交，等待审核",
  "data": {
    "id": "60f7b1c9e1d2c3d4e5f6g7h8",
    "status": "pending"
  }
}
```

#### 3.3.3. 点赞评论
```http
POST /api/v1/public/comments/{commentId}/like
```

#### 3.3.4. 取消点赞评论
```http
DELETE /api/v1/public/comments/{commentId}/like
```

### 3.4. 站点信息模块

#### 3.4.1. 获取站点基本信息
```http
GET /api/v1/public/site/info
```

**响应示例:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "title": "My Awesome Blog",
    "description": "A blog about technology and life",
    "logo": "https://cdn.example.com/logo.png",
    "language": "zh-CN",
    "timezone": "Asia/Shanghai",
    "social": {
      "twitter": "myblog",
      "facebook": "myblog",
      "github": "myblog"
    }
  }
}
```

#### 3.4.2. 获取导航菜单
```http
GET /api/v1/public/navigation
```

**响应示例:**
```json
{
  "code": 200,
  "message": "Success",
  "data": [
    {
      "label": "首页",
      "url": "/",
      "target": "_self"
    },
    {
      "label": "关于",
      "url": "/about",
      "target": "_self"
    }
  ]
}
```

### 3.5. SEO 和集成模块

#### 3.5.1. 获取站点地图
```http
GET /api/v1/public/sitemap.xml
Content-Type: application/xml
```

**响应示例:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://example.com/</loc>
    <changefreq>daily</changefreq>
    <priority>1.0</priority>
  </url>
  <url>
    <loc>https://example.com/posts/my-first-post</loc>
    <lastmod>2024-01-01T10:00:00Z</lastmod>
    <changefreq>weekly</changefreq>
    <priority>0.8</priority>
  </url>
</urlset>
```

#### 3.5.2. 获取RSS订阅
```http
GET /api/v1/public/rss.xml
Content-Type: application/rss+xml
```

#### 3.5.3. 获取Atom订阅
```http
GET /api/v1/public/atom.xml
Content-Type: application/atom+xml
```

#### 3.5.4. 获取JSON Feed
```http
GET /api/v1/public/feed.json
Content-Type: application/json
```

---

## 4. 错误码说明

| 错误码 | 说明 | 示例场景 |
|--------|------|----------|
| 200 | 成功 | 请求正常处理 |
| 400 | 请求参数错误 | 参数验证失败 |
| 401 | 未认证 | Token无效或过期 |
| 403 | 权限不足 | 无权限访问资源 |
| 404 | 资源不存在 | 文章、用户不存在 |
| 409 | 资源冲突 | 用户名、邮箱已存在 |
| 422 | 实体无法处理 | 数据格式正确但逻辑错误 |
| 429 | 请求过于频繁 | 触发限流 |
| 500 | 服务器内部错误 | 系统异常 |

---

## 5. 接口版本控制

### 5.1. 版本策略
- 使用URL路径版本控制: `/v1/admin/posts`
- 主要版本变更时创建新的API版本
- 向后兼容的小版本变更保持同一版本号

### 5.2. 弃用策略
- 新版本发布后，旧版本至少维护6个月
- 通过响应头 `X-API-Version` 标识当前版本
- 通过响应头 `X-API-Deprecated` 标识弃用版本

---

**注意**: 本API规范基于RESTful设计原则，所有接口均返回JSON格式数据。实际开发中可能会根据具体需求进行调整。 