# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 在处理此仓库代码时提供指导。

## 常用开发命令

```bash
# 构建应用程序
go build -o bin/lab ./lab/cmd/lab

# 运行应用程序
go run ./lab/cmd/lab

# 初始化数据库结构
mysql -u root -p < lab/database/schema.sql
# 或使用默认凭证：
mysql -u root -pcran1234 < lab/database/schema.sql

# 设置环境变量（示例）
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASS=cran1234
export DB_NAME=library_db
export SERVER_PORT=8080
```

## 代码架构概览

应用程序采用清晰的分层架构模式，各组件职责明确：

### 主要组件
- **入口点**: `lab/cmd/lab/main.go` - 初始化依赖项并启动 HTTP 服务器
- **配置**: `lab/config/config.go` - 加载环境变量并提供数据库连接字符串 (DSN)
- **路由**: `lab/internal/router/router.go` - 定义 API 端点并应用中间件
- **中间件**: `lab/internal/middleware/middleware.go` - 提供日志记录、异常恢复和 JSON 内容类型验证
- **处理器**: `lab/internal/handler/` - 处理 HTTP 请求、验证输入、调用服务、格式化响应
- **服务层**: `lab/internal/service/` - 包含业务逻辑、验证规则，并协调仓库操作
- **仓库层**: `lab/internal/repository/` - 处理特定实体的数据库操作 (CRUD)
- **模型**: `lab/internal/model/` - 定义映射到数据库表的数据结构
- **错误处理**: `lab/internal/errors/` - 自定义错误类型，确保一致的错误处理
- **数据库结构**: `lab/database/schema.sql` - 用户和图书表的 MySQL 结构定义

### API 端点
应用程序在 `/api/` 路径下暴露 RESTful APIs：
- **用户端点**: `/api/users` (CRUD 操作)
- **图书端点**: `/api/books` (CRUD 操作)
- **健康检查**: `/api/health`

### 数据库结构
- **users 表**: 存储用户信息，包含唯一邮箱约束
- **books 表**: 存储图书信息，包含唯一 ISBN 约束以及状态/分类索引

### 静态文件
静态文件从 `web/` 目录在根路径 (`/`) 下提供服务。

### 依赖管理
项目使用 Go 模块 (`go.mod`)，依赖项精简：
- `github.com/go-sql-driver/mysql` 用于 MySQL 数据库连接