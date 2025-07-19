// @title 小智服务端 API 文档
// @version 1.0
// @description 小智服务端，包含OTA与Vision等接口
// @host localhost:8080
// @BasePath /api
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"syscall"
	"time"

	"xiaozhi-server-go/src/configs"
	"xiaozhi-server-go/src/configs/database"
	cfg "xiaozhi-server-go/src/configs/server"
	"xiaozhi-server-go/src/core"
	"xiaozhi-server-go/src/core/utils"
	_ "xiaozhi-server-go/src/docs"
	"xiaozhi-server-go/src/ota"
	"xiaozhi-server-go/src/vision"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// 导入所有providers以确保init函数被调用
	_ "xiaozhi-server-go/src/core/providers/asr/doubao"
	_ "xiaozhi-server-go/src/core/providers/asr/gosherpa"
	_ "xiaozhi-server-go/src/core/providers/llm/coze"
	_ "xiaozhi-server-go/src/core/providers/llm/ollama"
	_ "xiaozhi-server-go/src/core/providers/llm/openai"
	_ "xiaozhi-server-go/src/core/providers/tts/doubao"
	_ "xiaozhi-server-go/src/core/providers/tts/edge"
	_ "xiaozhi-server-go/src/core/providers/tts/gosherpa"
	_ "xiaozhi-server-go/src/core/providers/vlllm/ollama"
	_ "xiaozhi-server-go/src/core/providers/vlllm/openai"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
)

func LoadConfigAndLogger() (*configs.Config, *utils.Logger, error) {
	// 加载配置,默认使用.config.yaml
	config, configPath, err := configs.LoadConfig()
	if err != nil {
		return nil, nil, err
	}

	// 初始化日志系统
	logger, err := utils.NewLogger(config)
	if err != nil {
		return nil, nil, err
	}
	logger.Info(fmt.Sprintf("日志系统初始化成功, 配置文件路径: %s", configPath))

	return config, logger, nil
}

func StartWSServer(config *configs.Config, logger *utils.Logger, g *errgroup.Group, groupCtx context.Context) (*core.WebSocketServer, error) {
	// 创建 WebSocket 服务
	wsServer, err := core.NewWebSocketServer(config, logger)
	if err != nil {
		return nil, err
	}

	// 启动 WebSocket 服务
	g.Go(func() error {
		// 监听关闭信号
		go func() {
			<-groupCtx.Done()
			logger.Info("收到关闭信号，开始关闭WebSocket服务...")
			if err := wsServer.Stop(); err != nil {
				logger.Error("WebSocket服务关闭失败", err)
			} else {
				logger.Info("WebSocket服务已优雅关闭")
			}
		}()

		if err := wsServer.Start(groupCtx); err != nil {
			if groupCtx.Err() != nil {
				return nil // 正常关闭
			}
			logger.Error("WebSocket 服务运行失败", err)
			return err
		}
		return nil
	})

	logger.Info("WebSocket 服务已成功启动")
	return wsServer, nil
}

