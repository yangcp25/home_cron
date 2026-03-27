package service

import (
	"context"
	"encoding/json"
	"fmt"
	"homecron/config"
	"homecron/internal/model"
	"os/exec"
	"time"

	"gitlab.hudonggz.cn/yangchunping/go-infra/log"
	"go.uber.org/zap"
)

// TrendingFetcher 定义获取trending数据的接口
type TrendingFetcher interface {
	Fetch(ctx context.Context, since string) ([]model.TrendingRepo, error)
	FetchAll(ctx context.Context) (*model.TrendingResult, error)
}

// TrendingService 实现TrendingFetcher接口
type TrendingService struct {
	topN int // 每个周期显示的仓库数量
}

// NewTrendingService 创建新的TrendingService
func NewTrendingService() *TrendingService {
	topN := config.AppConfig.CronTasks.GithubTrending.TopN
	if topN <= 0 {
		topN = 10 // 默认显示10个仓库
	}

	return &TrendingService{
		topN: topN,
	}
}

// Fetch 获取指定时间周期的trending仓库
func (s *TrendingService) Fetch(ctx context.Context, since string) ([]model.TrendingRepo, error) {
	// 构建gtrending命令
	args := []string{"repos", "--since", since, "--json"}

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 执行命令
	cmd := exec.CommandContext(ctx, "gtrending", args...)

	// 获取命令输出
	output, err := cmd.Output()
	if err != nil {
		// 如果是上下文取消，返回超时错误
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("获取%s趋势数据超时", since)
		}

		// 获取标准错误输出
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("执行gtrending命令失败: %s, stderr: %s", err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("执行gtrending命令失败: %w", err)
	}

	// 解析JSON输出
	var repos []model.TrendingRepo
	if err := json.Unmarshal(output, &repos); err != nil {
		return nil, fmt.Errorf("解析JSON输出失败: %w, output: %s", err, string(output))
	}

	// 限制返回的仓库数量
	if len(repos) > s.topN {
		repos = repos[:s.topN]
	}

	// 设置since字段
	for i := range repos {
		repos[i].Since = since
	}

	log.Info("获取GitHub Trending数据成功",
		zap.String("since", since),
		zap.Int("count", len(repos)),
	)

	return repos, nil
}

// FetchAll 获取所有时间周期的trending数据
func (s *TrendingService) FetchAll(ctx context.Context) (*model.TrendingResult, error) {
	result := &model.TrendingResult{}

	// 获取每日趋势
	dailyRepos, err := s.Fetch(ctx, "daily")
	if err != nil {
		log.Error("获取每日趋势失败", zap.Error(err))
		// 即使失败也继续，返回空结果
		dailyRepos = []model.TrendingRepo{}
	}
	result.Daily = model.TrendingSnapshot{
		Since: "daily",
		Repos: dailyRepos,
	}

	// 获取每周趋势
	weeklyRepos, err := s.Fetch(ctx, "weekly")
	if err != nil {
		log.Error("获取每周趋势失败", zap.Error(err))
		weeklyRepos = []model.TrendingRepo{}
	}
	result.Weekly = model.TrendingSnapshot{
		Since: "weekly",
		Repos: weeklyRepos,
	}

	// 获取每月趋势
	monthlyRepos, err := s.Fetch(ctx, "monthly")
	if err != nil {
		log.Error("获取每月趋势失败", zap.Error(err))
		monthlyRepos = []model.TrendingRepo{}
	}
	result.Monthly = model.TrendingSnapshot{
		Since: "monthly",
		Repos: monthlyRepos,
	}

	log.Info("获取所有GitHub Trending数据完成",
		zap.Int("daily", len(result.Daily.Repos)),
		zap.Int("weekly", len(result.Weekly.Repos)),
		zap.Int("monthly", len(result.Monthly.Repos)),
	)

	return result, nil
}
