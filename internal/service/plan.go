package service

import (
	configs "homecron/config"

	"gitlab.hudonggz.cn/yangchunping/go-infra/log"
	"gitlab.hudonggz.cn/yangchunping/go-infra/workerv2"
	"go.uber.org/zap"
)

type PlanService struct {
	// 独立的任务实例
	softwareJob *SoftwareJob
	trendingJob *TrendingJob
}

// NewPlanService 构造函数
func NewPlanService() *PlanService {
	// 初始化独立的任务实例
	softwareJob := NewSoftwareJob()
	trendingJob := NewTrendingJob()

	return &PlanService{
		softwareJob: softwareJob,
		trendingJob: trendingJob,
	}
}

// StartCron 启动主调度器
func (s *PlanService) StartCron() {
	c, err := workerv2.New(workerv2.WithLocation("Asia/Shanghai"))
	if err != nil {
		log.Error("create cron fail", zap.Error(err))
		return
	}

	// 1. 添加软著印钞机定时任务（如果配置存在）
	softwareCronSpec := configs.AppConfig.CronTasks.AutoGenerate.CronSpec
	if softwareCronSpec != "" {
		_, err = c.Cron.AddFunc(softwareCronSpec, s.softwareJob.Run)
		if err != nil {
			log.Error("添加软著印钞机定时任务失败", zap.Error(err))
		} else {
			log.Info("🚀 软著印钞机定时流水线已启动",
				zap.String("module", "cron"),
				zap.String("spec", softwareCronSpec),
			)
		}
	} else {
		log.Info("⚠️ 软著印钞机定时任务未配置，跳过注册")
	}

	// 2. 添加 GitHub Trending 定时任务（如果配置存在）
	trendingCronSpec := configs.AppConfig.CronTasks.GithubTrending.CronSpec
	if trendingCronSpec != "" {
		_, err = c.Cron.AddFunc(trendingCronSpec, s.trendingJob.Run)
		if err != nil {
			log.Error("添加 GitHub Trending 定时任务失败", zap.Error(err))
		} else {
			log.Info("🚀 GitHub Trending 定时任务已启动",
				zap.String("module", "cron"),
				zap.String("spec", trendingCronSpec),
			)
		}
	} else {
		log.Info("⚠️ GitHub Trending 定时任务未配置，跳过注册")
	}

	// 启动定时器
	log.Info("🚀 定时任务调度器已启动")
	c.Start()
	select {}
}

// RunSoftwareOnce 立即执行一次软著印钞机任务
func (s *PlanService) RunSoftwareOnce() {
	s.softwareJob.RunOnce()
}

// RunTrendingOnce 立即执行一次 GitHub Trending 任务
func (s *PlanService) RunTrendingOnce() {
	s.trendingJob.RunOnce()
}
