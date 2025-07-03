// MongoDB 索引初始化脚本
// 根据设计文档 design/DATA-MODEL-DESIGN.md 创建所有必要的索引

// 使用开发数据库
use heimdall_dev;

print("正在创建索引...");

// =============================================================================
// 1. users 集合索引
// =============================================================================
print("创建 users 集合索引...");

// 唯一索引：用户名
db.users.createIndex({ "username": 1 }, { "unique": true, "name": "idx_username_unique" });

// 唯一索引：邮箱
db.users.createIndex({ "email": 1 }, { "unique": true, "name": "idx_email_unique" });

// 复合索引：角色和状态
db.users.createIndex({ "role": 1, "status": 1 }, { "name": "idx_role_status" });

// 索引：账户锁定时间（用于解锁已过期的账号）
db.users.createIndex({ "lockedUntil": 1 }, { "name": "idx_locked_until" });

// 索引：最后登录时间
db.users.createIndex({ "lastLoginAt": -1 }, { "name": "idx_last_login_at" });

// 索引：创建时间
db.users.createIndex({ "createdAt": -1 }, { "name": "idx_users_created_at" });

print("users 集合索引创建完成");

// =============================================================================
// 2. loginLogs 集合索引
// =============================================================================
print("创建 loginLogs 集合索引...");

// 复合索引：用户ID和创建时间
db.loginLogs.createIndex({ "userId": 1, "createdAt": -1 }, { "name": "idx_user_login_logs" });

// 复合索引：IP地址和创建时间
db.loginLogs.createIndex({ "ipAddress": 1, "createdAt": -1 }, { "name": "idx_ip_login_logs" });

// 复合索引：成功状态和创建时间
db.loginLogs.createIndex({ "success": 1, "createdAt": -1 }, { "name": "idx_success_login_logs" });

// 索引：用户名（用于按用户名查询登录记录）
db.loginLogs.createIndex({ "username": 1 }, { "name": "idx_username_login_logs" });

print("loginLogs 集合索引创建完成");

// =============================================================================
// 3. posts 集合索引
// =============================================================================
print("创建 posts 集合索引...");

// 唯一索引：slug
db.posts.createIndex({ "slug": 1 }, { "unique": true, "name": "idx_slug_unique" });

// 复合索引：状态和发布时间（用于查询已发布的文章列表）
db.posts.createIndex({ "status": 1, "publishedAt": -1 }, { "name": "idx_status_published_at" });

// 复合索引：作者ID和状态
db.posts.createIndex({ "authorId": 1, "status": 1 }, { "name": "idx_author_status" });

// 索引：标签slug（用于按标签查询）
db.posts.createIndex({ "tags.slug": 1 }, { "name": "idx_tags_slug" });

// 复合索引：类型和状态（用于区分文章和页面）
db.posts.createIndex({ "type": 1, "status": 1 }, { "name": "idx_type_status" });

// 索引：浏览量（用于热门文章排序）
db.posts.createIndex({ "viewCount": -1 }, { "name": "idx_view_count" });

// 全文搜索索引
db.posts.createIndex({ 
    "title": "text", 
    "excerpt": "text",
    "markdown": "text"
}, { 
    "name": "idx_full_text_search",
    "default_language": "none",
    "weights": {
        "title": 10,
        "excerpt": 5,
        "markdown": 1
    }
});

// 索引：创建时间
db.posts.createIndex({ "createdAt": -1 }, { "name": "idx_posts_created_at" });

// 索引：更新时间
db.posts.createIndex({ "updatedAt": -1 }, { "name": "idx_posts_updated_at" });

// 复合索引：可见性和状态
db.posts.createIndex({ "visibility": 1, "status": 1 }, { "name": "idx_visibility_status" });

print("posts 集合索引创建完成");

// =============================================================================
// 4. comments 集合索引
// =============================================================================
print("创建 comments 集合索引...");

// 复合索引：文章ID、状态和创建时间（用于查询一篇文章下的评论）
db.comments.createIndex({ "postId": 1, "status": 1, "createdAt": -1 }, { "name": "idx_post_comments" });

// 索引：状态（用于后台管理审核评论）
db.comments.createIndex({ "status": 1 }, { "name": "idx_comment_status" });

// 索引：父评论ID（用于查询回复评论）
db.comments.createIndex({ "parentId": 1 }, { "name": "idx_parent_comments" });

// 索引：IP地址（用于反垃圾检测）
db.comments.createIndex({ "ipAddress": 1 }, { "name": "idx_comment_ip" });

// 索引：作者ID（如果是注册用户评论）
db.comments.createIndex({ "authorId": 1 }, { "name": "idx_comment_author" });

// 索引：创建时间
db.comments.createIndex({ "createdAt": -1 }, { "name": "idx_comments_created_at" });

// 索引：点赞数（用于热门评论排序）
db.comments.createIndex({ "likeCount": -1 }, { "name": "idx_comment_likes" });

print("comments 集合索引创建完成");

// =============================================================================
// 5. settings 集合索引
// =============================================================================
print("创建 settings 集合索引...");

// 唯一索引：配置键
db.settings.createIndex({ "key": 1 }, { "unique": true, "name": "idx_setting_key_unique" });

// 索引：配置组（便于按组管理）
db.settings.createIndex({ "group": 1 }, { "name": "idx_setting_group" });

print("settings 集合索引创建完成");

// =============================================================================
// 6. media 集合索引
// =============================================================================
print("创建 media 集合索引...");

// 索引：文件类型
db.media.createIndex({ "type": 1 }, { "name": "idx_media_type" });

// 索引：MIME类型
db.media.createIndex({ "mimeType": 1 }, { "name": "idx_media_mime_type" });

// 索引：创建时间
db.media.createIndex({ "createdAt": -1 }, { "name": "idx_media_created_at" });

// 索引：文件大小（用于存储管理）
db.media.createIndex({ "size": -1 }, { "name": "idx_media_size" });

// 索引：文件URL（确保快速查找）
db.media.createIndex({ "url": 1 }, { "name": "idx_media_url" });

print("media 集合索引创建完成");

// =============================================================================
// 显示索引创建结果
// =============================================================================
print("\n=== 索引创建完成统计 ===");

var collections = ["users", "loginLogs", "posts", "comments", "settings", "media"];

collections.forEach(function(collName) {
    var indexes = db[collName].getIndexes();
    print(collName + " 集合共有 " + indexes.length + " 个索引:");
    indexes.forEach(function(index) {
        print("  - " + index.name + ": " + JSON.stringify(index.key));
    });
    print("");
});

print("所有索引创建完成！");
print("建议在生产环境运行前检查索引的性能影响。"); 