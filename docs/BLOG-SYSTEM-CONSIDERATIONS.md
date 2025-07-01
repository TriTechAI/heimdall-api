# 博客系统设计考虑 (Blog System Design Considerations)

本文档补充了针对 Ghost 博客复刻项目的特殊设计考虑，涵盖了通用规范未涉及的博客系统特有功能。

## 1. 内容管理策略

### 1.1. Slug 系统 (SEO 友好 URL)
- **目的**: 为每篇文章生成人类可读、SEO 友好的 URL。
- **实现**: 
  - 每个 `BlogPost` 都应有一个唯一的 `slug` 字段。
  - Slug 应基于文章标题自动生成，但允许手动编辑。
  - 必须确保 slug 在集合中的唯一性。

### 1.2. 内容版本控制
- **草稿 vs 发布版本**: 
  - 使用 `status` 字段区分文章状态。
  - 草稿可以随时修改，发布后的文章修改应谨慎考虑。
- **预定发布**: 支持 `scheduled` 状态，配合 `publishedAt` 时间实现定时发布。

### 1.3. 内容格式
- **Markdown 支持**: 内容以 Markdown 格式存储，便于编辑和版本控制。
- **HTML 缓存**: 考虑在数据库中缓存渲染后的 HTML，提升访问性能。

## 2. 媒体管理

### 2.1. 图片上传与存储
- **存储策略**: 
  - 本地存储：开发环境
  - 云存储：生产环境 (如 腾讯云 COS)
- **图片处理**: 
  - 自动压缩和格式转换
  - 多尺寸缩略图生成
  - WebP 格式支持

### 2.2. 媒体引用
- **内容中的图片**: 在 Markdown 中使用相对路径或完整 URL
- **特色图片**: 每篇文章可设置一个 `featuredImage` 字段

## 3. 评论系统设计

### 3.1. 数据模型选择
```go
// 推荐：独立的 comments 集合（而非内嵌）
type Comment struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    PostID    primitive.ObjectID `bson:"postId"`     // 引用文章
    AuthorID  primitive.ObjectID `bson:"authorId,omitempty"` // 可选：注册用户
    
    // 访客评论支持
    GuestName  string `bson:"guestName,omitempty"`
    GuestEmail string `bson:"guestEmail,omitempty"`
    
    Content   string             `bson:"content"`
    Status    string             `bson:"status"`     // "approved", "pending", "spam"
    ParentID  primitive.ObjectID `bson:"parentId,omitempty"` // 回复功能
    
    CreatedAt time.Time `bson:"createdAt"`
    UpdatedAt time.Time `bson:"updatedAt"`
}
```

### 3.2. 评论管理功能
- **审核机制**: 支持评论审核（pending -> approved）
- **垃圾评论过滤**: 基本的反垃圾策略
- **嵌套回复**: 支持有限层级的回复功能

## 4. 搜索功能

### 4.1. 全文搜索
- **MongoDB 文本索引**: 对 `title` 和 `content` 字段创建文本索引
- **搜索优化**: 
  - 高亮搜索关键词
  - 按相关性排序
  - 支持模糊搜索

### 4.2. 过滤和排序
- **按标签过滤**: 支持多标签筛选
- **按日期排序**: 发布时间、更新时间
- **按作者过滤**: 多作者博客支持

## 5. 缓存策略

### 5.1. 内容缓存
- **热门文章**: 使用 Redis 缓存访问频率高的文章
- **标签云**: 缓存标签及其文章计数
- **归档页面**: 按月/年归档的文章列表

### 5.2. 缓存失效
- **文章更新**: 清除相关文章缓存
- **评论新增**: 更新文章评论计数缓存

## 6. 性能优化建议

### 6.1. 数据库索引
```javascript
// 推荐的 MongoDB 索引
db.blogPosts.createIndex({"slug": 1}, {"unique": true})
db.blogPosts.createIndex({"status": 1, "publishedAt": -1})
db.blogPosts.createIndex({"authorId": 1, "status": 1})
db.blogPosts.createIndex({"tags.slug": 1})
db.comments.createIndex({"postId": 1, "status": 1, "createdAt": -1})
```

### 6.2. 查询优化
- **分页查询**: 使用 `skip` 和 `limit`，但对大数据量考虑游标分页
- **字段投影**: 列表页面只查询必要字段，避免加载完整内容
- **聚合查询**: 使用 MongoDB 聚合管道进行复杂统计

## 7. 安全考虑

### 7.1. 内容安全
- **XSS 防护**: 对用户输入的 HTML 进行清理
- **CSRF 保护**: API 接口的 CSRF 令牌验证
- **内容审核**: 敏感词过滤和人工审核机制

### 7.2. 权限控制
- **细粒度权限**: 不同角色对内容的读写权限
- **文章所有权**: 作者只能编辑自己的文章
- **评论管理**: 管理员可以管理所有评论 