package service

import (
	"context"
	configs "homecron/config"
	"sync/atomic"
	"time"

	"gitlab.hudonggz.cn/yangchunping/go-infra/httpc"
	"gitlab.hudonggz.cn/yangchunping/go-infra/log"
	"go.uber.org/zap"
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

// SoftwareJob 软著印钞机定时任务
type SoftwareJob struct {
	httpClient *httpc.Client
	taskNames  []string
	taskIndex  int32
}

// NewSoftwareJob 创建新的 SoftwareJob
func NewSoftwareJob() *SoftwareJob {
	taskCfg := configs.AppConfig.CronTasks.AutoGenerate
	tasks := taskCfg.TaskNames

	if len(tasks) == 0 {
		log.Warn("⚠️ 警告：TaskNames 为空，印钞机将轮空等待！")
	} else {
		log.Info("✅ 加载印钞机配置成功", zap.Int("count", len(tasks)))
	}

	return &SoftwareJob{
		httpClient: httpc.NewClient(),
		taskNames:  tasks,
		taskIndex:  0,
	}
}

// Run 执行软著印钞机定时任务
func (j *SoftwareJob) Run() {
	// 1. 原子获取当前索引，并让游标 +1
	currIdx := atomic.AddInt32(&j.taskIndex, 1) - 1

	// 2. 检查游标是否越界
	if int(currIdx) >= len(j.taskNames) {
		log.Info("🎉 软著印钞机：任务数组已全部处理完毕，当前定时器轮空跳过")
		return
	}

	targetName := j.taskNames[currIdx]
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
	resp, err := httpc.Post[GenerateResponse](ctx, j.httpClient, apiURL, reqBody,
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
}

// RunOnce 立即执行一次软著印钞机任务
func (j *SoftwareJob) RunOnce() {
	log.Info("🚀 立即执行软著印钞机任务", zap.Time("executeTime", time.Now()))

	// 检查是否有任务
	if len(j.taskNames) == 0 {
		log.Warn("⚠️ 没有可执行的任务")
		return
	}

	// 执行第一个任务
	targetName := j.taskNames[0]
	log.Info("⏰ 准备下发任务", zap.String("taskName", targetName))

	// 构造请求参数
	apiURL := "http://127.0.0.1:8080/api/generate"
	reqBody := GenerateRequest{
		Name:  targetName,
		Type:  "all",
		Model: 1,
	}

	// 发送请求
	ctx := context.Background()
	resp, err := httpc.Post[GenerateResponse](ctx, j.httpClient, apiURL, reqBody,
		httpc.WithTimeout(15*time.Second),
	)

	// 结果校验
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
}
