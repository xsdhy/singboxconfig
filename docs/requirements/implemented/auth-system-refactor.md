# 认证系统重构

> 状态：已实现

## 背景

原始实现把管理端认证绑定在浏览器 HTTP Basic Auth 上，存在几个直接问题：

1. 用户体验差，浏览器原生弹窗不可控。
2. 前端没有独立登录态，无法实现退出登录、过期处理和账号设置。
3. 用户名虽然持久化了，但不能通过后台自行修改。
4. 认证配置和普通全局设置之间需要更明确的边界。

本次重构目标不是做完整用户系统，而是在保留“单管理员模型”的前提下，把认证升级为可运维、可扩展、前端可控的登录体系。

## 目标

- 支持首次启动初始化管理员账户
- 支持修改用户名
- 支持修改密码
- 不再依赖浏览器 HTTP Basic Auth
- 改为显式登录接口 + Bearer Token 鉴权
- 继续复用现有存储抽象，不新增认证表
- 保持单管理员模型

## 非目标

- 多用户系统
- RBAC
- OAuth2 / OIDC
- 刷新 token / 多端会话管理
- 登录审计、限流、2FA

## 方案结论

### 存储层

认证配置继续落在 `GlobalSetting` 中，但视为系统保留配置：

- `auth.username`
- `auth.password_hash`
- `auth.initialized_at`
- `auth.password_changed_at`
- `auth.token_secret`
- `auth.session_version`

约束不变：

- `auth.*` 是保留 key
- `/api/settings` 不允许读写删 `auth.*`
- 配置导入导出默认排除 `auth.*`

### 认证模型

管理端认证从 Basic Auth 切换为：

1. `POST /api/auth/login` 提交用户名密码
2. 后端校验后签发 HS256 Bearer Token
3. 前端把 token 存到本地并在后续 `/api/*` 请求中带 `Authorization: Bearer ...`
4. `/api/auth/me` 返回当前管理员资料
5. `/api/auth/change-credentials` 支持修改用户名和密码

`/api/admin` 本身不再受认证保护，页面加载后由前端自己完成登录态判断和跳转。

## 设计细节

### 1. 初始化与重置

首次启动判定仍然只在以下条件下触发：

- `auth.username` 不存在
- `auth.password_hash` 不存在

以下情况不能误判为首次启动：

- 存储异常
- 超时
- 只存在一半配置

初始化来源优先级：

1. `ADMIN_USERNAME` / `ADMIN_PASSWORD`
2. 默认用户名 `admin` + 随机生成密码

重置方式保持两种：

- `-reset-password`
- `FORCE_RESET_PASSWORD`

`FORCE_RESET_PASSWORD` 仍然是持续覆盖型参数，部署完成后要手动移除。

### 2. Token 策略

采用后端自行签发的 HS256 JWT 风格 Bearer Token，claims 至少包含：

- `sub`：当前用户名
- `exp`：过期时间
- `iat`：签发时间
- `sv`：当前会话版本

配套字段：

- `auth.token_secret`：签名密钥
- `auth.session_version`：会话版本

作用：

- 重启后 token 仍可校验
- 修改用户名或密码时，轮换 `session_version`
- 旧 token 在账号变更后立即失效

### 3. 账号修改

新增专用接口：

- `POST /api/auth/change-credentials`

请求体：

```json
{
  "old_password": "current-password",
  "new_username": "ops-admin",
  "new_password": "new-password-123"
}
```

规则：

- `old_password` 必填
- `new_username` 和 `new_password` 至少提供一项
- `new_password` 仍需满足最小长度要求
- 修改用户名或密码后都会轮换 `session_version`
- 成功响应返回新的 Bearer Token，前端立即替换本地 token

### 4. 前端交互

前端改为显式登录态管理：

- 打开 `/api/admin` 时，如果本地没有可用 token，先显示登录页
- 登录成功后进入后台
- 账号设置弹窗支持修改用户名和密码
- 提供退出登录按钮，清理本地 token

这避免了浏览器原生认证弹窗，也为后续扩展更完整的会话管理留出了空间。

## 实现范围

### 后端

- [x] 新增 `POST /api/auth/login`
- [x] 新增 Bearer Token 中间件
- [x] 新增 `POST /api/auth/change-credentials`
- [x] 保留 `GET /api/auth/me`
- [x] 保留首次初始化、`-reset-password`、`FORCE_RESET_PASSWORD`
- [x] 继续隔离 `auth.*` 与普通设置/导入导出

### 前端

- [x] 管理台登录页
- [x] Bearer Token 本地存储与请求注入
- [x] 账号设置弹窗支持修改用户名和密码
- [x] 退出登录

### 文档

- [x] 更新快速开始
- [x] 更新部署说明
- [x] 更新配置项说明
- [x] 更新 API 文档
- [x] 更新 FAQ

## 验收结果

- [x] 不再依赖浏览器 HTTP Basic Auth
- [x] 支持用户名修改
- [x] 支持密码修改
- [x] 支持管理员登录、退出和重新登录
- [x] 修改账号信息后旧 token 立即失效
- [x] `auth.*` 仍不通过普通设置和导入导出暴露
