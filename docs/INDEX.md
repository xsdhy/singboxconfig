# SingBox Config 文档中心

> **适用于人类和 AI 阅读的完整项目文档**
>
> 任何开发者或 AI 助手都可以通过此文档目录快速了解 SingBox Config 项目的所有细节。

## 项目简介

SingBox Config 是一个基于 Go 语言的 **sing-box 代理配置管理系统**，提供订阅源管理、节点分组、规则集配置、设备管理、WireGuard 端点管理等功能，能够为每台设备动态生成定制化的 sing-box JSON、Surge 文本配置与 Shadowrocket 文本配置文件。

项目采用前后端分离架构，前端构建产物嵌入 Go 二进制中一体化部署：
- **后端服务**: Go + Gin + GORM，支持多种存储后端（PostgreSQL / MySQL / Supabase / JSON 文件）
- **前端应用**: React 18 + TypeScript + Vite + Arco Design，提供完整的管理界面
- **协议解析**: 内置 SS、SSR、Trojan、VMess、VLESS 订阅协议解码器

## 文档目录

### 快速入门 (guides/)

| 文档 | 说明 |
|------|------|
| [快速开始](./guides/quickstart.md) | 环境要求、依赖安装、本地启动 |
| [部署说明](./guides/deployment.md) | Docker 部署、二进制部署、环境变量配置 |
| [开发指南](./guides/development.md) | 开发规范、前后端联调、新功能开发流程 |

### 架构设计 (architecture/)

| 文档 | 说明 |
|------|------|
| [项目概述](./architecture/overview.md) | 项目背景、整体架构、核心工作流程 |
| [技术栈](./architecture/tech-stack.md) | 技术选型与依赖说明 |
| [目录结构](./architecture/directory-structure.md) | 项目目录与代码组织详解 |
| [存储抽象层](./architecture/storage-layer.md) | 多存储后端设计（Memory / Database / Supabase） |
| [配置生成流程](./architecture/config-generation.md) | 设备配置动态生成的完整流程与数据流 |

### 核心模块 (modules/)

| 文档 | 说明 |
|------|------|
| [订阅管理](./modules/subscribe.md) | 订阅源 CRUD、多协议解析（SS/SSR/Trojan/VMess/VLESS） |
| [节点分组](./modules/node-group.md) | 节点分组规则、selector/urltest 策略 |
| [规则集管理](./modules/rule-set.md) | 路由规则集定义、本地/远程规则源 |
| [设备管理](./modules/device.md) | 设备注册、Token 认证、配置绑定 |
| [入站管理](./modules/inbound.md) | TUN/HTTP/SOCKS/Mixed 入站模板配置 |
| [WireGuard 管理](./modules/wireguard.md) | WireGuard 端点与 Peer 配置 |
| [Outbound 管理](./modules/outbound.md) | 统一管理手工节点与订阅缓存节点 |
| [DNS 配置](./modules/dns.md) | DNS 服务器与路由规则配置 |
| [配置导入导出](./modules/config-transfer.md) | 配置数据的导出、导入与默认值初始化 |
| [协议解码器](./modules/protocol-decoders.md) | SS/SSR/Trojan/VMess/VLESS URL 解析实现 |

### 前端文档 (frontend/)

| 文档 | 说明 |
|------|------|
| [前端架构](./frontend/architecture.md) | React 应用结构、组件设计、状态管理 |
| [页面说明](./frontend/pages.md) | 各管理页面功能与交互说明 |
| [API 客户端](./frontend/api-client.md) | Axios 封装、接口调用规范 |

### 参考资料 (reference/)

| 文档 | 说明 |
|------|------|
| [API 接口列表](./reference/api-reference.md) | 完整 REST API 接口文档 |
| [配置项说明](./reference/configuration.md) | 环境变量与存储后端配置说明 |
| [sing-box 配置结构](./reference/singbox-config-schema.md) | 生成的 sing-box JSON 配置结构说明 |
| [常见问题](./reference/faq.md) | 常见问题与解决方案 |

### 需求文档 (requirements/)

需求文档按实施状态分目录存放，每个需求单独一个文件。

| 目录 | 说明 |
|------|------|
| [requirements/implemented/](./requirements/implemented/) | 已实现的需求文档 |
| [requirements/planned/](./requirements/planned/) | 计划中的需求文档 |
| [requirements/bugfixes/](./requirements/bugfixes/) | Bug 排查与修复文档 |

## 目录结构

