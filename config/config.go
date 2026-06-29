package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"homecron/internal/model"

	"github.com/spf13/viper"
	infradb "gitlab.hudonggz.cn/yangchunping/go-infra/db"
	"gitlab.hudonggz.cn/yangchunping/go-infra/log"
	"go.uber.org/zap"
)

// 1. 最底层的具体任务配置
type AutoGenerateConfig struct {
	CronSpec  string   `yaml:"CronSpec" mapstructure:"CronSpec"`
	TaskNames []string `yaml:"TaskNames" mapstructure:"TaskNames"`
}

// GitHub Trending 爬虫配置
type GithubTrendingConfig struct {
	CronSpec string `yaml:"CronSpec" mapstructure:"CronSpec"` // 定时任务表达式
	TopN     int    `yaml:"TopN" mapstructure:"TopN"`         // 每个周期显示的仓库数量，默认10
}

// 2. 中间层：管理所有定时任务的域
type CronTasksConfig struct {
	AutoGenerate   AutoGenerateConfig   `yaml:"AutoGenerate" mapstructure:"AutoGenerate"`
	GithubTrending GithubTrendingConfig `yaml:"GithubTrending" mapstructure:"GithubTrending"`
}

// 飞书通知配置
type FeishuConfig struct {
	WebhookURL string `yaml:"WebhookURL" mapstructure:"WebhookURL"` // 飞书机器人Webhook URL
	Secret     string `yaml:"Secret" mapstructure:"Secret"`         // 飞书机器人签名密钥
}

// 3. 你的全局大配置结构体
type Config struct {
	// 🌟 挂载定时任务配置
	CronTasks CronTasksConfig `yaml:"CronTasks" mapstructure:"CronTasks"`
	// 飞书通知配置
	Feishu FeishuConfig `yaml:"Feishu" mapstructure:"Feishu"`
	// SoftgenDBPath 共享数据库文件路径（与 softgen 同一个 softgen.db）。
	// 留空则用 DefaultSoftgenDBPath。两个进程共用一个 SQLite 文件，靠 WAL + busy_timeout 协调。
	SoftgenDBPath string `yaml:"SoftgenDBPath" mapstructure:"SoftgenDBPath"`
}

// DefaultSoftgenDBPath softgen 服务的 SQLite 文件绝对路径，作为共享库的默认位置。
const DefaultSoftgenDBPath = "/Users/ycp/work/code/own/code/softgen/softgen.db"

// SqliteDSN 给 glebarez/sqlite 拼带 PRAGMA 的 DSN：busy_timeout + journal_mode，
// 让 home_cron 与 softgen 两个进程安全共用同一个 softgen.db。
// 默认 WAL；环境变量 SOFTGEN_SQLITE_WAL=off 时退回 DELETE（兼容 Windows bind mount）。
func SqliteDSN(path string) string {
	journal := "WAL"
	if v := strings.ToLower(os.Getenv("SOFTGEN_SQLITE_WAL")); v == "off" || v == "false" || v == "0" {
		journal = "DELETE"
	}
	return fmt.Sprintf(
		"file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(%s)&_pragma=foreign_keys(1)",
		path, journal,
	)
}

// SoftgenDBDSN 返回最终生效的共享库 DSN。
//
// 优先级：环境变量 SOFTGEN_DB_FILE（容器里与 softgen 共用，指向挂载卷）
//
//	> 配置文件 SoftgenDBPath > 默认绝对路径。
func (c *Config) SoftgenDBDSN() string {
	p := os.Getenv("SOFTGEN_DB_FILE")
	if p == "" {
		p = c.SoftgenDBPath
	}
	if p == "" {
		p = DefaultSoftgenDBPath
	}
	return SqliteDSN(p)
}

// 全局变量保持不变
var AppConfig Config

// InitConfig 使用 Viper 读取并解析 yaml 配置文件
func InitConfig(path string) {
	// 1. 告诉 Viper 配置文件的准确路径
	viper.SetConfigFile(path)

	// 2. 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("❌ Viper 读取配置文件失败: %w (路径: %s)", err, path))
	}

	// 3. 将配置解析到 AppConfig 全局变量中
	if err := viper.Unmarshal(&AppConfig); err != nil {
		panic(fmt.Errorf("❌ Viper 解析配置到结构体失败: %w", err))
	}

	fmt.Printf("✅ Viper 配置加载成功，使用文件: %s\n", viper.ConfigFileUsed())
}

