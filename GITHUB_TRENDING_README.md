# GitHub Trending 爬虫功能说明

## 功能概述

本功能实现了自动爬取 GitHub Trending 仓库信息，并通过飞书机器人发送每日报告的功能。

## 功能特性

1. **自动爬取**: 每天早上 8:45 自动执行
2. **多周期支持**: 支持每日、每周、每月热门仓库
3. **飞书通知**: 通过飞书机器人发送格式化的卡片消息
4. **可配置**: 支持自定义定时规则和显示数量

## 配置说明

### 1. 飞书配置

在 `config/config.yaml` 中添加飞书配置：

```yaml
# 飞书通知配置
Feishu:
  WebhookURL: "https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-id"
  Secret: ""  # 如果机器人开启了签名校验，需要填写密钥
```

### 2. GitHub Trending 配置

```yaml
# GitHub Trending 爬虫任务
GithubTrending:
  CronSpec: "0 45 8 * * *"  # 每天早上8:45执行
  TopN: 10  # 每个周期显示10个热门仓库
```

## 文件结构

```
home_cron/
├── config/
│   ├── config.go          # 配置结构定义
│   └── config.yaml        # 配置文件
├── internal/
│   ├── model/
│   │   └── trending.go    # Trending 数据模型
│   └── service/
│       ├── trending_service.go  # GitHub Trending 爬虫服务
│       ├── feishu_service.go    # 飞书通知服务
│       └── plan.go              # 定时任务调度（已集成）
└── cmd/
    └── test_trending/     # 测试程序
```

## 使用说明

### 1. 安装依赖

确保系统已安装 `gtrending` 工具：

```bash
pip3 install gtrending
```

### 2. 配置飞书机器人

1. 在飞书中创建自定义机器人
2. 获取 Webhook URL
3. 将 URL 填入 `config/config.yaml` 的 `Feishu.WebhookURL` 字段

### 3. 运行程序

```bash
# 编译
go build -o homecron ./cmd/main.go

# 运行
./homecron
```

### 4. 测试功能

```bash
# 运行测试程序
go run ./cmd/test_trending/main.go
```

## 消息格式

飞书消息将包含以下内容：

- **🔥 今日热门**: 每日热门仓库列表
- **📈 本周热门**: 每周热门仓库列表
- **📊 本月热门**: 每月热门仓库列表

每个仓库显示：
- 仓库名称（带链接）
- Star 数量和新增 Star 数
- 编程语言
- 仓库描述

## 定时任务说明

- **GitHub Trending**: 每天早上 8:45 执行
- **软著印钞机**: 根据配置的时间执行

## 注意事项

1. 首次运行时，如果飞书 Webhook URL 未配置，会看到错误日志，这是正常的
2. 确保 `gtrending` 命令在系统 PATH 中可用
3. 程序使用 Asia/Shanghai 时区
4. 每个周期默认显示 10 个仓库，可在配置中调整

## 故障排查

### 1. 飞书通知失败

- 检查 Webhook URL 是否正确
- 检查网络连接
- 查看错误日志获取详细信息

### 2. gtrending 命令失败

```bash
# 检查是否安装
which gtrending

# 手动测试
gtrending repos --since daily --json
```

### 3. 定时任务不执行

- 检查 CronSpec 表达式是否正确
- 查看程序启动日志确认任务是否注册成功

## 开发说明

### 添加新的通知渠道

1. 在 `internal/service/` 下创建新的服务文件
2. 实现 `FeishuNotifier` 接口
3. 在 `plan.go` 中集成新服务

### 修改消息格式

编辑 `feishu_service.go` 中的 `buildTrendingMessage` 方法

### 调整定时规则

修改 `config/config.yaml` 中的 `CronSpec` 配置