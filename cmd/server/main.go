package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"singboxconfig/service"
	"singboxconfig/storage"

	_ "embed"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed index.html
var indexHTML []byte

func main() {
	resetPassword := flag.String("reset-password", "", "reset admin password and exit")
	flag.Parse()

	// 创建存储实例
	var store storage.Storage
	var err error

	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_KEY")
	databaseURL := os.Getenv("DATABASE_URL")

	switch {
	case supabaseURL != "" && supabaseKey != "":
		store = storage.NewSupabaseStorage(supabaseURL, supabaseKey)
		log.Printf("Using Supabase REST storage: %s", supabaseURL)

	case databaseURL != "":
		db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to DB: %v", err)
		}
		store, err = storage.NewDatabaseStorage(db)
		if err != nil {
			log.Fatalf("Failed to initialize database storage: %v", err)
		}
		log.Printf("Using DB storage: %s", databaseURL)

	default:
		log.Fatal("No storage backend configured. Please set DATABASE_URL or SUPABASE_URL/SUPABASE_KEY environment variables.")
	}

	_ = err

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

	// 设置路由
	r := SetupRouter(svc)

	prot := os.Getenv("PORT")
	if prot == "" {
		prot = "7391"
	}

	// 启动服务器
	log.Printf("Starting server on port %s", prot)
	if err := r.Run(":" + prot); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// SetupRouter 设置路由
func SetupRouter(service *service.Service) *gin.Engine {
	r := gin.Default()

	r.GET("/open/generate/:device", service.Generated)
	r.GET("/open/surge/:device", service.SurgeGenerated)
	r.GET("/open/shadowrocket/:device", service.ShadowrocketGenerated)
	r.GET("/open/ruleset/:tag", service.GetRuleSetByTag)
	r.GET("/open/rules/:tag", service.GetRulesBySoftware)
	r.GET("/", func(c *gin.Context) { c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML) })

	r.NoRoute(func(c *gin.Context) {
		c.AbortWithStatus(http.StatusNotFound)
	})

	api := r.Group("/api")

	authPublicGroup := api.Group("/auth")
	{
		authPublicGroup.POST("/login", service.Login)
	}

	protected := api.Group("")
	protected.Use(service.AuthMiddleware())

	authGroup := protected.Group("/auth")
	{
		authGroup.GET("/me", service.GetAuthMe)
		authGroup.POST("/change-credentials", service.ChangeCredentialsHandler)
	}

	// 订阅管理
	subscribeGroup := protected.Group("/subscribes")
	{
		subscribeGroup.POST("", service.CreateSubscribe)
		subscribeGroup.PUT("/:name", service.UpdateSubscribe)
		subscribeGroup.DELETE("/:name", service.DeleteSubscribe)
		subscribeGroup.GET("", service.ListSubscribes)
		subscribeGroup.PUT("/:name/cache-config", service.UpdateSubscribeCacheConfig)
		subscribeGroup.POST("/:name/refresh-outbound", service.RefreshSubscribeOutbounds)
		subscribeGroup.GET("/:name/outbounds", service.ListSubscribeOutbounds)
	}

	// 节点分组管理
	nodeGroupGroup := protected.Group("/node-groups")
	{
		nodeGroupGroup.POST("", service.CreateNodeGroup)
		nodeGroupGroup.PUT("/:tag", service.UpdateNodeGroup)
		nodeGroupGroup.DELETE("/:tag", service.DeleteNodeGroup)
		nodeGroupGroup.GET("", service.ListNodeGroups)
	}

	// 规则集管理
	ruleSetGroup := protected.Group("/rule-sets")
	{
		ruleSetGroup.POST("", service.CreateRuleSet)
		ruleSetGroup.PUT("/:tag", service.UpdateRuleSet)
		ruleSetGroup.DELETE("/:tag", service.DeleteRuleSet)
		ruleSetGroup.GET("", service.ListRuleSets)
	}

	// 全局设置
	settingGroup := protected.Group("/settings")
	{
		settingGroup.POST("", service.CreateSetting)
		settingGroup.PUT("/:key", service.UpdateSetting)
		settingGroup.DELETE("/:key", service.DeleteSetting)
		settingGroup.GET("/:key", service.GetSetting)
		settingGroup.GET("/key/:key", service.GetSettingByKey)
		settingGroup.GET("", service.ListSettings)
	}

	configTransferGroup := protected.Group("/config-transfer")
	{
		configTransferGroup.GET("/export", service.ExportConfig)
		configTransferGroup.POST("/import", service.ImportConfig)
	}

	// 设备管理
	deviceGroup := protected.Group("/devices")
	{
		deviceGroup.POST("", service.CreateDevice)
		deviceGroup.GET("", service.ListDevices)
		deviceGroup.GET("/:code", service.GetDevice)
		deviceGroup.PUT("/:code", service.UpdateDevice)
		deviceGroup.DELETE("/:code", service.DeleteDevice)
		deviceGroup.PUT("/:code/inbounds", service.SetDeviceInbounds)
		deviceGroup.GET("/:code/inbounds", service.ListDeviceInbounds)
	}

	// Inbound 模板管理
	inboundGroup := protected.Group("/inbounds")
	{
		inboundGroup.POST("", service.CreateInbound)
		inboundGroup.GET("", service.ListInbounds)
		inboundGroup.GET("/:tag", service.GetInbound)
		inboundGroup.PUT("/:tag", service.UpdateInbound)
		inboundGroup.DELETE("/:tag", service.DeleteInbound)
	}

	// WireGuard 管理
	wireGuardGroup := protected.Group("/wire-guards")
	{
		wireGuardGroup.POST("", service.CreateWireGuard)
		wireGuardGroup.GET("", service.ListWireGuards)
		wireGuardGroup.GET("/:tag", service.GetWireGuard)
		wireGuardGroup.PUT("/:tag", service.UpdateWireGuard)
		wireGuardGroup.DELETE("/:tag", service.DeleteWireGuard)
		wireGuardGroup.POST("/:tag/peers", service.CreateWireGuardPeer)
		wireGuardGroup.GET("/:tag/peers", service.ListWireGuardPeers)
		wireGuardGroup.PUT("/:tag/peers/:id", service.UpdateWireGuardPeer)
		wireGuardGroup.DELETE("/:tag/peers/:id", service.DeleteWireGuardPeer)
	}

	// 统一 Outbound 管理
	outboundGroup := protected.Group("/outbounds")
	{
		outboundGroup.POST("", service.CreateOutbound)
		outboundGroup.GET("", service.ListOutbounds)
		outboundGroup.PUT("/:id", service.UpdateOutbound)
		outboundGroup.DELETE("/:id", service.DeleteOutbound)
		outboundGroup.PATCH("/batch-enable", service.BatchEnableOutbounds)
	}

	return r
}
