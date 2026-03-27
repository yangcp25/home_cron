package model

// TrendingRepo 代表一个GitHub trending仓库
type TrendingRepo struct {
	Fullname           string `json:"fullname"`           // 仓库全名，如 "golang/go"
	URL                string `json:"url"`                // 仓库URL
	Description        string `json:"description"`        // 仓库描述
	Language           string `json:"language"`           // 编程语言
	Stars              int    `json:"stars"`              // 总star数
	CurrentPeriodStars int    `json:"currentPeriodStars"` // 当前周期新增star数
	Since              string `json:"since"`              // 时间周期：daily, weekly, monthly
	// 以下字段来自gtrending输出，但不是必需的
	Name          string `json:"name,omitempty"`          // 仓库名称
	Author        string `json:"author,omitempty"`        // 作者
	Forks         int    `json:"forks,omitempty"`         // fork数
	LanguageColor string `json:"languageColor,omitempty"` // 语言颜色
}

// TrendingSnapshot 代表一个时间周期的trending快照
type TrendingSnapshot struct {
	Since string         `json:"since"` // 时间周期：daily, weekly, monthly
	Repos []TrendingRepo `json:"repos"` // 仓库列表
}

// TrendingResult 代表完整的trending结果
type TrendingResult struct {
	Daily   TrendingSnapshot `json:"daily"`   // 每日热门
	Weekly  TrendingSnapshot `json:"weekly"`  // 每周热门
	Monthly TrendingSnapshot `json:"monthly"` // 每月热门
}
