# MongoDB 数据模型设计规范

本规范旨在为项目提供一套统一的、高效的 MongoDB 数据模型设计标准，以确保数据结构的一致性、查询性能和可扩展性。

## 1. 命名规范

- **集合 (Collections)**:
  - 使用复数形式。
  - 使用 `camelCase` 命名法。
  - **示例**: `users`, `blogPosts`, `tags`

- **字段 (Fields)**:
  - 使用 `camelCase` 命名法。
  - 字段名应清晰、简洁，避免使用缩写。
  - **示例**: `firstName`, `createdAt`, `postContent`

## 2. 核心原则：内嵌 vs. 引用

这是 MongoDB 设计中最关键的决策。我们的原则是："**优先内嵌，除非有必要引用**"。

- **使用内嵌 (Embedding)**:
  - **场景**:
    - "包含" (contains) 或 "一对少" (one-to-few) 的关系。
    - 子文档总是与父文档一起被查询，很少独立存在。
    - 数据的原子性要求强。
  - **优点**: 查询性能高，一次查询即可获取所有数据，减少数据库往返。
  - **示例**:
    - 将一篇文章的多个 `tags` 直接内嵌到 `blogPosts` 集合的 `tags` 数组字段中。
    - 将用户的 `address` 对象内嵌到 `users` 集合中。
  - **注意**: 避免因内嵌导致单个文档超过 16MB 的 BSON 上限。

- **使用引用 (Referencing)**:
  - **场景**:
    - "一对多" (one-to-many) 且"多"的一端数量巨大或会无限增长。
    - "多对多" (many-to-many) 关系。
    - 引用的数据经常被独立查询。
  - **优点**: 避免大文档，数据冗余少。
  - **示例**:
    - 一篇文章 (`blogPosts`) 引用一个作者 (`users`)。在 `blogPosts` 中存储 `authorId` (`primitive.ObjectID`)。
    - 评论 (`comments`) 引用一篇文章 (`blogPosts`)。在 `comments` 集合中存储 `postId`。

## 3. 数据类型与 Go Struct 定义

- **ID**:
  - 所有集合的主键都应命名为 `_id`。
  - 在 Go struct 中，其类型必须是 `primitive.ObjectID`。
  - **必须**使用 `bson:"_id,omitempty"` 标签。`omitempty` 可以在插入新文档时让 MongoDB 自动生成 `_id`。

- **时间戳**:
  - 所有需要记录时间的字段都应使用 `time.Time` 类型。
  - 推荐为每个集合添加 `createdAt` 和 `updatedAt` 字段，用于追踪文档的创建和最后修改时间。

- **Struct 标签**:
  - 所有需要持久化到 MongoDB 的字段都**必须**有 `bson:"..."` 标签。
  - 使用 `omitempty` 标签可以避免将 Go 中的零值（如 `nil`, `""`, `0`）存入数据库，保持数据清洁。

**示例 `BlogPost` 模型**:

```go
// file: common/model/blog_post.go
package model

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// Tag 被内嵌到 BlogPost 中，作为内嵌文档不需要独立的 _id
type Tag struct {
    Name string `bson:"name"`
    Slug string `bson:"slug"`
}

type BlogPost struct {
    ID          primitive.ObjectID   `bson:"_id,omitempty"`
    Title       string               `bson:"title"`
    Slug        string               `bson:"slug"`
    Content     string               `bson:"content"`
    Status      string               `bson:"status"`      // e.g., "published", "draft"
    AuthorID    primitive.ObjectID   `bson:"authorId"`    // 引用自 'users' 集合
    Tags        []Tag                `bson:"tags"`        // 内嵌 Tag 数组
    PublishedAt time.Time            `bson:"publishedAt,omitempty"`
    CreatedAt   time.Time            `bson:"createdAt"`
    UpdatedAt   time.Time            `bson:"updatedAt"`
}
```

## 4. 索引 (Indexes)

- **性能关键**: 必须为所有常用的查询字段创建索引。
- **默认索引**: `_id` 字段会自动创建唯一索引。
- **创建时机**: 在 `DAO` 层的初始化逻辑中，通过调用 `mongo.IndexModel` 和 `CreateOne/CreateMany` 来确保服务启动时索引已存在。
- **复合索引**: 对于需要同时对多个字段进行排序或过滤的查询，应创建复合索引。 