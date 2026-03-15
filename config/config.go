package configs

// 1. 最底层的具体任务配置
type AutoGenerateConfig struct {
	CronSpec  string   `yaml:"CronSpec"`
	TaskNames []string `yaml:"TaskNames"`
}

// 2. 中间层：管理所有定时任务的域
type CronTasksConfig struct {
	AutoGenerate AutoGenerateConfig `yaml:"AutoGenerate"`
	// 以后加任务直接往这儿塞，比如：
	// DataSync DataSyncConfig `yaml:"DataSync"`
}

// 3. 你的全局大配置结构体
type Config struct { // 之前可能叫 AppConfig 或 GlobalConfig
	// 🌟 挂载定时任务配置
	CronTasks CronTasksConfig `yaml:"CronTasks"`
}

// 全局变量保持不变
var AppConfig Config
