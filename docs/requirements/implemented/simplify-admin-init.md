# 简化管理员账户初始化机制

## 需求背景

原有的管理员账户初始化机制过于复杂：
- 需要通过 `ADMIN_USERNAME` 和 `ADMIN_PASSWORD` 环境变量配置
- 如果不设置密码，会生成随机密码并输出到控制台
- 首次启动后环境变量不再生效，容易造成困惑
- 文档说明复杂，用户理解成本高

## 需求目标

简化管理员账户初始化流程：
- 去掉 `ADMIN_USERNAME` 和 `ADMIN_PASSWORD` 环境变量
- 程序启动时自动检查数据库是否存在 auth 配置
- 如果不存在（首次启动），默认设置为 `admin/admin`
- 用户首次登录后自行修改密码

## 实现方案

### 1. 代码修改

#### service/auth.go

修改 `InitializeAuth` 函数，去掉参数，使用固定的默认值：

```go
func (s *Service) InitializeAuth() (*AuthInitResult, error) {
	config, err := s.getAuthConfig(true)
	if err == nil {
		return &AuthInitResult{
			Initialized:   false,
			Username:      config.Username,
			InitializedAt: config.InitializedAt,
		}, nil
	}
	if !errors.Is(err, ErrAuthNotInitialized) {
		return nil, err
	}

	// 首次启动，使用默认的 admin/admin
	username := "admin"
	password := "admin"

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	initializedAt := time.Now().UTC()
	sessionVersion, err := generateRandomURLSafeToken(18)
	if err != nil {
		return nil, err
	}
	tokenSecret, err := generateRandomURLSafeToken(32)
	if err != nil {
		return nil, err
	}

	if err := s.saveAuthConfig(&AuthConfig{
		Username:       username,
		PasswordHash:   string(passwordHash),
		InitializedAt:  initializedAt,
		TokenSecret:    tokenSecret,
		SessionVersion: sessionVersion,
	}); err != nil {
		return nil, err
	}

	return &AuthInitResult{
		Initialized:       true,
		Username:          username,
		Password:          password,
		GeneratedPassword: false,
		InitializedAt:     initializedAt,
	}, nil
}
```

修改 `ResetPassword` 函数，去掉 `adminUsername` 参数，使用固定的 `admin`：

```go
func (s *Service) ResetPassword(_, newPassword string) (*AuthResetResult, error) {
	// ... 省略其他代码
	config = &AuthConfig{
		Username:       "admin",
		InitializedAt:  now,
		TokenSecret:    tokenSecret,
		SessionVersion: sessionVersion,
	}
	// ... 省略其他代码
}
```

#### cmd/server/main.go

去掉环境变量读取和传参：

```go
// 创建服务实例
svc := service.NewService(store)

forceResetPassword := os.Getenv("FORCE_RESET_PASSWORD")

if *resetPassword != "" {
	result, err := svc.ResetPassword("", *resetPassword)
	if err != nil {
		log.Fatalf("Failed to reset admin password: %v", err)
	}
	log.Printf("Admin password has been reset for user %q", result.Username)
	return
}

if forceResetPassword != "" {
	result, err := svc.ResetPassword("", forceResetPassword)
	if err != nil {
		log.Fatalf("Failed to force reset admin password: %v", err)
	}
	log.Printf("FORCE_RESET_PASSWORD applied for user %q", result.Username)
	log.Printf("Remember to remove FORCE_RESET_PASSWORD from deployment configuration after use")
} else {
	result, err := svc.InitializeAuth()
	if err != nil {
		log.Fatalf("Failed to initialize admin auth: %v", err)
	}
	if result.Initialized {
		log.Println("========================================")
		log.Println("首次启动，已初始化管理员账户")
		log.Printf("用户名: %s", result.Username)
		log.Printf("密码: %s", result.Password)
		log.Println("请登录后尽快修改密码")
		log.Println("========================================")
	}
}
```

### 2. 测试修改

更新所有相关测试用例：
- `cmd/server/main_test.go` - 更新 `TestSetupRouterUsesStorageBackedBearerAuth`
- `service/auth_test.go` - 更新所有 `InitializeAuth` 和 `ResetPassword` 调用

### 3. 文档更新

更新以下文档：
- `README.md` - 环境变量说明和管理员账户说明
- `docs/guides/deployment.md` - 部署说明和生产环境注意事项
- `docs/guides/quickstart.md` - 快速开始和初始化说明
- `docs/reference/configuration.md` - 配置项说明和示例配置

## 实施结果

### 代码变更

- ✅ `service/auth.go` - 简化 `InitializeAuth` 和 `ResetPassword` 函数
- ✅ `cmd/server/main.go` - 去掉环境变量读取
- ✅ `cmd/server/main_test.go` - 更新测试用例
- ✅ `service/auth_test.go` - 更新测试用例

### 文档变更

- ✅ `README.md` - 更新环境变量表格和管理员账户说明
- ✅ `docs/guides/deployment.md` - 更新部署示例和注意事项
- ✅ `docs/guides/quickstart.md` - 更新初始化说明
- ✅ `docs/reference/configuration.md` - 更新配置项说明

### 编译验证

- ✅ 代码编译成功，无语法错误

## 使用说明

### 首次启动

首次启动时，系统会自动初始化默认管理员账户：

```bash
========================================
首次启动，已初始化管理员账户
用户名: admin
密码: admin
请登录后尽快修改密码
========================================
```

### 登录管理台

使用默认账户登录：
- 用户名: `admin`
- 密码: `admin`

登录后请立即修改密码。

### 忘记密码

使用命令行重置密码：

```bash
# 二进制部署
./singboxconfig -reset-password 'new-password-123'

# Docker 部署
docker exec -it singboxconfig /app/singboxconfig -reset-password 'new-password-123'

# 或使用环境变量（重置后需移除）
docker run -e FORCE_RESET_PASSWORD='new-password-123' \
  -e DATABASE_URL='your-database-url' \
  xsdhy/singboxconfig:latest
```

## 优势

1. **简单直观** - 不需要配置环境变量，开箱即用
2. **易于理解** - 默认账户 `admin/admin` 清晰明了
3. **减少困惑** - 不再有"首次启动生效"的复杂逻辑
4. **文档简化** - 文档说明更加简洁清晰
5. **安全提示** - 明确提示用户首次登录后修改密码

## 注意事项

1. 首次启动后请立即修改默认密码
2. 密码重置功能仍然保留，可通过命令行或环境变量重置
3. `FORCE_RESET_PASSWORD` 是持续覆盖型参数，使用后需从配置中移除

## 实施日期

2026-04-14
