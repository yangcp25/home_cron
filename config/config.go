package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// 1. 最底层的具体任务配置
type AutoGenerateConfig struct {
	CronSpec  string   `yaml:"CronSpec" mapstructure:"CronSpec"`
	TaskNames []string `yaml:"TaskNames" mapstructure:"TaskNames"`
}

// 2. 中间层：管理所有定时任务的域
type CronTasksConfig struct {
	AutoGenerate AutoGenerateConfig `yaml:"AutoGenerate" mapstructure:"AutoGenerate"`
}

// 3. 你的全局大配置结构体
type Config struct {
	// 🌟 挂载定时任务配置
	CronTasks CronTasksConfig `yaml:"CronTasks" mapstructure:"CronTasks"`
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