```
docs/
├── INDEX.md                          # 本文件 - 文档总入口
├── guides/                           # 快速入门与指南
│   ├── quickstart.md                 # 快速开始
│   ├── deployment.md                 # 部署说明
│   └── development.md                # 开发指南
├── architecture/                     # 架构设计
│   ├── overview.md                   # 项目概述
│   ├── tech-stack.md                 # 技术栈
│   ├── directory-structure.md        # 目录结构
│   ├── storage-layer.md              # 存储抽象层
│   └── config-generation.md          # 配置生成流程
├── modules/                          # 核心模块文档
│   ├── subscribe.md                  # 订阅管理
│   ├── node-group.md                 # 节点分组
│   ├── rule-set.md                   # 规则集管理
│   ├── device.md                     # 设备管理
│   ├── inbound.md                    # 入站管理
│   ├── wireguard.md                  # WireGuard 管理
│   ├── outbound.md                   # Outbound 统一管理
│   ├── dns.md                        # DNS 配置
│   ├── config-transfer.md            # 配置导入导出
│   └── protocol-decoders.md          # 协议解码器
├── frontend/                         # 前端文档
│   ├── architecture.md               # 前端架构
│   ├── pages.md                      # 页面说明
│   └── api-client.md                 # API 客户端
├── reference/                        # 参考资料
│   ├── api-reference.md              # API 接口列表
│   ├── configuration.md              # 配置项说明
│   ├── singbox-config-schema.md      # sing-box 配置结构
│   └── faq.md                        # 常见问题
└── requirements/                     # 需求文档
    ├── implemented/                  # 已实现需求
    ├── planned/                      # 计划中需求
    └── bugfixes/                     # Bug 排查与修复
```

## 技术栈概览

### 后端
- **语言**: Go 1.25.0
- **Web 框架**: Gin v1.10.0
- **ORM**: GORM v1.31.1
- **数据库**: PostgreSQL / MySQL（多后端可选）
- **云存储**: Supabase REST API（可选后端）
- **文件存储**: JSON 文件（内存模式，适合轻量部署）
- **日志**: Logrus v1.9.3
- **协议解析**: 自研 SS/SSR/Trojan/VMess/VLESS 解码器

### 前端
- **框架**: React 18.3.1
- **语言**: TypeScript 5.5.0
- **构建工具**: Vite 5.4.0
- **UI 组件库**: Arco Design Web React v2.64.0
- **代码编辑器**: Monaco Editor React v4.6.0
- **HTTP 客户端**: Axios v1.7.0

### 部署
- **容器化**: Docker（基于 golang:1.25.0-alpine）
- **构建**: Makefile 自动化（前端构建 → Go 交叉编译 → 嵌入式 SPA）
- **默认端口**: 7391

## 核心功能

- **订阅管理**: 添加多个代理订阅源，自动解析 SS/SSR/Trojan/VMess/VLESS 节点
- **节点分组**: 通过关键字过滤动态组织节点，支持 selector 手动选择和 urltest 自动测速
- **规则集管理**: 管理路由规则集，支持本地和远程规则源，控制流量分流策略
- **设备管理**: 为不同设备定制配置，通过 Token 认证获取专属 sing-box 配置
- **入站配置**: 管理 TUN/HTTP/SOCKS/Mixed 入站模板，按设备灵活绑定
- **WireGuard 管理**: 配置 WireGuard 端点与 Peer，集成到 sing-box 配置中
- **Outbound 管理**: 统一管理手工节点与订阅缓存节点，支持缓存与筛选
- **DNS 配置**: 配置 DNS 服务器与路由规则
- **配置生成**: 一键生成设备专属的完整 sing-box JSON 配置
- **Surge 输出**: 复用同一数据层生成设备专属 Surge 配置文本，手工节点输出为 [Proxy] 行（覆盖 SS / Trojan / VMess / HTTP(HTTPS) / SOCKS5 协议），订阅源输出为携带 policy-path 的策略组由 Surge 自行拉取，并导出 WireGuard endpoint
- **Shadowrocket 输出**: 复用同一数据层生成设备专属 Shadowrocket 配置文本，覆盖 SS / SSR / Trojan / VMess / VLESS 等协议
- **导入导出**: 配置数据的备份、迁移与默认值初始化

## 文档维护说明

### 文档编写规范

1. **新功能开发**: 在 `requirements/planned/` 创建需求文档，实现后移至 `requirements/implemented/`
2. **Bug 修复**: 在 `requirements/bugfixes/` 记录排查与修复过程
3. **模块变更**: 更新 `modules/` 下对应文档
4. **API 变更**: 同步更新 `reference/api-reference.md`
5. **架构调整**: 更新 `architecture/` 下相关文档
6. **前端变更**: 更新 `frontend/` 下相关文档

### AI 助手使用建议

- **了解项目**: `INDEX.md` → `architecture/overview.md` → `architecture/tech-stack.md`
- **快速上手**: 查阅 `guides/quickstart.md`
- **开发新功能**: 查阅 `guides/development.md` 和相关 `modules/` 文档
- **理解配置生成**: 查阅 `architecture/config-generation.md`
- **理解存储设计**: 查阅 `architecture/storage-layer.md`
- **部署运维**: 查阅 `guides/deployment.md` 和 `reference/configuration.md`
- **API 集成**: 查阅 `reference/api-reference.md`
- **前端开发**: 查阅 `frontend/architecture.md` 和 `frontend/pages.md`
- **问题排查**: 查阅 `reference/faq.md`

## 快速链接

- [快速开始](./guides/quickstart.md) - 环境配置与本地启动
- [项目概述](./architecture/overview.md) - 整体架构与设计理念
- [API 文档](./reference/api-reference.md) - 完整 REST API 接口参考
- [配置生成流程](./architecture/config-generation.md) - 核心配置生成逻辑
- [部署说明](./guides/deployment.md) - 生产环境部署指南
- [常见问题](./reference/faq.md) - 常见问题解答
