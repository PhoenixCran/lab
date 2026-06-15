// 程序入口文件
// 负责初始化应用配置、数据库连接、服务组件和 HTTP 服务器
// 实现优雅的启动和关闭机制
package main

import (
	"context"      // 用于上下文管理，支持取消和超时控制
	"database/sql" // 标准数据库 SQL 包
	"log/slog"     // 结构化日志记录
	"net/http"     // HTTP 服务器和客户端
	"os"           // 操作系统相关功能，如退出进程
	"os/signal"    // 信号处理，用于捕获系统信号
	"syscall"      // 系统调用，定义信号常量
	"time"         // 时间相关功能

	_ "github.com/go-sql-driver/mysql" // MySQL 驱动（匿名导入以注册驱动）

	"lab/config"              // 配置管理
	"lab/internal/handler"    // HTTP 请求处理器
	"lab/internal/repository" // 数据仓库层
	"lab/internal/router"     // 路由配置
	"lab/internal/service"    // 业务逻辑服务层
)

// main 函数：应用程序入口点
func main() {
	// 加载配置文件（从环境变量读取）
	cfg := config.Load()

	// 打开 MySQL 数据库连接
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		// 如果打开数据库失败，记录错误并退出
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	// 确保在函数退出时关闭数据库连接
	defer db.Close()

	// 设置数据库连接池最大打开连接数为 25
	db.SetMaxOpenConns(25)
	// 设置数据库连接池最大空闲连接数为 10
	db.SetMaxIdleConns(10)
	// 设置数据库连接最长生命周期为 5 分钟
	db.SetConnMaxLifetime(5 * time.Minute)

	// 创建带超时的上下文（5 秒），用于数据库连接测试
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// 确保取消函数被调用以释放资源
	defer cancel()
	// 测试数据库连接是否可用
	if err := db.PingContext(ctx); err != nil {
		// 如果数据库连接失败，记录错误并退出
		slog.Error("failed to ping database", "error", err)
		os.Exit(1)
	}
	// 记录数据库连接成功信息
	slog.Info("database connected", "host", cfg.DBHost, "port", cfg.DBPort, "db", cfg.DBName)

	// 初始化用户数据仓库（负责用户数据的 CRUD 操作）
	userRepo := repository.NewUserRepo(db)
	// 初始化图书数据仓库（负责图书数据的 CRUD 操作）
	bookRepo := repository.NewBookRepo(db)

	// 创建用户服务实例（封装用户业务逻辑）
	userSvc := service.NewUserService(userRepo)
	// 创建图书服务实例（封装图书业务逻辑）
	bookSvc := service.NewBookService(bookRepo)

	// 创建用户 HTTP 处理器（处理用户相关的 HTTP 请求）
	userH := handler.NewUserHandler(userSvc)
	// 创建图书 HTTP 处理器（处理图书相关的 HTTP 请求）
	bookH := handler.NewBookHandler(bookSvc)

	// 设置路由配置，返回完整的 HTTP 处理程序（包含中间件）
	h := router.Setup(db, userH, bookH)

	// 配置 HTTP 服务器
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort, // 监听端口
		Handler:      h,                    // 请求处理程序
		ReadTimeout:  10 * time.Second,     // 读取超时时间
		WriteTimeout: 10 * time.Second,     // 写入超时时间
		IdleTimeout:  30 * time.Second,     // 空闲超时时间
	}

	// 在后台协程中启动 HTTP 服务器
	go func() {
		// 记录服务器启动信息
		slog.Info("server starting", "port", cfg.ServerPort)
		// 开始监听和处理请求
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// 如果是非正常关闭错误，记录并退出
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// 创建信号接收通道，用于捕获中断信号
	quit := make(chan os.Signal, 1)
	// 监听从键盘输入的 Ctrl+C (SIGINT) 和系统终止信号 (SIGTERM)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	// 阻塞等待接收关闭信号
	<-quit

	// 记录接收到关闭信号
	slog.Info("shutting down server...")
	// 创建带 10 秒超时的上下文用于优雅关闭
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	// 确保取消函数被调用以释放资源
	defer shutdownCancel()

	// 优雅关闭服务器（等待正在处理的请求完成）
	if err := srv.Shutdown(shutdownCtx); err != nil {
		// 如果关闭失败，记录强制关闭的错误
		slog.Error("server forced to shutdown", "error", err)
	}
	// 记录服务器已成功停止
	slog.Info("server stopped")
}
