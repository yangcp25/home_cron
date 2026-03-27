package service

import (
	"context"
	"fmt"
	"homecron/config"
	"homecron/internal/model"
	"strings"
	"time"

	"gitlab.hudonggz.cn/yangchunping/go-infra/log"
	"gitlab.hudonggz.cn/yangchunping/go-infra/notify/feishu"
	"go.uber.org/zap"
)

// FeishuNotifier 定义飞书通知接口
type FeishuNotifier interface {
	SendTrending(ctx context.Context, result *model.TrendingResult) error
}

// FeishuService 实现FeishuNotifier接口
type FeishuService struct {
	client *feishu.Client
}

// NewFeishuService 创建新的FeishuService
func NewFeishuService() (*FeishuService, error) {
	cfg := config.AppConfig.Feishu
	if cfg.WebhookURL == "" {
		return nil, fmt.Errorf("飞书WebhookURL未配置")
	}

	// 创建飞书客户端
	opts := []feishu.Option{}
	if cfg.Secret != "" {
		opts = append(opts, feishu.WithSecret(cfg.Secret))
	}

	client := feishu.NewClient(cfg.WebhookURL, opts...)

	return &FeishuService{
		client: client,
	}, nil
}

// SendTrending 发送GitHub Trending数据到飞书
func (s *FeishuService) SendTrending(ctx context.Context, result *model.TrendingResult) error {
	if result == nil {
		return fmt.Errorf("trending结果为空")
	}

	// 构建消息内容
	content := s.buildTrendingMessage(result)

	// 发送卡片消息
	err := s.client.SendAlertCard(ctx, "GitHub Trending 每日报告", content, "blue")
	if err != nil {
		log.Error("发送飞书消息失败", zap.Error(err))
		return fmt.Errorf("发送飞书消息失败: %w", err)
	}

	log.Info("飞书消息发送成功")
	return nil
}

// buildTrendingMessage 构建trending消息内容
func (s *FeishuService) buildTrendingMessage(result *model.TrendingResult) string {
	var msg strings.Builder

	// 每日热门
	if len(result.Daily.Repos) > 0 {
		msg.WriteString("🔥 **今日热门**\n\n")
		for i, repo := range result.Daily.Repos {
			msg.WriteString(fmt.Sprintf("**%d. [%s](%s)**\n", i+1, repo.Fullname, repo.URL))
			msg.WriteString(fmt.Sprintf("⭐ %d (+%d) | %s\n", repo.Stars, repo.CurrentPeriodStars, repo.Language))
			if repo.Description != "" {
				msg.WriteString(fmt.Sprintf("💡 %s\n", repo.Description))
			}
			msg.WriteString("\n")
		}
	}

	// 每周热门
	if len(result.Weekly.Repos) > 0 {
		msg.WriteString("📈 **本周热门**\n\n")
		for i, repo := range result.Weekly.Repos {
			msg.WriteString(fmt.Sprintf("**%d. [%s](%s)**\n", i+1, repo.Fullname, repo.URL))
			msg.WriteString(fmt.Sprintf("⭐ %d (+%d) | %s\n", repo.Stars, repo.CurrentPeriodStars, repo.Language))
			if repo.Description != "" {
				msg.WriteString(fmt.Sprintf("💡 %s\n", repo.Description))
			}
			msg.WriteString("\n")
		}
	}

	// 每月热门
	if len(result.Monthly.Repos) > 0 {
		msg.WriteString("📊 **本月热门**\n\n")
		for i, repo := range result.Monthly.Repos {
			msg.WriteString(fmt.Sprintf("**%d. [%s](%s)**\n", i+1, repo.Fullname, repo.URL))
			msg.WriteString(fmt.Sprintf("⭐ %d (+%d) | %s\n", repo.Stars, repo.CurrentPeriodStars, repo.Language))
			if repo.Description != "" {
				msg.WriteString(fmt.Sprintf("💡 %s\n", repo.Description))
			}
			msg.WriteString("\n")
		}
	}

	// 如果没有数据
	if len(result.Daily.Repos) == 0 && len(result.Weekly.Repos) == 0 && len(result.Monthly.Repos) == 0 {
		msg.WriteString("📭 暂无GitHub Trending数据")
	}

	// 添加时间戳
	msg.WriteString(fmt.Sprintf("\n---\n📅 报告生成时间: %s", time.Now().Format("2006-01-02 15:04:05")))

	return msg.String()
}
