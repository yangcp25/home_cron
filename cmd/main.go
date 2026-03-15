package main

import (
	"context"
	"fmt"
	"homecron/config"

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

	// 1. 必须先加载配置！把 config.yaml 读到 configs.AppConfig 里
	// 比如 configs.InitConfig() 或者 viper.ReadInConfig() 等等
	config.InitConfig("config/config.yaml")

	// 🌟 加一行调试代码，直接打印出来看看读到没有
	fmt.Printf("🧐 当前读取的配置: %+v\n", config.AppConfig.CronTasks)

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
