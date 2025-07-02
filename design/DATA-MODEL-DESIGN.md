 # 数据模型设计 (Data Model Design)

本文档详细定义了 Heimdall 项目在 MongoDB 中的数据模型结构，并根据业务模块进行划分。所有设计遵循 `MONGODB-MODELING-GUIDELINES.md` 规范。

## 1. 核心设计原则

- **引用与内嵌**: "优先内嵌，除非有必要引用"。我们将在适当的地方使用内嵌来提升查询性能，在必要时使用引用来避免数据冗余和文档过大。
- **命名规范**: 集合使用复数 `camelCase` (如 `blogPosts`)，字段使用 `camelCase` (如 `createdAt`)。
- **时间戳**: 所有核心模型都包含 `createdAt` 和 `updatedAt` 字段，类型为 `time.Time`。
- **ID**: 主键固定为 `_id`，类型为 `primitive.ObjectID`。

---

## 2. 用户与权限模块 (User & Access Module)

### 2.1. `users` 集合

该集合存储所有可以登录后台的用户信息，包括作者、管理员等。

> **命名说明**: 根据 MongoDB 建模规范，集合应使用 `camelCase` 命名。但考虑到与其他主流博客系统的兼容性和迁移便利性，本项目中的核心集合继续使用小写复数形式。

| 字段名         | Go 类型              | BSON 类型         | 描述                                     | 示例                               |
| -------------- | -------------------- | ----------------- | ---------------------------------------- | ---------------------------------- |
| `_id`          | `primitive.ObjectID` | `ObjectID`        | 文档唯一标识符                           | `ObjectId("...")`                  |
| `username`     | `string`             | `String`          | 唯一的用户名，用于登录和URL              | `"john-doe"`                       |
| `email`        | `string`             | `String`          | 唯一的邮箱地址，用于登录和通知           | `"john.doe@example.com"`           |
| `passwordHash` | `string`             | `String`          | 加密后的用户密码                         | `"$2a$10$..."`                     |
| `displayName`  | `string`             | `String`          | 公开显示的姓名                           | `"John Doe"`                       |
| `role`         | `string`             | `String`          | 用户角色 (常量定义)                      | `"Author"`, `"Admin"`, `"Owner"`   |
| `profileImage` | `string`             | `String`          | 用户头像的URL                            | `"https://.../avatar.png"`         |
| `coverImage`   | `string`             | `String`          | 用户封面图片的URL                        | `"https://.../cover.png"`          |
| `bio`          | `string`             | `String`          | 用户简介                                 | `"A passionate writer and..."`     |
| `location`     | `string`             | `String`          | 用户所在地                               | `"Beijing, China"`                 |
| `website`      | `string`             | `String`          | 用户个人网站                             | `"https://johndoe.com"`            |
| `twitter`      | `string`             | `String`          | Twitter 用户名                           | `"johndoe"`                        |
| `facebook`     | `string`             | `String`          | Facebook 用户名                          | `"john.doe"`                       |
| `status`       | `string`             | `String`          | 用户状态 (常量定义)                      | `"active"`, `"inactive"`, `"locked"` |
| `loginFailCount` | `int`                | `Int32`           | 登录失败次数                             | `3`                                |
| `lockedUntil`  | `time.Time`          | `Date`            | 账号锁定至（可选）                       | `ISODate("...")`                   |
| `lastLoginAt`  | `time.Time`          | `Date`            | 最后登录时间                             | `ISODate("...")`                   |
| `lastLoginIP`  | `string`             | `String`          | 最后登录IP地址                           | `"192.168.1.1"`                   |
| `createdAt`    | `time.Time`          | `Date`            | 创建时间                                 | `ISODate("...")`                   |
| `updatedAt`    | `time.Time`          | `Date`            | 最后更新时间                             | `ISODate("...")`                   |

#### 关系
- 与 `posts` 集合是 **一对多** 关系 (一个用户可以有多篇文章)。`posts` 集合通过 `authorId` 字段引用 `users` 集合的 `_id`。

#### 建议索引
- `db.users.createIndex({ "username": 1 }, { "unique": true })`
- `db.users.createIndex({ "email": 1 }, { "unique": true })`
- `db.users.createIndex({ "role": 1, "status": 1 })`
- `db.users.createIndex({ "lockedUntil": 1 })` (用于解锁已过期的账号)

### 2.2. `loginLogs` 集合

记录所有登录尝试和安全事件，用于安全监控和审计。

