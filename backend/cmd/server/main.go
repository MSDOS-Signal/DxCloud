package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dxcloud/cloud-api/internal/api"
	"github.com/dxcloud/cloud-api/internal/config"
	"github.com/dxcloud/cloud-api/internal/database"
	"github.com/dxcloud/cloud-api/internal/docker"
	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/internal/repository"
	"github.com/dxcloud/cloud-api/internal/scheduler"
	"github.com/dxcloud/cloud-api/internal/service"
	"github.com/dxcloud/cloud-api/migrations"
	"github.com/dxcloud/cloud-api/pkg/logger"
	"github.com/dxcloud/cloud-api/pkg/redisx"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.LogLevel)
	defer func() { _ = log.Sync() }()

	// 配置校验：生产环境强制安全 JWT 密钥，拒绝弱默认值启动
	if err := cfg.Validate(); err != nil {
		log.Fatal("config validation failed", zap.Error(err))
	}

	log.Info("cloud-api starting",
		zap.String("env", cfg.Env),
		zap.String("port", cfg.Port),
		zap.String("docker_host", cfg.DockerHost),
	)

	// MySQL：带重试连接（等待 compose 依赖 healthy）
	db, err := database.Connect(cfg, log)
	if err != nil {
		log.Fatal("mysql connect failed", zap.Error(err))
	}

	// 数据库迁移：按版本号顺序执行 migrations/*.sql，只增不改
	if err := database.Migrate(db, migrations.FS, log); err != nil {
		log.Fatal("migration failed", zap.Error(err))
	}

	// 幂等种子：权限点 / 系统角色 / 角色-权限 / 初始管理员
	if err := database.Seed(db, cfg, log); err != nil {
		log.Fatal("seed failed", zap.Error(err))
	}

	// Redis：带重试连接
	rdb, err := redisx.Connect(cfg, log)
	if err != nil {
		log.Fatal("redis connect failed", zap.Error(err))
	}

	// Docker Engine：唯一入口 Provider（fail-fast，避免"半活"状态）
	compute, err := docker.NewDockerProvider(cfg)
	if err != nil {
		log.Fatal("docker provider init failed", zap.Error(err))
	}

	repos := repository.New(db)
	iamSvc := iam.NewService(cfg, log, db, rdb, repos)
	quotaSvc := service.NewQuotaService(repos)
	billingSvc := service.NewBillingService(repos, log)
	ecsSvc := service.NewEcsService(repos, compute, iamSvc, rdb, quotaSvc, billingSvc, log)

	// Reconciler 后台对账（DB 期望态 ↔ Docker 实况）
	reconCtx, cancelRecon := context.WithCancel(context.Background())
	defer cancelRecon()
	go scheduler.NewReconciler(ecsSvc, log, 15*time.Second).Run(reconCtx)

	// 指标采样器（分钟级落库 + 7 天保留）
	go scheduler.NewMetricsCollector(repos, compute, log, 15*time.Second).Run(reconCtx)

	// 虚拟计费结算（每小时）
	go func() {
		t := time.NewTicker(time.Hour)
		defer t.Stop()
		for {
			select {
			case <-reconCtx.Done():
				return
			case <-t.C:
				if err := billingSvc.Collect(reconCtx); err != nil {
					log.Warn("billing collect failed", zap.Error(err))
				}
			}
		}
	}()

	router, pipeEngine, stopImageSvc := api.NewRouter(cfg, log, db, rdb, compute)
	defer stopImageSvc() // 服务关闭时取消所有进行中的镜像拉取

	// Pipeline 引擎：崩溃恢复 + 内嵌 Worker（Redis 队列消费）
	pipeCtx, cancelPipe := context.WithCancel(context.Background())
	defer cancelPipe()
	go pipeEngine.Start(pipeCtx)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Info("http server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("http server error", zap.Error(err))
		}
	}()

	// 优雅退出：SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down ...")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("shutdown error", zap.Error(err))
	}
	log.Info("server exited")
}
