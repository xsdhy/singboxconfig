# SingBox Config

> 基于 Go 的 sing-box 代理配置管理系统

SingBox Config 提供订阅源管理、节点分组、规则集配置、设备管理等功能，能够为每台设备动态生成定制化的 sing-box JSON 配置文件。

## ✨ 核心特性

- 🔄 **订阅管理** - 支持 SS/SSR/Trojan/VMess 协议自动解析
- 📦 **节点分组** - 关键字过滤 + selector/urltest 策略
- 🎯 **规则集管理** - 灵活的流量分流规则配置
- 📱 **设备管理** - 多设备独立配置，Token 认证
- 🔌 **入站配置** - TUN/HTTP/SOCKS/Mixed 模板管理
- 🔐 **WireGuard** - 端点与 Peer 配置集成
- 🌐 **DNS 配置** - DNS 服务器与路由规则
- 💾 **多存储后端** - PostgreSQL / MySQL / Supabase

## 🚀 快速开始

### 使用 Docker Compose

创建 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  singboxconfig:
    image: xsdhy/singboxconfig:latest
    container_name: singboxconfig
    ports:
      - "7391:7391"
    environment:
      # 基础配置
      # PORT: 7391

      # 存储配置（2选一）

      # 1. PostgreSQL（需要外部数据库）
      # DATABASE_URL: postgres://user:pass@your-db-host:5432/singboxconfig?sslmode=disable

      # 2. Supabase
      # SUPABASE_URL: https://xxx.supabase.co
      # SUPABASE_KEY: your-service-role-key
    restart: unless-stopped
```

启动服务：

```bash
docker-compose up -d
```


## ⚙️ 环境变量说明

| 变量名 | 说明 | 默认值 | 必填 |
|--------|------|--------|------|
| `PORT` | HTTP 服务端口 | `7391` | 否 |
| `DATABASE_URL` | PostgreSQL/MySQL 连接串 | - | 是* |
| `SUPABASE_URL` | Supabase 项目地址 | - | 是* |
| `SUPABASE_KEY` | Supabase Service Key | - | 是* |

*必须配置 `DATABASE_URL` 或 `SUPABASE_URL` + `SUPABASE_KEY` 其中之一

### 管理员账户说明

**首次启动时：**

系统会自动检查数据库中是否存在管理员配置，如果不存在（首次启动），会自动初始化默认管理员账户：

- 用户名: `admin`
- 密码: `admin`

**控制台日志示例：**

```text
========================================
首次启动，已初始化管理员账户
用户名: admin
密码: admin
请登录后尽快修改密码
========================================
```

**重要提示：**

- 首次启动后，请立即登录管理台修改默认密码
- 后续启动会使用已保存的管理员配置，不会重置为默认值

**忘记密码？**

使用以下命令重置密码：

```bash
# Docker 环境
docker exec -it singboxconfig /app/singboxconfig -reset-password 'new-password-123'

# 或使用环境变量（重置后需移除此变量）
docker run -e FORCE_RESET_PASSWORD='new-password-123' \
  -e DATABASE_URL='your-database-url' \
  xsdhy/singboxconfig:latest
```

## 💾 存储方式

系统按以下优先级选择存储后端：

1. **Supabase** - 同时配置 `SUPABASE_URL` 和 `SUPABASE_KEY`
2. **PostgreSQL/MySQL** - 配置 `DATABASE_URL`


### PostgreSQL

适合生产环境：

```bash
DATABASE_URL=postgres://user:password@host:5432/dbname?sslmode=disable
```

### MySQL

适合已有 MySQL 环境：

```bash
DATABASE_URL=mysql://user:password@host:3306/dbname?charset=utf8mb4&parseTime=True
```

### Supabase

适合云端部署：

```bash
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_KEY=your-service-role-key
```

## 📚 完整文档

**所有详细文档请查看：[docs/INDEX.md](docs/INDEX.md)**

- [快速开始](docs/guides/quickstart.md) - 环境配置与本地启动
- [部署说明](docs/guides/deployment.md) - Docker 部署与环境变量配置
- [开发指南](docs/guides/development.md) - 开发规范与新功能开发流程
- [项目概述](docs/architecture/overview.md) - 整体架构与设计理念
- [API 文档](docs/reference/api-reference.md) - 完整 REST API 接口参考
- [常见问题](docs/reference/faq.md) - 常见问题解答

## 🛠️ 技术栈

- **后端**: Go 1.25 + Gin + GORM
- **前端**: React 18 + TypeScript + Vite + Arco Design
- **存储**: PostgreSQL / MySQL / Supabase / JSON 文件
- **部署**: Docker / 单二进制文件（前端嵌入）

## ⚠️ 免责声明

本项目仅供学习研究使用，请遵守当地法律法规。使用者需对自己的行为负责，开发者不对任何滥用行为承担责任。

## 📄 许可证

本项目采用 [MIT License](LICENSE) 开源。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request。详细的贡献指南请查看 [开发文档](docs/guides/development.md)。