| 字段名       | Go 类型              | BSON 类型  | 描述                           | 示例                                     |
| ------------ | -------------------- | ---------- | ------------------------------ | ---------------------------------------- |
| `_id`        | `primitive.ObjectID` | `ObjectID` | 文档唯一标识符                 | `ObjectId("...")`                        |
| `userId`     | `primitive.ObjectID` | `ObjectID` | (可选) **引用** `users` 集合的 `_id` | `ObjectId("...")`                    |
| `username`   | `string`             | `String`   | 尝试登录的用户名               | `"john-doe"`                             |
| `email`      | `string`             | `String`   | 尝试登录的邮箱                 | `"john@example.com"`                     |
| `ipAddress`  | `string`             | `String`   | 登录IP地址                     | `"192.168.1.1"`                         |
| `userAgent`  | `string`             | `String`   | 用户代理字符串                 | `"Mozilla/5.0 ..."`                     |
| `success`    | `bool`               | `Boolean`  | 登录是否成功                   | `true`, `false`                          |
| `failReason` | `string`             | `String`   | 失败原因 (可选)                | `"invalid_password"`, `"user_locked"`    |
| `createdAt`  | `time.Time`          | `Date`     | 登录尝试时间                   | `ISODate("...")`                         |

#### 建议索引
- `db.loginLogs.createIndex({ "userId": 1, "createdAt": -1 })`
- `db.loginLogs.createIndex({ "ipAddress": 1, "createdAt": -1 })`
- `db.loginLogs.createIndex({ "success": 1, "createdAt": -1 })`

---

## 3. 内容管理模块 (Content Management Module)

### 3.1. `posts` 集合

博客的核心内容，存储所有文章和页面。

| 字段名            | Go 类型              | BSON 类型         | 描述                                     | 示例                                  |
| ----------------- | -------------------- | ----------------- | ---------------------------------------- | ------------------------------------- |
| `_id`             | `primitive.ObjectID` | `ObjectID`        | 文档唯一标识符                           | `ObjectId("...")`                     |
| `title`           | `string`             | `String`          | 文章标题                                 | `"My First Post"`                     |
| `slug`            | `string`             | `String`          | SEO友好的URL片段，唯一                   | `"my-first-post"`                     |
| `excerpt`         | `string`             | `String`          | 文章摘要/简介                            | `"This is a brief introduction..."`   |
| `markdown`        | `string`             | `String`          | Markdown 格式的原文                      | `"## Title\n\nContent..."`            |
| `html`            | `string`             | `String`          | 渲染后的HTML，用于提高前台性能           | `"<h2>Title</h2><p>Content...</p>"` |
| `featuredImage`   | `string`             | `String`          | 文章特色图片的URL                        | `"https://.../image.png"`             |
| `type`            | `string`             | `String`          | 内容类型 (常量定义)                      | `"post"`, `"page"`                    |
| `status`          | `string`             | `String`          | 文章状态 (常量定义)                      | `"published"`, `"draft"`, `"scheduled"` |
| `visibility`      | `string`             | `String`          | 可见性 (常量定义)                        | `"public"`, `"members_only"`          |
| `authorId`        | `primitive.ObjectID` | `ObjectID`        | **引用** `users` 集合的 `_id`            | `ObjectId("...")`                     |
| `tags`            | `[]Tag`              | `Array` of `Doc`  | **内嵌** 标签数组                        | `[{name:"Go", slug:"go"}, ...]`       |
| `metaTitle`       | `string`             | `String`          | SEO标题 (可选)                           | `"Learn Go Programming - Blog"`       |
| `metaDescription` | `string`             | `String`          | SEO描述                                  | `"A comprehensive guide to..."`       |
| `canonicalUrl`    | `string`             | `String`          | 规范化URL (可选)                         | `"https://example.com/canonical"`     |
| `readingTime`     | `int`                | `Int32`           | 预估阅读时间(分钟)                       | `5`                                   |
| `wordCount`       | `int`                | `Int32`           | 字数统计                                 | `1250`                                |
| `viewCount`       | `int64`              | `Int64`           | 浏览量                                   | `1024`                                |
| `publishedAt`     | `time.Time`          | `Date`            | 文章发布或计划发布的时间                 | `ISODate("...")`                      |
| `createdAt`       | `time.Time`          | `Date`            | 创建时间                                 | `ISODate("...")`                      |
| `updatedAt`       | `time.Time`          | `Date`            | 最后更新时间                             | `ISODate("...")`                      |

