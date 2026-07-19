# go-admin

Golang 后台管理基础框架，内置用户、角色、菜单、权限全套基础模块，前后端分离设计，开箱即用，快速实现各类管理系统开发。

## 特性

- 基于 [infra-go](https://github.com/chihqiang/infra-go) 构建，使用 `httpx`、`orm`、`jwt`、`logger` 等开箱即用的能力
- 账号 - 角色 - 菜单 RBAC 权限模型，支持 API 级别细粒度权限控制
- JWT 双 Token 鉴权（access_token + refresh_token）
- 请求日志自动记录（路径、方法、耗时、IP、UA、响应码），敏感字段脱敏
- 数据库自动迁移 + 种子数据初始化，首次启动即拥有完整演示数据
- 支持 SQLite / MySQL / PostgreSQL 多数据库切换
- 前后端分离，前端项目 [next-admin](https://github.com/chihqiang/next-admin)

## 技术栈

| 层级 | 技术 |
|------|------|
| 语言 | Go 1.25+ |
| HTTP 框架 | [infra-go/httpx](https://github.com/chihqiang/infra-go) |
| ORM | GORM |
| 鉴权 | JWT (HS256) |
| 数据库 | SQLite（默认）/ MySQL / PostgreSQL |
| 密码加密 | bcrypt |

## 项目结构

```bash
go-admin/
├── config/         # 配置结构定义
├── db/             # 数据库迁移 & 种子数据
├── handler/        # HTTP 请求处理层（参数校验、调用 logic、返回响应）
├── logic/          # 业务逻辑层（数据库操作、事务管理）
├── middleware/      # 中间件（鉴权、权限、日志、上下文注入）
├── model/          # 数据模型定义（GORM）
├── route/          # 路由注册 & 中间件编排
├── config.yaml     # 配置文件
└── main.go         # 入口
```

### 调用链路

```bash
Route → Middleware → Handler → Logic → Model → DB
```

- **Middleware**：不直接操作数据库，只负责鉴权、上下文注入等横切关注点
- **Handler**：参数解析与校验，调用 Logic，构造响应
- **Logic**：封装所有数据库操作和业务规则，供 Middleware 和 Handler 调用

## 快速开始

### 环境要求

- Go 1.25+
- CGO（SQLite 需要）

### 启动

```bash
# 克隆项目
git clone https://github.com/chihqiang/go-admin.git
cd go-admin

# 启动（首次启动自动迁移数据库并初始化种子数据）
go run main.go
```

服务默认监听 `http://0.0.0.0:8080`。

### 默认账号

| 邮箱 | 密码 |
|------|------|
| admin@example.com | 123456 |

> 仅首次启动时初始化，已存在账号数据则跳过。

## 配置

编辑 `config.yaml`：

```yaml
app:
  name: go-admin
  version: 0.0.1

server:
  host: 0.0.0.0
  port: 8080

db:
  driver: sqlite                    # sqlite / mysql / postgres
  database: ./data.db              # SQLite 文件路径

# MySQL 示例
# db:
#   driver: mysql
#   host: 127.0.0.1
#   port: 3306
#   username: root
#   password: root
#   database: go_admin
#   charset: utf8mb4

# PostgreSQL 示例
# db:
#   driver: postgres
#   host: 127.0.0.1
#   port: 5432
#   username: postgres
#   password: postgres
#   database: go_admin

jwt:
  secret: go-admin-secret-key
  issuer: go-admin
  access_token_expire: 2h
  refresh_token_expire: 168h
  algorithm: HS256

logger:
  level: 0
  encoding: json
  output:
    - stdout
  caller: true
```

## API 接口

基础路径：`/api/v1`

### 认证（无需鉴权）

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/auth/login` | 登录获取 Token |

### 鉴权接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/auth/me` | 获取当前登录用户信息 |

### 账号管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/sys/accounts` | 账号列表 |
| GET | `/sys/accounts/{id}` | 账号详情 |
| POST | `/sys/accounts` | 创建账号 |
| PUT | `/sys/accounts/{id}` | 更新账号 |
| DELETE | `/sys/accounts/{id}` | 删除账号（禁止删除自己） |

### 角色管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/sys/roles` | 角色列表（分页） |
| GET | `/sys/roles/all` | 角色列表（全量，下拉选择用） |
| GET | `/sys/roles/{id}` | 角色详情（含关联菜单） |
| POST | `/sys/roles` | 创建角色 |
| PUT | `/sys/roles/{id}` | 更新角色 |
| DELETE | `/sys/roles/{id}` | 删除角色 |
| POST | `/sys/roles/{id}/menus` | 关联角色与菜单 |

### 菜单管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/sys/menus` | 菜单列表（树形） |
| GET | `/sys/menus/all` | 菜单列表（全量） |
| GET | `/sys/menus/{id}` | 菜单详情 |
| POST | `/sys/menus` | 创建菜单 |
| PUT | `/sys/menus/{id}` | 更新菜单 |
| DELETE | `/sys/menus/{id}` | 删除菜单（含子菜单及角色关联） |

### 日志管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/sys/logs` | 操作日志列表 |

## 权限模型

采用 RBAC 模型：**账号 → 角色 → 菜单**

- 每个菜单（`sys_menus`）的 `api_url` + `api_method` 字段定义了该菜单可访问的 API 权限
- 角色通过关联表 `sys_role_menus` 绑定多个菜单
- 账号通过关联表 `sys_account_roles` 绑定多个角色
- 权限中间件逐条匹配当前请求的 `Method + URI` 与账号所拥有的菜单权限，支持前缀匹配（用于 `{id}` 参数路由）
- 种子数据使用 `*` 通配符处理参数路由，如 `/api/v1/sys/accounts/*`

## 中间件

| 中间件 | 说明 |
|--------|------|
| `WithCors` | 跨域处理 |
| `WithRecovery` | panic 恢复 |
| `WithLogger` | 请求日志（httpx 内置） |
| `Log` | 业务操作日志记录，支持按路由和 HTTP 方法过滤 |
| `Auth` | JWT Token 校验 |
| `LoadAccount` | 从 Token 中解析账号 ID，加载账号及角色菜单信息注入上下文 |
| `Permission` | 基于上下文中已加载的账号信息进行 API 权限校验，零数据库查询 |

## 前端

配套前端项目：[next-admin](https://github.com/chihqiang/next-admin)

## License

[Apache License 2.0](LICENSE) - Copyright 2026 chihqiang
