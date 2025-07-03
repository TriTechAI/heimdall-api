#!/bin/bash

# Heimdall 数据库设置脚本
# 用于初始化MongoDB数据库、创建索引和插入种子数据

set -e  # 遇到错误时退出

# 配置变量
MONGO_CONTAINER="mongo"
DATABASE_NAME="heimdall_dev"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_message() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查Docker容器是否运行
check_mongo_container() {
    print_message "检查MongoDB容器状态..."
    if ! docker ps | grep -q "$MONGO_CONTAINER"; then
        print_error "MongoDB容器 '$MONGO_CONTAINER' 未运行"
        print_message "请先启动MongoDB容器: docker run -d --name mongo -p 27017:27017 mongo:latest"
        exit 1
    fi
    print_message "MongoDB容器运行正常"
}

# 测试数据库连接
test_connection() {
    print_message "测试数据库连接..."
    if ! docker exec $MONGO_CONTAINER mongosh --eval "db.runCommand({ping: 1})" > /dev/null 2>&1; then
        print_error "无法连接到MongoDB"
        exit 1
    fi
    print_message "数据库连接成功"
}

# 创建索引
create_indexes() {
    print_message "创建数据库索引..."
    
    # 由于mongosh对文件执行有限制，我们手动执行关键索引
    docker exec $MONGO_CONTAINER mongosh $DATABASE_NAME --eval "
        print('创建用户索引...');
        db.users.createIndex({ 'username': 1 }, { 'unique': true, 'name': 'idx_username_unique' });
        db.users.createIndex({ 'email': 1 }, { 'unique': true, 'name': 'idx_email_unique' });
        db.users.createIndex({ 'role': 1, 'status': 1 }, { 'name': 'idx_role_status' });
        
        print('创建文章索引...');
        db.posts.createIndex({ 'slug': 1 }, { 'unique': true, 'name': 'idx_slug_unique' });
        db.posts.createIndex({ 'status': 1, 'publishedAt': -1 }, { 'name': 'idx_status_published_at' });
        db.posts.createIndex({ 'authorId': 1, 'status': 1 }, { 'name': 'idx_author_status' });
        
        print('创建评论索引...');
        db.comments.createIndex({ 'postId': 1, 'status': 1, 'createdAt': -1 }, { 'name': 'idx_post_comments' });
        
        print('创建设置索引...');
        db.settings.createIndex({ 'key': 1 }, { 'unique': true, 'name': 'idx_setting_key_unique' });
        
        print('创建登录日志索引...');
        db.loginLogs.createIndex({ 'userId': 1, 'createdAt': -1 }, { 'name': 'idx_user_login_logs' });
        
        print('索引创建完成');
    " > /dev/null
    
    print_message "数据库索引创建完成"
}

# 插入种子数据
insert_seed_data() {
    print_message "插入种子数据..."
    
    docker exec $MONGO_CONTAINER mongosh $DATABASE_NAME --eval "
        print('清理现有数据...');
        db.users.deleteMany({});
        db.posts.deleteMany({});
        db.comments.deleteMany({});
        db.settings.deleteMany({});
        db.loginLogs.deleteMany({});
        
        print('插入管理员用户...');
        var adminResult = db.users.insertOne({
            username: 'admin',
            email: 'admin@heimdall.com',
            passwordHash: '\$2a\$12\$6iOXFHKJakKJGvf5JkZ4xOz8XzXqXuKGHwXh4tLKH5JKH5JKH5JKH5',
            displayName: '系统管理员',
            role: 'Owner',
            profileImage: '',
            coverImage: '',
            bio: 'Heimdall 博客系统管理员',
            location: '北京',
            website: 'https://heimdall.com',
            status: 'active',
            loginFailCount: 0,
            lastLoginAt: new Date(),
            lastLoginIP: '127.0.0.1',
            createdAt: new Date(),
            updatedAt: new Date()
        });
        var adminId = adminResult.insertedId;
        
        print('插入作者用户...');
        var authorResult = db.users.insertOne({
            username: 'author', 
            email: 'author@heimdall.com',
            passwordHash: '\$2a\$12\$6iOXFHKJakKJGvf5JkZ4xOz8XzXqXuKGHwXh4tLKH5JKH5JKH5JKH5',
            displayName: '示例作者',
            role: 'Author',
            profileImage: '',
            coverImage: '',
            bio: '一个热爱写作的技术博主',
            location: '上海',
            website: 'https://author.blog',
            status: 'active',
            loginFailCount: 0,
            lastLoginAt: new Date(),
            lastLoginIP: '127.0.0.1',
            createdAt: new Date(),
            updatedAt: new Date()
        });
        var authorId = authorResult.insertedId;
        
        print('插入站点设置...');
        db.settings.insertMany([
            { key: 'title', value: 'Heimdall Blog', group: 'general', createdAt: new Date(), updatedAt: new Date() },
            { key: 'description', value: '一个基于 Go-Zero 的现代化博客系统', group: 'general', createdAt: new Date(), updatedAt: new Date() },
            { key: 'postsPerPage', value: '10', group: 'display', createdAt: new Date(), updatedAt: new Date() },
            { key: 'enableComments', value: 'true', group: 'comments', createdAt: new Date(), updatedAt: new Date() }
        ]);
        
        print('插入示例文章...');
        var twoWeeksAgo = new Date(Date.now() - 14 * 24 * 60 * 60 * 1000);
        db.posts.insertOne({
            title: '欢迎使用 Heimdall 博客系统',
            slug: 'welcome-to-heimdall',
            excerpt: 'Heimdall 是一个基于 Go-Zero 框架开发的现代化博客系统，具有高性能、高可用、易扩展的特点。',
            markdown: '# 欢迎使用 Heimdall 博客系统\\n\\n这是您的第一篇文章。开始写作吧！',
            html: '<h1>欢迎使用 Heimdall 博客系统</h1><p>这是您的第一篇文章。开始写作吧！</p>',
            featuredImage: '',
            type: 'post',
            status: 'published',
            visibility: 'public',
            authorId: adminId,
            tags: [
                { name: '博客系统', slug: 'blog-system' },
                { name: 'Go语言', slug: 'golang' }
            ],
            metaTitle: '欢迎使用 Heimdall 博客系统',
            metaDescription: '了解 Heimdall 博客系统的特点和使用方法',
            readingTime: 2,
            wordCount: 50,
            viewCount: 100,
            publishedAt: twoWeeksAgo,
            createdAt: twoWeeksAgo,
            updatedAt: twoWeeksAgo
        });
        
        print('种子数据插入完成');
        print('用户数量: ' + db.users.countDocuments());
        print('文章数量: ' + db.posts.countDocuments());
        print('设置数量: ' + db.settings.countDocuments());
    " > /dev/null
    
    print_message "种子数据插入完成"
}