#### `Tag` 内嵌文档结构

| 字段名 | Go 类型  | BSON 类型 | 描述             | 示例     |
| ------ | -------- | --------- | ---------------- | -------- |
| `name` | `string` | `String`  | 标签显示名称     | `"Go"`   |
| `slug` | `string` | `String`  | 标签URL友好型名称 | `"go"`   |

#### 关系
- 与 `users` 集合是 **多对一** 关系。
- 与 `comments` 集合是 **一对多** 关系。

#### 建议索引
- `db.posts.createIndex({ "slug": 1 }, { "unique": true })`
- `db.posts.createIndex({ "status": 1, "publishedAt": -1 })` (用于查询已发布的文章列表)
- `db.posts.createIndex({ "authorId": 1, "status": 1 })`
- `db.posts.createIndex({ "tags.slug": 1 })` (用于按标签查询)
- `db.posts.createIndex({ "type": 1, "status": 1 })` (用于区分文章和页面)
- `db.posts.createIndex({ "viewCount": -1 })` (用于热门文章排序)
- `db.posts.createIndex({ "title": "text", "excerpt": "text" })` (全文搜索)

---

## 4. 互动系统模块 (Interaction System Module)

### 4.1. `comments` 集合

存储所有文章的评论，支持嵌套。

| 字段名        | Go 类型              | BSON 类型  | 描述                                     | 示例                             |
| ------------- | -------------------- | ---------- | ---------------------------------------- | -------------------------------- |
| `_id`         | `primitive.ObjectID` | `ObjectID` | 文档唯一标识符                           | `ObjectId("...")`                |
| `postId`      | `primitive.ObjectID` | `ObjectID` | **引用** `posts` 集合的 `_id`            | `ObjectId("...")`                |
| `authorId`    | `primitive.ObjectID` | `ObjectID` | (可选) **引用** `users` 集合的 `_id`     | `ObjectId("...")`                |
| `authorName`  | `string`             | `String`   | 评论者昵称 (注册用户或游客)              | `"John Doe"` / `"A Visitor"`     |
| `authorEmail` | `string`             | `String`   | 评论者邮箱 (不公开显示)                  | `"visitor@test.com"`             |
| `authorUrl`   | `string`             | `String`   | 评论者网站 (可选)                        | `"https://example.com"`          |
| `content`     | `string`             | `String`   | 评论内容                                 | `"Great post!"`                  |
| `htmlContent` | `string`             | `String`   | 渲染后的HTML内容                         | `"<p>Great post!</p>"`           |
| `status`      | `string`             | `String`   | 评论状态 (常量定义)                      | `"approved"`, `"pending"`, `"spam"` |
| `parentId`    | `primitive.ObjectID` | `ObjectID` | (可选) **引用** 同集合的 `_id` 用于回复 | `ObjectId("...")`                |
| `ipAddress`   | `string`             | `String`   | 评论者IP地址 (用于反垃圾)                | `"192.168.1.1"`                  |
| `userAgent`   | `string`             | `String`   | 用户代理字符串 (用于反垃圾)              | `"Mozilla/5.0 ..."`              |
| `likeCount`   | `int`                | `Int32`    | 点赞数                                   | `10`                             |
| `createdAt`   | `time.Time`          | `Date`     | 创建时间                                 | `ISODate("...")`                 |
| `updatedAt`   | `time.Time`          | `Date`     | 最后更新时间                             | `ISODate("...")`                 |

#### 关系
- 与 `posts` 集合是 **多对一** 关系。
- 与 `users` 集合是 **多对一** 关系 (可选)。
- 与自身是 **多对一** 关系 (通过 `parentId` 实现嵌套)。

#### 建议索引
- `db.comments.createIndex({ "postId": 1, "status": 1, "createdAt": -1 })` (用于查询一篇文章下的评论)
- `db.comments.createIndex({ "status": 1 })` (用于后台管理审核评论)
- `db.comments.createIndex({ "parentId": 1 })` (用于查询回复评论)
- `db.comments.createIndex({ "ipAddress": 1 })` (用于反垃圾检测)

---

## 5. 站点设置模块 (Site Settings Module)

### 5.1. `settings` 集合

以键值对形式存储所有站点级别的配置。

