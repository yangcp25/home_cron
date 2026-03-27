package service

import (
	"context"
	"time"

	"gitlab.hudonggz.cn/yangchunping/go-infra/log"
	"go.uber.org/zap"
)

// TrendingJob GitHub Trending 定时任务
type TrendingJob struct {
	trendingService *TrendingService
	feishuService   *FeishuService
}

// NewTrendingJob 创建新的 TrendingJob
func NewTrendingJob() *TrendingJob {
	return &TrendingJob{
		trendingService: NewTrendingService(),
		feishuService:   nil, // 在需要时初始化
	}
}

// Run 执行 GitHub Trending 定时任务
func (j *TrendingJob) Run() {
	log.Info("⏰ GitHub Trending 定时任务触发", zap.Time("triggerTime", time.Now()))

	ctx := context.Background()

	// 1. 获取 GitHub Trending 数据
	result, err := j.trendingService.FetchAll(ctx)
	if err != nil {
		log.Error("❌ 获取 GitHub Trending 数据失败", zap.Error(err))
		return
	}

	// 2. 初始化飞书服务（如果还没有初始化）
	if j.feishuService == nil {
		feishuService, err := NewFeishuService()
		if err != nil {
			log.Error("初始化飞书服务失败", zap.Error(err))
			return
		}
		j.feishuService = feishuService
	}

	// 3. 发送飞书通知
	err = j.feishuService.SendTrending(ctx, result)
	if err != nil {
		log.Error("❌ 发送飞书通知失败", zap.Error(err))
		return
	}

	log.Info("✅ GitHub Trending 通知发送成功")
}

// RunOnce 立即执行一次 GitHub Trending 任务
func (j *TrendingJob) RunOnce() {
	log.Info("🚀 立即执行 GitHub Trending 任务", zap.Time("executeTime", time.Now()))

	ctx := context.Background()

	// 1. 获取 GitHub Trending 数据
	result, err := j.trendingService.FetchAll(ctx)
	if err != nil {
		log.Error("❌ 获取 GitHub Trending 数据失败", zap.Error(err))
		return
	}

	// 2. 初始化飞书服务（如果还没有初始化）
	if j.feishuService == nil {
		feishuService, err := NewFeishuService()
		if err != nil {
			log.Error("初始化飞书服务失败", zap.Error(err))
			return
		}
		j.feishuService = feishuService
	}

	// 3. 发送飞书通知
	err = j.feishuService.SendTrending(ctx, result)
	if err != nil {
		log.Error("❌ 发送飞书通知失败", zap.Error(err))
		return
	}

	log.Info("✅ GitHub Trending 通知发送成功")
}
