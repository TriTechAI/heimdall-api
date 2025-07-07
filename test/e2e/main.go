package main

import (
	"flag"
	"os"
	"testing"

	"github.com/heimdall-api/test/e2e"
)

func main() {
	// 解析命令行参数
	flag.Parse()

	// 运行E2E测试
	e2e.RunE2ETests(&testing.M{})
}