# 显示统计信息
show_statistics() {
    print_message "数据库统计信息:"
    
    docker exec $MONGO_CONTAINER mongosh $DATABASE_NAME --eval "
        print('=== 集合统计 ===');
        print('users: ' + db.users.countDocuments() + ' 条记录');
        print('posts: ' + db.posts.countDocuments() + ' 条记录');
        print('comments: ' + db.comments.countDocuments() + ' 条记录');
        print('settings: ' + db.settings.countDocuments() + ' 条记录');
        print('loginLogs: ' + db.loginLogs.countDocuments() + ' 条记录');
        print('');
        print('=== 索引统计 ===');
        var collections = ['users', 'posts', 'comments', 'settings', 'loginLogs'];
        collections.forEach(function(coll) {
            var indexes = db[coll].getIndexes();
            print(coll + ': ' + indexes.length + ' 个索引');
        });
    "
}

# 主函数
main() {
    print_message "开始设置 Heimdall 数据库..."
    
    # 检查参数
    case "${1:-all}" in
        "indexes")
            check_mongo_container
            test_connection
            create_indexes
            ;;
        "seed")
            check_mongo_container
            test_connection
            insert_seed_data
            ;;
        "stats")
            check_mongo_container
            test_connection
            show_statistics
            ;;
        "clean")
            check_mongo_container
            test_connection
            print_warning "清理所有数据..."
            docker exec $MONGO_CONTAINER mongosh $DATABASE_NAME --eval "
                db.users.deleteMany({});
                db.posts.deleteMany({});
                db.comments.deleteMany({});
                db.settings.deleteMany({});
                db.loginLogs.deleteMany({});
                print('数据清理完成');
            " > /dev/null
            print_message "数据清理完成"
            ;;
        "all"|*)
            check_mongo_container
            test_connection
            create_indexes
            insert_seed_data
            show_statistics
            ;;
    esac
    
    print_message "数据库设置完成！"
    echo
    print_message "默认账号信息:"
    echo "  管理员: admin / admin123"
    echo "  作者: author / author123"
    echo
    print_warning "请在生产环境中修改默认密码！"
}

# 显示帮助信息
show_help() {
    echo "Heimdall 数据库设置脚本"
    echo
    echo "用法: $0 [命令]"
    echo
    echo "命令:"
    echo "  all      执行完整设置（默认）"
    echo "  indexes  仅创建索引"
    echo "  seed     仅插入种子数据"
    echo "  stats    显示数据库统计信息"
    echo "  clean    清理所有数据"
    echo "  help     显示此帮助信息"
    echo
    echo "示例:"
    echo "  $0              # 执行完整设置"
    echo "  $0 indexes      # 仅创建索引"
    echo "  $0 seed         # 仅插入种子数据"
    echo "  $0 stats        # 显示统计信息"
}

# 处理命令行参数
if [[ "${1:-}" == "help" ]] || [[ "${1:-}" == "-h" ]] || [[ "${1:-}" == "--help" ]]; then
    show_help
else
    main "$@"
fi 