| 字段名    | Go 类型              | BSON 类型  | 描述                     | 示例                                     |
| --------- | -------------------- | ---------- | ------------------------ | ---------------------------------------- |
| `_id`     | `primitive.ObjectID` | `ObjectID` | 文档唯一标识符           | `ObjectId("...")`                        |
| `key`     | `string`             | `String`   | 配置项的唯一键           | `"title"`, `"logo"`, `"postsPerPage"`    |
| `value`   | `string`             | `String`   | 配置项的值               | `"My Awesome Blog"`, `"/logo.png"`, `"10"` |
| `group`   | `string`             | `String`   | 配置项分组，便于后台管理 | `"general"`, `"design"`, `"social"`      |
| `createdAt` | `time.Time`          | `Date`     | 创建时间                 | `ISODate("...")`                         |
| `updatedAt` | `time.Time`          | `Date`     | 最后更新时间             | `ISODate("...")`                         |

#### 建议索引
- `db.settings.createIndex({ "key": 1 }, { "unique": true })`

---

## 6. 媒体管理模块 (Media Management Module)

### 6.1. `media` 集合

存储所有上传的媒体文件信息。

| 字段名       | Go 类型              | BSON 类型  | 描述                           | 示例                                     |
| ------------ | -------------------- | ---------- | ------------------------------ | ---------------------------------------- |
| `_id`        | `primitive.ObjectID` | `ObjectID` | 文档唯一标识符                 | `ObjectId("...")`                        |
| `filename`   | `string`             | `String`   | 原始文件名                     | `"my-image.jpg"`                         |
| `url`        | `string`             | `String`   | 文件访问URL                    | `"https://example.com/media/image.jpg"`  |
| `type`       | `string`             | `String`   | 文件类型 (常量定义)            | `"image"`, `"document"`, `"video"`       |
| `mimeType`   | `string`             | `String`   | MIME类型                       | `"image/jpeg"`, `"application/pdf"`      |
| `size`       | `int64`              | `Int64`    | 文件大小 (字节)                | `2048576`                                |
| `width`      | `int`                | `Int32`    | 图片宽度 (仅图片文件)          | `1920`                                   |
| `height`     | `int`                | `Int32`    | 图片高度 (仅图片文件)          | `1080`                                   |
| `alt`        | `string`             | `String`   | 图片替代文本                   | `"A beautiful sunset"`                   |
| `title`      | `string`             | `String`   | 文件标题                       | `"Sunset Photo"`                         |
| `uploaderId` | `primitive.ObjectID` | `ObjectID` | **引用** `users` 集合的 `_id`  | `ObjectId("...")`                        |
| `createdAt`  | `time.Time`          | `Date`     | 上传时间                       | `ISODate("...")`                         |
| `updatedAt`  | `time.Time`          | `Date`     | 最后更新时间                   | `ISODate("...")`                         |

#### 建议索引
- `db.media.createIndex({ "uploaderId": 1, "createdAt": -1 })`
- `db.media.createIndex({ "type": 1 })`

---

## 7. 导航菜单模块 (Navigation Module)

### 7.1. `navigation` 集合

存储博客的导航菜单配置。

| 字段名      | Go 类型              | BSON 类型  | 描述                     | 示例                     |
| ----------- | -------------------- | ---------- | ------------------------ | ------------------------ |
| `_id`       | `primitive.ObjectID` | `ObjectID` | 文档唯一标识符           | `ObjectId("...")`        |
| `label`     | `string`             | `String`   | 菜单显示名称             | `"首页"`, `"关于"`       |
| `url`       | `string`             | `String`   | 菜单链接地址             | `"/"`, `"/about"`        |
| `order`     | `int`                | `Int32`    | 排序顺序                 | `1`, `2`, `3`            |
| `target`    | `string`             | `String`   | 链接打开方式             | `"_self"`, `"_blank"`    |
| `isActive`  | `bool`               | `Boolean`  | 是否启用                 | `true`, `false`          |
| `createdAt` | `time.Time`          | `Date`     | 创建时间                 | `ISODate("...")`         |
| `updatedAt` | `time.Time`          | `Date`     | 最后更新时间             | `ISODate("...")`         |

#### 建议索引
- `db.navigation.createIndex({ "order": 1 })`
- `db.navigation.createIndex({ "isActive": 1, "order": 1 })`

---

## 8. 标签管理模块 (Tags Management Module)

### 8.1. `tags` 集合

独立管理博客标签，支持标签的统计和SEO优化。

