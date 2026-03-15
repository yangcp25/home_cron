// internal/model/task.go
package model

import "time"

// SoftwareTask 软著任务主表
type SoftwareTask struct {
	ID           uint      `gorm:"primaryKey"`
	SoftwareName string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	Status       string    `gorm:"type:varchar(20);index;default:'pending'"` // pending, processing, success, failed
	CurrentStep  string    `gorm:"type:varchar(50);default:'init'"`          // init, theme_done, code_done, screenshot_done
	ThemeConfig  string    `gorm:"type:text"`                                // 存放 Gemini 生成的主题 JSON
	CodePath     string    `gorm:"type:varchar(255)"`                        // 代码相对路径
	RetryCount   int       `gorm:"default:0"`
	LastError    string    `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

// TaskStepLog 步骤执行日志表
type TaskStepLog struct {
	ID        uint   `gorm:"primaryKey"`
	TaskID    uint   `gorm:"index;not null"`
	StepName  string `gorm:"type:varchar(50);not null"`
	Status    string `gorm:"type:varchar(20);not null"` // success, failed
	Message   string `gorm:"type:text"`
	CostMs    int64
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
