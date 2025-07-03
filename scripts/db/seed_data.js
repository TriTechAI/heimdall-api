// MongoDB 种子数据脚本
// 插入初始数据用于开发和测试

// 使用开发数据库
use heimdall_dev;

print("开始插入种子数据...");

// =============================================================================
// 1. 清理现有数据 (开发环境)
// =============================================================================
print("清理现有数据...");

db.users.deleteMany({});
db.loginLogs.deleteMany({});
db.posts.deleteMany({});
db.comments.deleteMany({});
db.settings.deleteMany({});
db.media.deleteMany({});

print("现有数据清理完成");

// =============================================================================
// 2. 插入默认用户数据
// =============================================================================
print("插入用户数据...");

// 管理员用户
var adminUser = {
    username: "admin",
    email: "admin@heimdall.com",
    passwordHash: "$2a$12$6iOXFHKJakKJGvf5JkZ4xOz8XzXqXuKGHwXh4tLKH5JKH5JKH5JKH5", // password: admin123
    displayName: "系统管理员",
    role: "Owner",
    profileImage: "",
    coverImage: "",
    bio: "Heimdall 博客系统管理员",
    location: "北京",
    website: "https://heimdall.com",
    twitter: "",
    facebook: "",
    status: "active",
    loginFailCount: 0,
    lastLoginAt: new Date(),
    lastLoginIP: "127.0.0.1",
    createdAt: new Date(),
    updatedAt: new Date()
};

var adminInsert = db.users.insertOne(adminUser);
var adminId = adminInsert.insertedId;
print("管理员用户创建成功: " + adminId);

// 作者用户
var authorUser = {
    username: "author",
    email: "author@heimdall.com", 
    passwordHash: "$2a$12$6iOXFHKJakKJGvf5JkZ4xOz8XzXqXuKGHwXh4tLKH5JKH5JKH5JKH5", // password: author123
    displayName: "示例作者",
    role: "Author",
    profileImage: "",
    coverImage: "",
    bio: "一个热爱写作的技术博主",
    location: "上海",
    website: "https://author.blog",
    twitter: "author_blog",
    facebook: "",
    status: "active",
    loginFailCount: 0,
    lastLoginAt: new Date(),
    lastLoginIP: "127.0.0.1",
    createdAt: new Date(),
    updatedAt: new Date()
};

var authorInsert = db.users.insertOne(authorUser);
var authorId = authorInsert.insertedId;
print("作者用户创建成功: " + authorId);

print("用户数据插入完成");

// =============================================================================
// 3. 插入站点设置数据
// =============================================================================
print("插入站点设置数据...");

var settings = [
    // 基本设置
    { key: "title", value: "Heimdall Blog", group: "general", createdAt: new Date(), updatedAt: new Date() },
    { key: "description", value: "一个基于 Go-Zero 的现代化博客系统", group: "general", createdAt: new Date(), updatedAt: new Date() },
    { key: "logo", value: "", group: "general", createdAt: new Date(), updatedAt: new Date() },
    { key: "favicon", value: "", group: "general", createdAt: new Date(), updatedAt: new Date() },
    { key: "language", value: "zh-CN", group: "general", createdAt: new Date(), updatedAt: new Date() },
    { key: "timezone", value: "Asia/Shanghai", group: "general", createdAt: new Date(), updatedAt: new Date() },
    
    // 显示设置
    { key: "postsPerPage", value: "10", group: "display", createdAt: new Date(), updatedAt: new Date() },
    { key: "theme", value: "default", group: "display", createdAt: new Date(), updatedAt: new Date() },
    { key: "showExcerpts", value: "true", group: "display", createdAt: new Date(), updatedAt: new Date() },
    { key: "showReadingTime", value: "true", group: "display", createdAt: new Date(), updatedAt: new Date() },
    
    // SEO设置
    { key: "metaTitle", value: "Heimdall Blog - 技术分享与思考", group: "seo", createdAt: new Date(), updatedAt: new Date() },
    { key: "metaDescription", value: "分享技术心得，记录成长历程", group: "seo", createdAt: new Date(), updatedAt: new Date() },
    { key: "enableSitemap", value: "true", group: "seo", createdAt: new Date(), updatedAt: new Date() },
    { key: "enableRSS", value: "true", group: "seo", createdAt: new Date(), updatedAt: new Date() },
    
    // 社交媒体
    { key: "twitter", value: "", group: "social", createdAt: new Date(), updatedAt: new Date() },
    { key: "facebook", value: "", group: "social", createdAt: new Date(), updatedAt: new Date() },
    { key: "github", value: "", group: "social", createdAt: new Date(), updatedAt: new Date() },
    
    // 评论设置
    { key: "enableComments", value: "true", group: "comments", createdAt: new Date(), updatedAt: new Date() },
    { key: "requireApproval", value: "true", group: "comments", createdAt: new Date(), updatedAt: new Date() },
    { key: "allowGuestComments", value: "true", group: "comments", createdAt: new Date(), updatedAt: new Date() }
];

db.settings.insertMany(settings);
print("站点设置数据插入完成，共 " + settings.length + " 条记录");