func StartHttpServer(config *configs.Config, logger *utils.Logger, g *errgroup.Group, groupCtx context.Context) (*http.Server, error) {
	// 初始化Gin引擎
	if config.Log.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()
	router.SetTrustedProxies([]string{"0.0.0.0"})

	// API路由全部挂载到/api前缀下
	apiGroup := router.Group("/api")
	// 启动OTA服务
	otaService := ota.NewDefaultOTAService(config.Web.Websocket)
	if err := otaService.Start(groupCtx, router, apiGroup); err != nil {
		logger.Error("OTA 服务启动失败", err)
		return nil, err
	}

	// 启动Vision服务
	visionService, err := vision.NewDefaultVisionService(config, logger)
	if err != nil {
		logger.Error("Vision 服务初始化失败 %v", err)
		return nil, err
	}
	if err := visionService.Start(groupCtx, router, apiGroup); err != nil {
		logger.Error("Vision 服务启动失败", err)
		return nil, err
	}

	cfgServer, err := cfg.NewDefaultCfgService(config, logger)
	if err != nil {
		logger.Error("配置服务初始化失败 %v", err)
		return nil, err
	}
	if err := cfgServer.Start(groupCtx, router, apiGroup); err != nil {
		logger.Error("配置服务启动失败", err)
		return nil, err
	}

	// 注册 /api/push 路由
	apiGroup.POST("/push", func(c *gin.Context) {
		var req struct {
			SessionID string `json:"session_id"`
			ID        string `json:"id"`
			Text      string `json:"text"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "bad request"})
			return
		}
		fmt.Println("当前MacSessionMap keys:", reflect.ValueOf(core.MacSessionMap).MapKeys())
		fmt.Println("收到推送请求，id:", req.ID)
		sessionID := req.SessionID
		if sessionID == "" && req.ID != "" {
			core.MacSessionMapLock.RLock()
			sessionID = core.MacSessionMap[req.ID]
			core.MacSessionMapLock.RUnlock()
			if sessionID == "" {
				c.JSON(404, gin.H{"error": "id not found"})
				return
			}
		}
		core.WsConnMapLock.RLock()
		handler, ok := core.WsConnMap[sessionID]
		core.WsConnMapLock.RUnlock()
		if !ok {
			c.JSON(404, gin.H{"error": "session not found"})
			return
		}
		handler.SpeakAndPlay(req.Text, 1, handler.GetTalkRound())
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 注册设备管理 API
	// 获取设备列表
	apiGroup.GET("/devices", func(c *gin.Context) {
		core.WsConnMapLock.RLock()
		devices := make([]map[string]interface{}, 0)
		for sessionID, handler := range core.WsConnMap {
			device := map[string]interface{}{
				"session_id":  sessionID,
				"device_id":   handler.GetDeviceID(),
				"client_id":   handler.GetClientID(),
				"device_name": handler.GetHeaders()["Device-Name"],
				"status":      "online",
				"last_seen":   time.Now().Format("2006-01-02 15:04:05"),
			}
			devices = append(devices, device)
		}
		core.WsConnMapLock.RUnlock()
		c.JSON(200, gin.H{"devices": devices})
	})

	// 绑定设备
	apiGroup.POST("/devices/bind", func(c *gin.Context) {
		var req struct {
			DeviceID   string `json:"device_id"`
			ClientID   string `json:"client_id"`
			DeviceName string `json:"device_name"`
			Token      string `json:"token"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "bad request"})
			return
		}
		// 这里可以添加设备绑定逻辑，比如保存到数据库或配置文件
		c.JSON(200, gin.H{"status": "ok", "message": "设备绑定成功"})
	})

	// 解绑设备
	apiGroup.DELETE("/devices/unbind/:device_id", func(c *gin.Context) {
		// 这里可以添加设备解绑逻辑
		c.JSON(200, gin.H{"status": "ok", "message": "设备解绑成功"})
	})

	// 获取设备状态
	apiGroup.GET("/devices/:device_id/status", func(c *gin.Context) {
		deviceID := c.Param("device_id")
		core.WsConnMapLock.RLock()
		var deviceStatus map[string]interface{}
		for sessionID, handler := range core.WsConnMap {
			if handler.GetDeviceID() == deviceID {
				deviceStatus = map[string]interface{}{
					"session_id": sessionID,
					"device_id":  handler.GetDeviceID(),
					"client_id":  handler.GetClientID(),
					"status":     "online",
					"last_seen":  time.Now().Format("2006-01-02 15:04:05"),
				}
				break
			}
		}
		core.WsConnMapLock.RUnlock()
		if deviceStatus == nil {
			c.JSON(404, gin.H{"error": "device not found"})
			return
		}
		c.JSON(200, deviceStatus)
	})

	// HTTP Server（支持优雅关机）
	httpServer := &http.Server{
		Addr:    ":" + strconv.Itoa(config.Web.Port),
		Handler: router,
	}

	// 注册Swagger文档路由
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	g.Go(func() error {
		logger.Info(fmt.Sprintf("Gin 服务已启动，访问地址: http://0.0.0.0:%d", config.Web.Port))

		// 在单独的 goroutine 中监听关闭信号
		go func() {
			<-groupCtx.Done()
			logger.Info("收到关闭信号，开始关闭HTTP服务...")

			// 创建关闭超时上下文
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				logger.Error("HTTP服务关闭失败", err)
			} else {
				logger.Info("HTTP服务已优雅关闭")
			}
		}()

		// ListenAndServe 返回 ErrServerClosed 时表示正常关闭
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP 服务启动失败", err)
			return err
		}
		return nil
	})

	logger.Info("已注册 /api/push 路由")
	return httpServer, nil
}

func GracefulShutdown(cancel context.CancelFunc, logger *utils.Logger, g *errgroup.Group) {
	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// 等待信号
	sig := <-sigChan
	logger.Info(fmt.Sprintf("接收到系统信号: %v，开始优雅关闭服务", sig))

	// 取消上下文，通知所有服务开始关闭
	cancel()

	// 等待所有服务关闭，设置超时保护
	done := make(chan error, 1)
	go func() {
		done <- g.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			logger.Error("服务关闭过程中出现错误", err)
			os.Exit(1)
		}
		logger.Info("所有服务已优雅关闭")
	case <-time.After(15 * time.Second):
		logger.Error("服务关闭超时，强制退出")
		os.Exit(1)
	}
}

func startServices(config *configs.Config, logger *utils.Logger, g *errgroup.Group, groupCtx context.Context) error {
	// 启动 WebSocket 服务
	if _, err := StartWSServer(config, logger, g, groupCtx); err != nil {
		return fmt.Errorf("启动 WebSocket 服务失败: %w", err)
	}

	// 启动 Http 服务
	if _, err := StartHttpServer(config, logger, g, groupCtx); err != nil {
		return fmt.Errorf("启动 Http 服务失败: %w", err)
	}

	return nil
}

func main() {
	// 加载配置和初始化日志系统
	config, logger, err := LoadConfigAndLogger()
	if err != nil {
		fmt.Println("加载配置或初始化日志系统失败:", err)
		os.Exit(1)
	}

	// 加载 .env 文件
	err = godotenv.Load()
	if err != nil {
		logger.Warn("未找到 .env 文件，使用系统环境变量")
	}

	// 初始化数据库连接
	db, dbType, err := database.InitDB(logger)
	_, _ = db, dbType // 避免未使用变量警告
	if err != nil {
		logger.Error(fmt.Sprintf("数据库连接失败: %v", err))
		return
	}

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 用 errgroup 管理两个服务
	g, groupCtx := errgroup.WithContext(ctx)

	// 启动所有服务
	if err := startServices(config, logger, g, groupCtx); err != nil {
		logger.Error("启动服务失败:", err)
		cancel()
		os.Exit(1)
	}

	// 启动优雅关机处理
	GracefulShutdown(cancel, logger, g)

	logger.Info("程序已成功退出")
}
