// Package middleware HTTP 中间件包
// 提供 HTTP 请求处理的可复用组件
// 包含日志记录、异常恢复和内容类型验证等功能
package middleware

import (
	"log/slog"      // 结构化日志记录
	"net/http"      // HTTP 服务器和客户端
	"runtime/debug" // 运行时调试信息，用于获取堆栈跟踪
	"time"          // 时间相关功能，用于计算请求耗时
)

// responseWriter 是一个包装器，用于拦截响应状态码
// 继承自 http.ResponseWriter，增加状态码捕获功能
type responseWriter struct {
	http.ResponseWriter // 嵌入原始响应写入器
	statusCode int      // 记录响应状态码，默认值为 0
}

// WriteHeader 重写父类方法，在发送响应头之前记录状态码
// 这是 HTTP 协议规定的方法，仅在第一次调用时生效
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code                    // 保存状态码到包装器中
	rw.ResponseWriter.WriteHeader(code)     // 调用原始响应写入器发送状态码
}

// Logging 日志中间件
// 记录每个请求的详细信息，包括方法、路径、状态码、耗时等
// 参数 next: 下一个处理器（被包装的处理器）
// 返回：包装后的处理器
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()                                    // 记录请求开始时间
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK} // 创建包装器，默认状态码为 200
		next.ServeHTTP(rw, r)                                  // 调用下一个处理器
		// 记录请求日志（在响应完成后执行）
		slog.Info("request",
			"method", r.Method,             // HTTP 请求方法 (GET/POST 等)
			"path", r.URL.Path,            // 请求的 URL 路径
			"status", rw.statusCode,       // 响应状态码（由包装器捕获）
			"duration", time.Since(start).String(), // 请求处理耗时
			"remote", r.RemoteAddr,        // 客户端远程地址
		)
	})
}

// Recovery 异常恢复中间件
// 捕获 panic 并转换为 HTTP 500 错误，防止服务器崩溃
// 参数 next: 下一个处理器（被包装的处理器）
// 返回：包装后的处理器
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 使用 defer 确保在处理完成后执行恢复逻辑
		defer func() {
			if rec := recover(); rec != nil { // 如果有 panic 发生，recover() 会捕获它
				// 记录 panic 详情（错误信息和堆栈跟踪）
				slog.Error("panic recovered",
					"error", rec,              // 捕获的 panic 值
					"stack", string(debug.Stack()), // 当前堆栈跟踪
				)
				// 返回 500 内部服务器错误给客户端
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r) // 调用下一个处理器
	})
}

// JSONContentType JSON 内容类型验证中间件
// 确保 POST/PUT/PATCH 请求的 Content-Type 为 application/json
// 参数 next: 下一个处理器（被包装的处理器）
// 返回：包装后的处理器
func JSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 仅对修改数据的请求进行验证
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			// 检查 Content-Type 是否为 application/json
			if r.Header.Get("Content-Type") != "application/json" {
				// 如果不是 JSON 格式，返回 415 不支持的媒体类型错误
				http.Error(w, `{"error":"content-type must be application/json"}`, http.StatusUnsupportedMediaType)
				return // 直接返回，不再调用下一个处理器
			}
		}
		next.ServeHTTP(w, r) // 验证通过，继续调用下一个处理器
	})
}
