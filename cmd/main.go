package main

import (
	"context"

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

	// 2. 实例化数据库 (调用咱们刚才写的那个工厂函数)
	dbClient, err := db.Init("softgen.db")
	if err != nil {
		log.Fatal("数据库启动失败", zap.Error(err))
	}

	// 自动建表
	err = dbClient.DB(context.Background()).AutoMigrate(
		&model.SoftwareTask{},
		&model.TaskStepLog{},
	)
	if err != nil {
		log.Fatal("建表失败", zap.Error(err))
	}

	// 3. 依赖注入
	planSvc := service.NewPlanService(dbClient)

	// 4. 发车！
	planSvc.StartCron()
}
