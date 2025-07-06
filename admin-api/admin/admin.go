package main

import (
	"flag"
	"fmt"

	"github.com/heimdall-api/admin-api/admin/internal/config"
	"github.com/heimdall-api/admin-api/admin/internal/handler"
	"github.com/heimdall-api/admin-api/admin/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/admin-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	
	// 添加全局中间件
	// 1. IP限流中间件 (最先执行)
	if ctx.IPRateLimitMiddleware != nil {
		server.Use(ctx.IPRateLimitMiddleware.Handle)
	}
	
	// 2. JWT黑名单检查中间件 (JWT验证之前)
	if ctx.JWTBlacklistMiddleware != nil {
		server.Use(ctx.JWTBlacklistMiddleware.Handle)
	}
	
	// 3. 审计中间件 (记录所有操作)
	if ctx.AuditMiddleware != nil {
		server.Use(ctx.AuditMiddleware.Handle)
	}
	
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