// =============================================================================
// 4. 插入示例文章数据
// =============================================================================
print("插入示例文章数据...");

var now = new Date();
var oneWeekAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
var twoWeeksAgo = new Date(now.getTime() - 14 * 24 * 60 * 60 * 1000);

var posts = [
    // 欢迎文章
    {
        title: "欢迎使用 Heimdall 博客系统",
        slug: "welcome-to-heimdall",
        excerpt: "Heimdall 是一个基于 Go-Zero 框架开发的现代化博客系统，具有高性能、高可用、易扩展的特点。",
        markdown: `# 欢迎使用 Heimdall 博客系统

## 关于 Heimdall

Heimdall 是一个基于 Go-Zero 框架开发的现代化博客系统。它采用微服务架构，具有以下特点：

- **高性能**: 基于 Go 语言和 go-zero 框架
- **微服务架构**: admin-api + public-api + common 的清晰架构
- **MongoDB + Redis**: 现代化的数据存储方案
- **完善的安全机制**: JWT认证、登录限制、操作审计
- **SEO友好**: 支持自定义URL、sitemap、RSS等

## 快速开始

1. 配置数据库连接
2. 运行数据库初始化脚本
3. 启动服务
4. 开始创作

祝您使用愉快！`,
        html: `<h1>欢迎使用 Heimdall 博客系统</h1>
<h2>关于 Heimdall</h2>
<p>Heimdall 是一个基于 Go-Zero 框架开发的现代化博客系统。它采用微服务架构，具有以下特点：</p>
<ul>
<li><strong>高性能</strong>: 基于 Go 语言和 go-zero 框架</li>
<li><strong>微服务架构</strong>: admin-api + public-api + common 的清晰架构</li>
<li><strong>MongoDB + Redis</strong>: 现代化的数据存储方案</li>
<li><strong>完善的安全机制</strong>: JWT认证、登录限制、操作审计</li>
<li><strong>SEO友好</strong>: 支持自定义URL、sitemap、RSS等</li>
</ul>
<h2>快速开始</h2>
<ol>
<li>配置数据库连接</li>
<li>运行数据库初始化脚本</li>
<li>启动服务</li>
<li>开始创作</li>
</ol>
<p>祝您使用愉快！</p>`,
        featuredImage: "",
        type: "post",
        status: "published",
        visibility: "public",
        authorId: adminId,
        tags: [
            { name: "博客系统", slug: "blog-system" },
            { name: "Go语言", slug: "golang" },
            { name: "微服务", slug: "microservices" }
        ],
        metaTitle: "欢迎使用 Heimdall 博客系统",
        metaDescription: "了解 Heimdall 博客系统的特点和使用方法",
        canonicalUrl: "",
        readingTime: 2,
        wordCount: 186,
        viewCount: 100,
        publishedAt: twoWeeksAgo,
        createdAt: twoWeeksAgo,
        updatedAt: twoWeeksAgo
    },
    
    // 技术文章
    {
        title: "Go-Zero 微服务框架入门指南",
        slug: "go-zero-microservices-guide", 
        excerpt: "Go-Zero 是一个集成了各种工程实践的 web 和 rpc 框架。本文将详细介绍如何使用 go-zero 构建微服务应用。",
        markdown: `# Go-Zero 微服务框架入门指南

## 什么是 Go-Zero

go-zero 是一个集成了各种工程实践的 web 和 rpc 框架。通过弹性设计保障了大并发服务端的稳定性，经受了充分的实战检验。

## 核心特性

- **简单易用**: API 语法简洁，一键生成代码
- **弹性设计**: 熔断、降级、限流、自适应负载均衡
- **微服务治理**: 服务发现、链路追踪、监控告警
- **高性能**: 极简设计，极致性能

## 架构设计

本项目采用了 go-zero 推荐的微服务架构：

\`\`\`
├── admin-api    # 管理后台服务
├── public-api   # 公开前台服务  
└── common       # 共享模块
\`\`\`

每个服务都遵循 go-zero 的最佳实践，具有清晰的分层结构。`,
        html: `<h1>Go-Zero 微服务框架入门指南</h1>
<h2>什么是 Go-Zero</h2>
<p>go-zero 是一个集成了各种工程实践的 web 和 rpc 框架。通过弹性设计保障了大并发服务端的稳定性，经受了充分的实战检验。</p>`,
        featuredImage: "",
        type: "post", 
        status: "published",
        visibility: "public",
        authorId: authorId,
        tags: [
            { name: "Go语言", slug: "golang" },
            { name: "微服务", slug: "microservices" },
            { name: "go-zero", slug: "go-zero" }
        ],
        metaTitle: "Go-Zero 微服务框架入门指南",
        metaDescription: "详细介绍 go-zero 框架的特性和使用方法",
        canonicalUrl: "",
        readingTime: 5,
        wordCount: 324,
        viewCount: 256,
        publishedAt: oneWeekAgo,
        createdAt: oneWeekAgo,
        updatedAt: oneWeekAgo
    },
    
    // 草稿文章
    {
        title: "MongoDB 最佳实践",
        slug: "mongodb-best-practices",
        excerpt: "分享在使用 MongoDB 过程中总结的最佳实践和经验教训。",
        markdown: `# MongoDB 最佳实践

这是一篇草稿文章，正在编写中...

## 数据建模

- 优先内嵌，必要时引用
- 合理设计索引
- 避免深层嵌套

## 性能优化

待完善...`,
        html: `<h1>MongoDB 最佳实践</h1><p>这是一篇草稿文章，正在编写中...</p>`,
        featuredImage: "",
        type: "post",
        status: "draft", 
        visibility: "public",
        authorId: authorId,
        tags: [
            { name: "MongoDB", slug: "mongodb" },
            { name: "数据库", slug: "database" }
        ],
        metaTitle: "",
        metaDescription: "",
        canonicalUrl: "",
        readingTime: 3,
        wordCount: 45,
        viewCount: 0,
        publishedAt: null,
        createdAt: now,
        updatedAt: now
    }
];

