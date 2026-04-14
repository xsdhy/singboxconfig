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
- 💾 **多存储后端** - PostgreSQL / MySQL / Supabase / JSON 文件

## 🚀 快速开始

```bash
# 克隆项目
git clone https://github.com/xsdhy/singboxconfig.git
cd singboxconfig

# 启动服务（默认端口 7391）
make run

# 或使用 Docker
docker build -t singboxconfig .
docker run -p 7391:7391 singboxconfig
```

访问 `http://localhost:7391` 开始使用。

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