// 路由配置文件
// 负责设置所有API路由和中间件
// 实现了RESTful API路由结构和中间件链
package router

import (
	"database/sql"
	"net/http"

	"lab/internal/handler"
	"lab/internal/middleware"
)

func Setup(db *sql.DB, userH *handler.UserHandler, bookH *handler.BookHandler) http.Handler {
	mux := http.NewServeMux()

	// API 路由
	apiMux := http.NewServeMux()

	// Health check
	apiMux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			http.Error(w, `{"error":"database unavailable"}`, http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// User routes
	apiMux.HandleFunc("POST /api/users", userH.Create)
	apiMux.HandleFunc("GET /api/users", userH.List)
	apiMux.HandleFunc("GET /api/users/{id}", userH.Get)
	apiMux.HandleFunc("PUT /api/users/{id}", userH.Update)
	apiMux.HandleFunc("DELETE /api/users/{id}", userH.Delete)

	// Book routes
	apiMux.HandleFunc("POST /api/books", bookH.Create)
	apiMux.HandleFunc("GET /api/books", bookH.List)
	apiMux.HandleFunc("GET /api/books/{id}", bookH.Get)
	apiMux.HandleFunc("PUT /api/books/{id}", bookH.Update)
	apiMux.HandleFunc("DELETE /api/books/{id}", bookH.Delete)

	// 为 API 路由应用中间件
	var apiHandler http.Handler = apiMux
	apiHandler = middleware.Logging(apiHandler)
	apiHandler = middleware.Recovery(apiHandler)
	apiHandler = middleware.JSONContentType(apiHandler)

	// 将 API 处理程序挂载到主 mux
	mux.Handle("/api/", apiHandler)

	// 静态文件服务 - web 目录
	webDir := "web"
	mux.Handle("/", http.FileServer(http.Dir(webDir)))

	return mux
}