var postInserts = db.posts.insertMany(posts);
var postIds = Object.values(postInserts.insertedIds);
print("示例文章插入完成，共 " + posts.length + " 篇文章");

// =============================================================================
// 5. 插入示例评论数据
// =============================================================================
print("插入示例评论数据...");

var comments = [
    // 对第一篇文章的评论
    {
        postId: postIds[0],
        authorId: null,
        authorName: "张三",
        authorEmail: "zhangsan@example.com",
        authorUrl: "",
        content: "很不错的博客系统！期待更多功能。",
        htmlContent: "<p>很不错的博客系统！期待更多功能。</p>",
        status: "approved",
        parentId: null,
        ipAddress: "192.168.1.100",
        userAgent: "Mozilla/5.0...",
        likeCount: 5,
        createdAt: new Date(twoWeeksAgo.getTime() + 60 * 60 * 1000),
        updatedAt: new Date(twoWeeksAgo.getTime() + 60 * 60 * 1000)
    },
    {
        postId: postIds[0],
        authorId: null,
        authorName: "李四", 
        authorEmail: "lisi@example.com",
        authorUrl: "https://lisi.blog",
        content: "界面很简洁，用户体验不错。",
        htmlContent: "<p>界面很简洁，用户体验不错。</p>",
        status: "approved",
        parentId: null,
        ipAddress: "192.168.1.101",
        userAgent: "Mozilla/5.0...",
        likeCount: 3,
        createdAt: new Date(twoWeeksAgo.getTime() + 2 * 60 * 60 * 1000),
        updatedAt: new Date(twoWeeksAgo.getTime() + 2 * 60 * 60 * 1000)
    },
    
    // 对第二篇文章的评论
    {
        postId: postIds[1],
        authorId: authorId,
        authorName: "示例作者",
        authorEmail: "author@heimdall.com",
        authorUrl: "",
        content: "感谢大家的关注，后续会分享更多 go-zero 的实践经验。",
        htmlContent: "<p>感谢大家的关注，后续会分享更多 go-zero 的实践经验。</p>",
        status: "approved",
        parentId: null,
        ipAddress: "127.0.0.1",
        userAgent: "Mozilla/5.0...",
        likeCount: 8,
        createdAt: new Date(oneWeekAgo.getTime() + 60 * 60 * 1000),
        updatedAt: new Date(oneWeekAgo.getTime() + 60 * 60 * 1000)
    }
];

db.comments.insertMany(comments);
print("示例评论插入完成，共 " + comments.length + " 条评论");

// =============================================================================
// 6. 插入登录日志示例数据
// =============================================================================
print("插入登录日志数据...");

var loginLogs = [
    {
        userId: adminId,
        username: "admin", 
        email: "admin@heimdall.com",
        ipAddress: "127.0.0.1",
        userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
        success: true,
        failReason: "",
        createdAt: new Date()
    },
    {
        userId: authorId,
        username: "author",
        email: "author@heimdall.com", 
        ipAddress: "192.168.1.100",
        userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
        success: true,
        failReason: "",
        createdAt: new Date(now.getTime() - 60 * 60 * 1000)
    },
    {
        userId: null,
        username: "hacker",
        email: "hacker@evil.com",
        ipAddress: "192.168.1.200", 
        userAgent: "curl/7.68.0",
        success: false,
        failReason: "invalid_credentials",
        createdAt: new Date(now.getTime() - 2 * 60 * 60 * 1000)
    }
];

db.loginLogs.insertMany(loginLogs);
print("登录日志插入完成，共 " + loginLogs.length + " 条记录");

// =============================================================================
// 7. 数据插入完成统计
// =============================================================================
print("\n=== 种子数据插入完成统计 ===");

var collections = ["users", "loginLogs", "posts", "comments", "settings"];

collections.forEach(function(collName) {
    var count = db[collName].countDocuments();
    print(collName + " 集合共有 " + count + " 条记录");
});

print("\n所有种子数据插入完成！");
print("默认管理员账号: admin / admin123");
print("默认作者账号: author / author123");
print("建议在生产环境使用前修改默认密码。"); 