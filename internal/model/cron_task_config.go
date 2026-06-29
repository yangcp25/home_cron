package model

import "time"

// 定时任务名称常量（对应数据库里每一行的 Name）
const (
	CronTaskAutoGenerate   = "AutoGenerate"   // 软著印钞机
	CronTaskGithubTrending = "GithubTrending" // GitHub Trending 爬虫
)

// CronTaskConfig 定时任务配置表，每个定时任务一行。
//
// 通用字段（CronSpec / Enabled）直接落库，
// 各任务专属的差异化参数（如印钞机的 TaskNames、Trending 的 TopN）
// 统一以 JSON 形式存放在 Payload 字段里，方便后续扩展新任务而不动表结构。
type CronTaskConfig struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"` // AutoGenerate / GithubTrending
	CronSpec  string    `gorm:"type:varchar(100)" json:"cron_spec"`                // 定时表达式，空表示不注册
	Enabled   bool      `gorm:"default:true" json:"enabled"`                       // 是否启用
	Payload   string    `gorm:"type:text" json:"payload"`                          // 任务专属参数(JSON)
	Remark    string    `gorm:"type:varchar(255)" json:"remark"`                   // 备注
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (CronTaskConfig) TableName() string {
	return "cron_task_config"
}

// ===== Payload 结构定义 =====

// AutoGeneratePayload 软著印钞机的专属参数
type AutoGeneratePayload struct {
	TaskNames []string `json:"taskNames"`
}

// GithubTrendingPayload GitHub Trending 的专属参数
type GithubTrendingPayload struct {
	TopN int `json:"topN"`
}
