package e2e

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TestDataManager 测试数据管理器
type TestDataManager struct {
	mongoClient *mongo.Client
	database    *mongo.Database
	config      *TestConfig
}

// NewTestDataManager 创建测试数据管理器
func NewTestDataManager(config *TestConfig) (*TestDataManager, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 连接MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.Database.MongoURL))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// 测试连接
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	database := client.Database(config.Database.TestDB)

	return &TestDataManager{
		mongoClient: client,
		database:    database,
		config:      config,
	}, nil
}

// SetupTestData 设置测试数据
func (tdm *TestDataManager) SetupTestData() error {
	ctx := context.Background()

	// 创建测试用户
	if err := tdm.createTestUser(ctx); err != nil {
		return fmt.Errorf("failed to create test user: %v", err)
	}

	log.Println("✅ 测试数据设置完成")
	return nil
}

// CleanupTestData 清理测试数据
func (tdm *TestDataManager) CleanupTestData() error {
	ctx := context.Background()

	// 清理所有测试相关的集合
	collections := []string{
		"users",
		"posts", 
		"pages",
		"login_logs",
		"tags",
	}

	for _, collName := range collections {
		collection := tdm.database.Collection(collName)
		
		// 删除测试数据（保留非测试数据）
		filter := bson.M{
			"$or": []bson.M{
				{"username": bson.M{"$regex": "^test"}},
				{"email": bson.M{"$regex": "@example.com$"}},
				{"title": bson.M{"$regex": "测试|Test|E2E"}},
				{"slug": bson.M{"$regex": "test|e2e"}},
			},
		}
		
		result, err := collection.DeleteMany(ctx, filter)
		if err != nil {
			log.Printf("⚠️  清理集合 %s 时出错: %v", collName, err)
		} else {
			log.Printf("🗑️  清理集合 %s: 删除了 %d 条记录", collName, result.DeletedCount)
		}
	}

	return nil
}

// Close 关闭数据库连接
func (tdm *TestDataManager) Close() error {
	if tdm.mongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return tdm.mongoClient.Disconnect(ctx)
	}
	return nil
}

// createTestUser 创建测试用户
func (tdm *TestDataManager) createTestUser(ctx context.Context) error {
	usersCollection := tdm.database.Collection("users")

	// 检查测试用户是否已存在
	var existingUser bson.M
	err := usersCollection.FindOne(ctx, bson.M{"username": tdm.config.TestUser.Username}).Decode(&existingUser)
	if err == nil {
		log.Printf("✅ 测试用户 %s 已存在", tdm.config.TestUser.Username)
		return nil
	}

	// 创建测试用户文档
	now := time.Now()
	testUser := bson.M{
		"username":     tdm.config.TestUser.Username,
		"email":        tdm.config.TestUser.Email,
		"password":     "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // bcrypt hash of "testpass123"
		"displayName":  "Test User",
		"role":         tdm.config.TestUser.Role,
		"status":       "active",
		"profileImage": "",
		"bio":          "E2E测试用户",
		"location":     "",
		"website":      "",
		"twitter":      "",
		"facebook":     "",
		"loginFailCount": 0,
		"lockedUntil":    nil,
		"lastLoginAt":    nil,
		"lastLoginIP":    "",
		"createdAt":      now,
		"updatedAt":      now,
	}

	// 插入测试用户
	_, err = usersCollection.InsertOne(ctx, testUser)
	if err != nil {
		return fmt.Errorf("failed to insert test user: %v", err)
	}

	log.Printf("✅ 创建测试用户: %s", tdm.config.TestUser.Username)
	return nil
}

// CreateTestPosts 创建测试文章
func (tdm *TestDataManager) CreateTestPosts(count int) error {
	ctx := context.Background()
	postsCollection := tdm.database.Collection("posts")

	now := time.Now()
	posts := make([]interface{}, count)

	for i := 0; i < count; i++ {
		posts[i] = bson.M{
			"title":           fmt.Sprintf("测试文章 %d", i+1),
			"slug":            fmt.Sprintf("test-post-%d", i+1),
			"excerpt":         fmt.Sprintf("这是第 %d 篇测试文章的摘要", i+1),
			"markdown":        fmt.Sprintf("# 测试文章 %d\n\n这是第 %d 篇测试文章的内容。", i+1, i+1),
			"html":            fmt.Sprintf("<h1>测试文章 %d</h1><p>这是第 %d 篇测试文章的内容。</p>", i+1, i+1),
			"featuredImage":   "",
			"type":            "post",
			"status":          "published",
			"visibility":      "public",
			"authorId":        "test_author_id",
			"tags":            []bson.M{{"name": "测试", "slug": "test"}},
			"metaTitle":       fmt.Sprintf("测试文章 %d - SEO标题", i+1),
			"metaDescription": fmt.Sprintf("这是第 %d 篇测试文章的SEO描述", i+1),
			"canonicalUrl":    "",
			"readingTime":     2,
			"wordCount":       100,
			"viewCount":       int64(i * 10),
			"publishedAt":     now.Add(-time.Duration(i) * time.Hour),
			"createdAt":       now.Add(-time.Duration(i) * time.Hour),
			"updatedAt":       now.Add(-time.Duration(i) * time.Hour),
		}
	}

	// 批量插入测试文章
	_, err := postsCollection.InsertMany(ctx, posts)
	if err != nil {
		return fmt.Errorf("failed to insert test posts: %v", err)
	}

	log.Printf("✅ 创建了 %d 篇测试文章", count)
	return nil
}

// GetTestUserID 获取测试用户ID
func (tdm *TestDataManager) GetTestUserID() (string, error) {
	ctx := context.Background()
	usersCollection := tdm.database.Collection("users")

	var user bson.M
	err := usersCollection.FindOne(ctx, bson.M{"username": tdm.config.TestUser.Username}).Decode(&user)
	if err != nil {
		return "", fmt.Errorf("test user not found: %v", err)
	}

	if id, ok := user["_id"]; ok {
		return fmt.Sprintf("%v", id), nil
	}

	return "", fmt.Errorf("test user ID not found")
}