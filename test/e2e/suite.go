package e2e

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// TestSuite E2E测试套件
type TestSuite struct {
	config         *TestConfig
	dataManager    *TestDataManager
	adminProcess   *os.Process
	publicProcess  *os.Process
	setupComplete  bool
}

// NewTestSuite 创建新的测试套件
func NewTestSuite() *TestSuite {
	config := GetTestConfig()
	return &TestSuite{
		config: config,
	}
}

// Setup 设置测试环境
func (ts *TestSuite) Setup() error {
	if ts.setupComplete {
		return nil
	}

	log.Println("🚀 开始设置E2E测试环境...")

	// 初始化数据管理器
	var err error
	ts.dataManager, err = NewTestDataManager(ts.config)
	if err != nil {
		return fmt.Errorf("failed to create test data manager: %v", err)
	}

	// 清理旧的测试数据
	if err := ts.dataManager.CleanupTestData(); err != nil {
		log.Printf("⚠️  清理旧测试数据时出错: %v", err)
	}

	// 设置测试数据
	if err := ts.dataManager.SetupTestData(); err != nil {
		return fmt.Errorf("failed to setup test data: %v", err)
	}

	// 启动服务
	if err := ts.startServices(); err != nil {
		return fmt.Errorf("failed to start services: %v", err)
	}

	// 等待服务启动
	if err := ts.waitForServices(); err != nil {
		return fmt.Errorf("services failed to start: %v", err)
	}

	ts.setupComplete = true
	log.Println("✅ E2E测试环境设置完成")
	return nil
}

// Cleanup 清理测试环境
func (ts *TestSuite) Cleanup() {
	log.Println("🧹 开始清理E2E测试环境...")

	// 停止服务
	ts.stopServices()

	// 清理测试数据
	if ts.dataManager != nil {
		if err := ts.dataManager.CleanupTestData(); err != nil {
			log.Printf("⚠️  清理测试数据时出错: %v", err)
		}
		ts.dataManager.Close()
	}

	log.Println("✅ E2E测试环境清理完成")
}

// startServices 启动服务
func (ts *TestSuite) startServices() error {
	// 获取项目根目录
	rootDir, err := ts.getProjectRoot()
	if err != nil {
		return fmt.Errorf("failed to get project root: %v", err)
	}

	// 启动admin-api服务
	if err := ts.startAdminService(rootDir); err != nil {
		return fmt.Errorf("failed to start admin service: %v", err)
	}

	// 启动public-api服务
	if err := ts.startPublicService(rootDir); err != nil {
		return fmt.Errorf("failed to start public service: %v", err)
	}

	return nil
}

// startAdminService 启动管理服务
func (ts *TestSuite) startAdminService(rootDir string) error {
	adminDir := filepath.Join(rootDir, "admin-api", "admin")
	configPath := filepath.Join(adminDir, "etc", "admin-api.yaml")

	cmd := exec.Command("go", "run", ".", "-f", configPath)
	cmd.Dir = adminDir
	cmd.Env = append(os.Environ(), 
		"HEIMDALL_ENV=test",
		fmt.Sprintf("MONGO_URL=%s", ts.config.Database.MongoURL),
		fmt.Sprintf("REDIS_URL=%s", ts.config.Database.RedisURL),
	)

	// 启动进程
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start admin service: %v", err)
	}

	ts.adminProcess = cmd.Process
	log.Printf("✅ Admin服务已启动 (PID: %d)", ts.adminProcess.Pid)
	return nil
}

// startPublicService 启动公开服务
func (ts *TestSuite) startPublicService(rootDir string) error {
	publicDir := filepath.Join(rootDir, "public-api", "public")
	configPath := filepath.Join(publicDir, "etc", "public-api.yaml")

	cmd := exec.Command("go", "run", ".", "-f", configPath)
	cmd.Dir = publicDir
	cmd.Env = append(os.Environ(),
		"HEIMDALL_ENV=test",
		fmt.Sprintf("MONGO_URL=%s", ts.config.Database.MongoURL),
		fmt.Sprintf("REDIS_URL=%s", ts.config.Database.RedisURL),
	)

	// 启动进程
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start public service: %v", err)
	}

	ts.publicProcess = cmd.Process
	log.Printf("✅ Public服务已启动 (PID: %d)", ts.publicProcess.Pid)
	return nil
}

// stopServices 停止服务
func (ts *TestSuite) stopServices() {
	if ts.adminProcess != nil {
		if err := ts.adminProcess.Signal(syscall.SIGTERM); err != nil {
			log.Printf("⚠️  停止Admin服务时出错: %v", err)
			ts.adminProcess.Kill()
		} else {
			log.Printf("🛑 Admin服务已停止 (PID: %d)", ts.adminProcess.Pid)
		}
		ts.adminProcess = nil
	}

	if ts.publicProcess != nil {
		if err := ts.publicProcess.Signal(syscall.SIGTERM); err != nil {
			log.Printf("⚠️  停止Public服务时出错: %v", err)
			ts.publicProcess.Kill()
		} else {
			log.Printf("🛑 Public服务已停止 (PID: %d)", ts.publicProcess.Pid)
		}
		ts.publicProcess = nil
	}
}

// waitForServices 等待服务启动
func (ts *TestSuite) waitForServices() error {
	client := NewTestClient()

	log.Println("⏳ 等待服务启动...")

	// 等待Admin服务
	if err := client.WaitForService(ts.config.AdminAPIURL, 30); err != nil {
		return fmt.Errorf("admin service not ready: %v", err)
	}
	log.Println("✅ Admin服务已就绪")

	// 等待Public服务
	if err := client.WaitForService(ts.config.PublicAPIURL, 30); err != nil {
		return fmt.Errorf("public service not ready: %v", err)
	}
	log.Println("✅ Public服务已就绪")

	return nil
}

// getProjectRoot 获取项目根目录
func (ts *TestSuite) getProjectRoot() (string, error) {
	// 从当前目录开始向上查找go.mod文件
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("project root not found")
}

// RunE2ETests 运行E2E测试主函数
func RunE2ETests(m *testing.M) {
	suite := NewTestSuite()

	// 设置测试环境
	if err := suite.Setup(); err != nil {
		log.Fatalf("❌ 测试环境设置失败: %v", err)
	}

	// 运行测试
	log.Println("🧪 开始运行E2E测试...")
	code := m.Run()

	// 清理测试环境
	suite.Cleanup()

	// 退出
	if code == 0 {
		log.Println("✅ 所有E2E测试通过")
	} else {
		log.Println("❌ 部分E2E测试失败")
	}
	
	os.Exit(code)
}