// DBClient 全局数据库客户端，供其他包按需使用
var DBClient *infradb.DB

// InitCronTasksFromDB 让数据库成为定时任务配置的唯一来源。
//
// 流程：
//  1. 用当前内存里的 YAML 配置作为「默认值」，首次运行时写入数据库（已存在则跳过）；
//  2. 再从数据库把配置读回来，覆盖 AppConfig.CronTasks。
//
// 这样以后改定时任务（CronSpec / TaskNames / TopN）直接改数据库即可，无需动 YAML。
func InitCronTasksFromDB(dbClient *infradb.DB) error {
	DBClient = dbClient

	if err := seedCronTasks(dbClient); err != nil {
		return fmt.Errorf("初始化定时任务默认配置失败: %w", err)
	}
	if err := loadCronTasksFromDB(dbClient); err != nil {
		return fmt.Errorf("从数据库加载定时任务配置失败: %w", err)
	}
	return nil
}

// seedCronTasks 首次启动时，把 YAML 里的默认配置写入数据库（按 Name 判断，已存在则跳过）
func seedCronTasks(dbClient *infradb.DB) error {
	ctx := context.Background()
	db := dbClient.DB(ctx)

	autoPayload, _ := json.Marshal(model.AutoGeneratePayload{
		TaskNames: AppConfig.CronTasks.AutoGenerate.TaskNames,
	})
	trendingPayload, _ := json.Marshal(model.GithubTrendingPayload{
		TopN: AppConfig.CronTasks.GithubTrending.TopN,
	})

	defaults := []model.CronTaskConfig{
		{
			Name:     model.CronTaskAutoGenerate,
			CronSpec: AppConfig.CronTasks.AutoGenerate.CronSpec,
			Enabled:  AppConfig.CronTasks.AutoGenerate.CronSpec != "",
			Payload:  string(autoPayload),
			Remark:   "软著印钞机",
		},
		{
			Name:     model.CronTaskGithubTrending,
			CronSpec: AppConfig.CronTasks.GithubTrending.CronSpec,
			Enabled:  AppConfig.CronTasks.GithubTrending.CronSpec != "",
			Payload:  string(trendingPayload),
			Remark:   "GitHub Trending 爬虫",
		},
	}

	for _, item := range defaults {
		var existing model.CronTaskConfig
		result := db.Where("name = ?", item.Name).First(&existing)
		if result.RowsAffected == 0 {
			if err := db.Create(&item).Error; err != nil {
				return err
			}
			log.Info("🌱 写入定时任务默认配置", zap.String("name", item.Name))
		}
	}
	return nil
}

// loadCronTasksFromDB 从数据库读取定时任务配置，覆盖内存里的 AppConfig.CronTasks
func loadCronTasksFromDB(dbClient *infradb.DB) error {
	ctx := context.Background()
	db := dbClient.DB(ctx)

	var rows []model.CronTaskConfig
	if err := db.Find(&rows).Error; err != nil {
		return err
	}

	for _, row := range rows {
		switch row.Name {
		case model.CronTaskAutoGenerate:
			var p model.AutoGeneratePayload
			_ = json.Unmarshal([]byte(row.Payload), &p)
			cfg := AutoGenerateConfig{TaskNames: p.TaskNames}
			if row.Enabled {
				cfg.CronSpec = row.CronSpec
			}
			AppConfig.CronTasks.AutoGenerate = cfg

		case model.CronTaskGithubTrending:
			var p model.GithubTrendingPayload
			_ = json.Unmarshal([]byte(row.Payload), &p)
			cfg := GithubTrendingConfig{TopN: p.TopN}
			if row.Enabled {
				cfg.CronSpec = row.CronSpec
			}
			AppConfig.CronTasks.GithubTrending = cfg

		default:
			log.Warn("⚠️ 数据库存在未知的定时任务配置", zap.String("name", row.Name))
		}
	}

	log.Info("✅ 已从数据库加载定时任务配置", zap.Int("count", len(rows)))
	return nil
}