| 字段名          | Go 类型              | BSON 类型  | 描述                     | 示例                     |
| --------------- | -------------------- | ---------- | ------------------------ | ------------------------ |
| `_id`           | `primitive.ObjectID` | `ObjectID` | 文档唯一标识符           | `ObjectId("...")`        |
| `name`          | `string`             | `String`   | 标签名称                 | `"Go 语言"`              |
| `slug`          | `string`             | `String`   | URL友好的标签名称，唯一  | `"go-language"`          |
| `description`   | `string`             | `String`   | 标签描述                 | `"Go编程语言相关文章"`   |
| `color`         | `string`             | `String`   | 标签颜色                 | `"#007d9c"`              |
| `featuredImage` | `string`             | `String`   | 标签特色图片             | `"https://.../tag.png"`  |
| `metaTitle`     | `string`             | `String`   | SEO标题                  | `"Go编程 - 技术博客"`    |
| `metaDescription`| `string`            | `String`   | SEO描述                  | `"Go语言编程技术文章"`   |
| `postCount`     | `int`                | `Int32`    | 使用该标签的文章数       | `25`                     |
| `visibility`    | `string`             | `String`   | 可见性                   | `"public"`, `"internal"` |
| `createdAt`     | `time.Time`          | `Date`     | 创建时间                 | `ISODate("...")`         |
| `updatedAt`     | `time.Time`          | `Date`     | 最后更新时间             | `ISODate("...")`         |

#### 建议索引
- `db.tags.createIndex({ "slug": 1 }, { "unique": true })`
- `db.tags.createIndex({ "visibility": 1, "postCount": -1 })`
- `db.tags.createIndex({ "name": "text", "description": "text" })` (全文搜索)

---

## 9. 数据模型关系总结

### 9.1. 集合概览

| 集合名        | 主要用途                 | 预估文档数量     | 访问频率 |
| ------------- | ------------------------ | ---------------- | -------- |
| `users`       | 用户账户管理             | 少量 (< 100)     | 中等     |
| `loginLogs`   | 登录日志和安全审计       | 大量 (数万)      | 中等     |
| `posts`       | 文章和页面内容           | 中等 (数千)      | 高       |
| `comments`    | 文章评论                 | 大量 (数万)      | 高       |
| `settings`    | 站点配置                 | 少量 (< 50)      | 中等     |
| `media`       | 媒体文件管理             | 中等 (数千)      | 中等     |
| `navigation`  | 导航菜单                 | 少量 (< 20)      | 高       |
| `tags`        | 标签管理                 | 少量 (< 200)     | 高       |

### 9.2. 关系映射

- **用户 -> 文章**: `users._id` <- `posts.authorId` (一对多)
- **用户 -> 评论**: `users._id` <- `comments.authorId` (一对多，可选)
- **用户 -> 媒体**: `users._id` <- `media.uploaderId` (一对多)
- **用户 -> 登录日志**: `users._id` <- `loginLogs.userId` (一对多)
- **文章 -> 评论**: `posts._id` <- `comments.postId` (一对多)
- **文章 -> 标签**: `posts.tags[]` (内嵌关系)
- **评论 -> 评论**: `comments._id` <- `comments.parentId` (自引用，用于嵌套回复)

### 9.3. 缓存策略建议

- **热点数据**:
  - 文章列表: 缓存 1 小时
  - 热门文章: 缓存 6 小时  
  - 标签云: 缓存 24 小时
  - 导航菜单: 缓存 24 小时
  - 站点设置: 缓存 24 小时

- **动态数据**:
  - 文章内容: 缓存 1 小时 (发布后)
  - 评论列表: 缓存 30 分钟
  - 用户会话: 缓存至Token过期

- **安全数据**:
  - 登录失败计数: 缓存 1 小时 (自动重置)
  - 用户锁定状态: 缓存至锁定期结束
  - IP访问频率: 缓存 15 分钟 (API限流)

### 9.4. 性能优化建议

1. **分页查询**: 所有列表查询都实现分页，默认每页 10-20 条记录
2. **字段投影**: 列表查询只返回必要字段，详情查询返回完整数据
3. **复合索引**: 为常用的多字段查询组合创建复合索引
4. **全文搜索**: 为 `posts` 和 `tags` 集合启用 MongoDB 全文搜索索引
5. **聚合管道**: 复杂统计查询使用 MongoDB 聚合管道
6. **读写分离**: Public API 主要进行读操作，Admin API 进行读写操作

---

**注意**: 本数据模型设计遵循 MongoDB 最佳实践，平衡了查询性能、数据一致性和扩展性。在实际开发中，可能会根据具体业务需求进行微调。
