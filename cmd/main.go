package main

import (
	"fmt"
	"homecron/config"
	"os"

	"gitlab.hudonggz.cn/yangchunping/go-infra/log"

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

	// 2. 依赖注入
	planSvc := service.NewPlanService()

	// 3. 检查是否需要立即执行
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

	// 4. 发车！
	planSvc.StartCron()
}
