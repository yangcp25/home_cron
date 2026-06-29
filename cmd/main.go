package main

import (
	"context"
	"fmt"
	"homecron/config"
	"os"
	_ "time/tzdata" // 内嵌时区数据库：容器缺 tzdata 时 LoadLocation("Asia/Shanghai") 仍可用

	"gitlab.hudonggz.cn/yangchunping/go-infra/log"
	"go.uber.org/zap"

	"homecron/internal/db"
	"homecron/internal/model"
	"homecron/internal/service"
)

func main() {
	// 1. 初始化基建日志
	log.Init(log.Config{Level: "info", Format: "console"})
	defer log.Sync()

	// 2. 加载 YAML 配置（作为定时任务的首次启动默认值）
	config.InitConfig("config/config.yaml")

	// 3. 初始化数据库连接（共享 softgen.db，WAL + busy_timeout 多进程安全）
	dbClient, err := db.Init(config.AppConfig.SoftgenDBDSN())
	if err != nil {
		log.Fatal("数据库启动失败", zap.Error(err))
	}

	// 4. 自动建表
	if err := dbClient.DB(context.Background()).AutoMigrate(
		&model.CronTaskConfig{},
	); err != nil {
		log.Fatal("数据库建表失败", zap.Error(err))
	}

	// 5. 让数据库成为定时任务配置的唯一来源（首次用 YAML 播种，之后以库为准）
	if err := config.InitCronTasksFromDB(dbClient); err != nil {
		log.Fatal("初始化定时任务配置失败", zap.Error(err))
	}

	// 🌟 调试：打印最终生效的定时任务配置
	fmt.Printf("🧐 当前生效的定时任务配置: %+v\n", config.AppConfig.CronTasks)

	// 6. 依赖注入
	planSvc := service.NewPlanService()

	// 7. 检查是否需要立即执行
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "trending":
			log.Info("🚀 立即执行 GitHub Trending 任务")
			planSvc.RunTrendingOnce()
			return
		case "software":
			log.Info("🚀 立即执行软著印钞机任务")
			planSvc.RunSoftwareOnce()
			return
		default:
			log.Warn("⚠️ 未知参数，支持的参数: trending, software")
		}
	}

	// 8. 发车！
	planSvc.StartCron()
}
