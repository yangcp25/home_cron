package service

import (
	"context"
	configs "homecron/config"
	"sync/atomic"
	"time"

	"gitlab.hudonggz.cn/yangchunping/go-infra/log"
	"gitlab.hudonggz.cn/yangchunping/go-infra/workerv2"
	"go.uber.org/zap"

	// 🌟 引入你的基建库
	infradb "gitlab.hudonggz.cn/yangchunping/go-infra/db"
	"gitlab.hudonggz.cn/yangchunping/go-infra/httpc"
	// 🌟 引入你的配置包 (请注意核对这个模块路径是否和你的项目一致)
)

// GenerateRequest 对应 POST 请求的 JSON Body
type GenerateRequest struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Model int    `json:"model"`
}

// GenerateResponse 对应 8080 服务返回的 JSON 结构
type GenerateResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Task    GenerateRequest `json:"task"`
}

type PlanService struct {
	dbClient   *infradb.DB
	httpClient *httpc.Client

	// 任务状态管理
	cronSpec  string
	taskNames []string
	taskIndex int32
}

// NewPlanService 构造函数
func NewPlanService(dbClient *infradb.DB) *PlanService {
	// 1. 从配置树的深处优雅地取值
	taskCfg := configs.AppConfig.CronTasks.AutoGenerate

	tasks := taskCfg.TaskNames
	cronSpec := taskCfg.CronSpec

	// 2. 防御性兜底
	if cronSpec == "" {
		cronSpec = "0 0 8-19 * * *"
		log.Warn("⚠️ 配置未设置 CronSpec，使用默认值", zap.String("spec", cronSpec))
	}

	if len(tasks) == 0 {
		log.Warn("⚠️ 警告：TaskNames 为空，印钞机将轮空等待！")
	} else {
		log.Info("✅ 加载印钞机配置成功",
			zap.Int("count", len(tasks)),
			zap.String("cron", cronSpec),
		)
	}

	return &PlanService{
		dbClient:   dbClient,
		httpClient: httpc.NewClient(),

		// 3. 赋值
		cronSpec:  cronSpec,
		taskNames: tasks,
		taskIndex: 0,
	}
}

// StartCron 启动主调度器
func (s *PlanService) StartCron() {
	c, err := workerv2.New(workerv2.WithLocation("Asia/Shanghai"))
	if err != nil {
		log.Error("create cron fail", zap.Error(err))
		return
	}

	_, err = c.Cron.AddFunc(s.cronSpec, func() {
		// 1. 原子获取当前索引，并让游标 +1
		currIdx := atomic.AddInt32(&s.taskIndex, 1) - 1

		// 2. 检查游标是否越界
		if int(currIdx) >= len(s.taskNames) {
			log.Info("🎉 软著印钞机：任务数组已全部处理完毕，当前定时器轮空跳过")
			return
		}

		targetName := s.taskNames[currIdx]
		log.Info("⏰ 定时任务触发，准备下发任务",
			zap.String("taskName", targetName),
			zap.Time("triggerTime", time.Now()),
		)

		// 3. 构造请求参数
		apiURL := "http://127.0.0.1:8080/api/generate"
		reqBody := GenerateRequest{
			Name:  targetName,
			Type:  "all",
			Model: 1,
		}

		// 4. 发送请求
		ctx := context.Background()
		resp, err := httpc.Post[GenerateResponse](ctx, s.httpClient, apiURL, reqBody,
			httpc.WithTimeout(15*time.Second),
		)

		// 5. 结果校验
		if err != nil {
			log.Error("❌ 任务发送到 8080 失败",
				zap.String("taskName", targetName),
				zap.Error(err),
			)
			return
		}

		if resp.Code != 0 {
			log.Warn("⚠️ 8080 接收了任务，但返回业务异常",
				zap.String("taskName", targetName),
				zap.Int("respCode", resp.Code),
				zap.String("respMsg", resp.Message),
			)
			return
		}

		log.Info("✅ 任务成功送达印钞机",
			zap.String("taskName", targetName),
			zap.String("serverMsg", resp.Message),
		)
	})

	if err != nil {
		log.Error("add func fail", zap.Error(err))
		return
	}
	log.Info("🚀 软著印钞机定时流水线已启动",
		zap.String("module", "cron"),
		zap.String("spec", s.cronSpec),
	)
	c.Start()
}
