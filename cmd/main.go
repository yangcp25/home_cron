package main

import (
	"fmt"
	"time"

	"gitlab.hudonggz.cn/yangchunping/go-infra/log3"
	"gitlab.hudonggz.cn/yangchunping/go-infra/workerv2"
	"go.uber.org/zap"
)

func main() {
	startCron()
}

func startCron() {
	// 创建秒级调度器
	c, err := workerv2.New(workerv2.WithLocation("Asia/Shanghai"))
	if err != nil {
		log3.Error("crate cron fail", zap.Any("err", err))
	}
	_, err = c.Cron.AddFunc("* * * * * *", func() {
		fmt.Println(time.Now().String())
	})
	if err != nil {
		return
	}
	c.Start()
	select {}
}
