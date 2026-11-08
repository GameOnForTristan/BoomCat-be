// Package main 是 boom-cat-be 服务器的启动入口
// 负责加载配置、初始化依赖、组装路由、启动 HTTP 服务
package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"boom-cat-be/internal/config"
	"boom-cat-be/internal/controller"
	"boom-cat-be/internal/network"
	"boom-cat-be/internal/persistence"
	"boom-cat-be/internal/router"
	"boom-cat-be/internal/service"
	"boom-cat-be/pkg/logger"
)

func main() {
	configPath := flag.String("config", "internal/config/config.yaml", "path to config file")
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Default.Fatal("failed to load config", logger.F("error", err.Error()))
	}

	// 初始化日志
	initLogger(cfg.Log)
	logger.Default.Info("config loaded",
		logger.F("server_addr", cfg.ServerAddr()),
		logger.F("redis_addr", cfg.RedisAddr()),
		logger.F("db_host", cfg.Database.Host),
	)

	// 初始化持久化层
	ctx := context.Background()
	redisAdapter, err := persistence.NewRedisAdapter(cfg.Redis)
	if err != nil {
		logger.Default.Fatal("failed to init redis", logger.F("error", err.Error()))
	}
	defer redisAdapter.Close()
	logger.Default.Info("redis connected", logger.F("addr", cfg.RedisAddr()))

	db, err := persistence.NewDatabase(ctx, cfg.Database)
	if err != nil {
		logger.Default.Fatal("failed to init postgres", logger.F("error", err.Error()))
	}
	defer db.Close()
	logger.Default.Info("postgres connected", logger.F("host", cfg.Database.Host))

	// 初始化 Service 层
	authSvc := service.NewAuthService(db.User, db.Token, redisAdapter.Session, cfg.Auth)

	// Broadcaster - 管理所有 WebSocket 连接
	broadcaster := network.NewBroadcaster()

	// TurnService - 纯逻辑，无依赖
	turnSvc := service.NewTurnService()

	// RoomService 与 GameService 互依赖 -> 通过后置注入打破循环
	// 先创建 RoomService（GameService 字段为 nil）
	roomSvc := service.NewRoomService(
		redisAdapter.RoomMeta,
		redisAdapter.InviteCode,
		redisAdapter.RoomPlayers,
		cfg.Game.MinPlayers,
		cfg.Game.MaxPlayers,
	)

	// TimerService 依赖 RoomRegistry（RoomService 实现）
	timerSvc := service.NewTimerService(
		roomSvc,
		cfg.Game.TurnTimeoutSec,
		cfg.Game.NopeWindowSec,
		cfg.Game.ReconnectTimeoutSec,
	)

	// GameService 依赖 RoomRegistry + Broadcaster + TimerService + TurnService
	gameSvc := service.NewGameService(
		roomSvc,
		broadcaster,
		timerSvc,
		turnSvc,
		cfg.Game.TurnTimeoutSec,
		cfg.Game.NopeWindowSec,
		cfg.Game.ReconnectTimeoutSec,
	)

	// 后置注入 GameService 到 RoomService
	setGameService(roomSvc, gameSvc)

	// 初始化 Controller 层
	authCtrl := controller.NewAuthController(authSvc)
	roomCtrl := controller.NewRoomController(roomSvc, authSvc)
	wsCtrl := controller.NewWSController(authSvc, roomSvc, broadcaster)
	gameCtrl := controller.NewGameController(gameSvc, roomSvc, authSvc, broadcaster)

	// 组装路由
	handler := router.Build(&router.Dependencies{
		AuthCtrl: authCtrl,
		RoomCtrl: roomCtrl,
		WSCtrl:   wsCtrl,
		GameCtrl: gameCtrl,
	})

	// 启动 HTTP 服务
	server := &http.Server{
		Addr:         cfg.ServerAddr(),
		Handler:      handler,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	go func() {
		logger.Default.Info("server starting", logger.F("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Default.Fatal("server failed", logger.F("error", err.Error()))
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Default.Info("server shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Default.Error("server forced to shutdown", logger.F("error", err.Error()))
	}
	logger.Default.Info("server exited")
}

// initLogger 根据配置初始化 logger
func initLogger(cfg config.LogConfig) {
	level := logger.ParseLevel(cfg.Level)
	format := logger.ParseFormat(cfg.Format)

	var out *os.File
	switch cfg.Output {
	case "stderr":
		out = os.Stderr
	case "file":
		f, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			out = os.Stdout
		} else {
			out = f
		}
	default:
		out = os.Stdout
	}

	logger.Default.SetLevel(level)
	logger.Default.SetFormat(format)
	logger.Default.SetOutput(out)
}

// setGameService 后置注入 GameService 到 RoomService（打破循环依赖）
// RoomService 是接口，需要类型断言为具体指针类型访问 SetGameService
func setGameService(roomSvc service.RoomService, gs service.GameService) {
	type gameServiceSetter interface {
		SetGameService(service.GameService)
	}
	if s, ok := roomSvc.(gameServiceSetter); ok {
		s.SetGameService(gs)
		return
	}
	logger.Default.Fatal("RoomService does not implement SetGameService")
}
