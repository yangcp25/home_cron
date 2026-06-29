package service

import (
	"context"
	"errors"
	configs "homecron/config"
	"time"

	"gitlab.hudonggz.cn/yangchunping/go-infra/httpc"
	"gitlab.hudonggz.cn/yangchunping/go-infra/log"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GenerateRequest 对应 POST /api/generate 的 JSON Body
type GenerateRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Account string `json:"account,omitempty"`
	Model   int    `json:"model"`
}

// GenerateResponse 对应 8080 服务返回的 JSON 结构
type GenerateResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Task    GenerateRequest `json:"task"`
}

// softwareTaskRow 只取生成调度需要的几列，映射共享库的 software_tasks 表。
type softwareTaskRow struct {
	ID           uint
	SoftwareName string
	AccountID    string
	Status       string
}

// SoftwareJob 软著印钞机定时任务（生成触发）。
//
// 旧版用内存里的固定 taskNames + 游标，进程一重启游标归零会重发。现在改为从共享库
// software_tasks 取最早的一条 status=pending：名字录入 = 往该表插 pending 行（后台 UI 负责）。
// 这样进程重启安全、每个名字状态可查、与 softgen 的任务生命周期天然统一。
type SoftwareJob struct {
	httpClient *httpc.Client
}

// NewSoftwareJob 创建 SoftwareJob
func NewSoftwareJob() *SoftwareJob {
	return &SoftwareJob{httpClient: httpc.NewClient()}
}

const generateAPIURL = "http://127.0.0.1:8080/api/generate"

// Run 定时触发：取一条 pending 任务下发生成。
//
// 8080 一次只跑一个任务。被「忙」拒单时不做任何状态改动，任务仍是 pending，
// 下一轮自然重试同一个——绝不跳过、绝不并发。
func (j *SoftwareJob) Run() {
	j.dispatchOnePending(false)
}

// RunOnce 立即执行一次（命令行 `software` 参数触发）
func (j *SoftwareJob) RunOnce() {
	log.Info("🚀 立即执行软著印钞机任务", zap.Time("executeTime", time.Now()))
	j.dispatchOnePending(true)
}

// dispatchOnePending 取最早的一条 pending 任务并下发生成。
func (j *SoftwareJob) dispatchOnePending(once bool) {
	ctx := context.Background()

	if configs.DBClient == nil {
		log.Error("❌ 共享数据库未初始化，软著印钞机跳过本轮")
		return
	}
	gdb := configs.DBClient.DB(ctx)

	// 1. 取最早的 pending 任务
	var task softwareTaskRow
	err := gdb.Table("software_tasks").
		Where("status = ?", "pending").
		Order("id ASC").
		First(&task).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Info("🎉 软著印钞机：没有待生成(pending)任务，本轮轮空")
			return
		}
		log.Error("❌ 查询待生成任务失败", zap.Error(err))
		return
	}

	log.Info("⏰ 定时任务触发，准备下发生成",
		zap.String("taskName", task.SoftwareName),
		zap.Uint("id", task.ID),
		zap.Time("triggerTime", time.Now()),
	)

	// 2. 下发到 8080
	reqBody := GenerateRequest{
		Name:    task.SoftwareName,
		Account: task.AccountID,
		Type:    "all",
		Model:   1,
	}
	resp, err := httpc.Post[GenerateResponse](ctx, j.httpClient, generateAPIURL, reqBody,
		httpc.WithTimeout(15*time.Second),
	)

	// 3. 网络/请求失败：不动状态，任务仍 pending，下一轮重试
	if err != nil {
		log.Error("❌ 生成请求发送到 8080 失败（任务保持 pending，下轮重试）",
			zap.String("taskName", task.SoftwareName), zap.Error(err))
		return
	}

	// 4. 8080 忙或拒单（最常见是「上一个还没走完」）：不动状态，下一轮重试同一个任务
	if resp.Code != 0 {
		log.Info("⏳ 8080 忙或拒单，本轮跳过、任务保持 pending，下轮重试",
			zap.String("taskName", task.SoftwareName),
			zap.Int("respCode", resp.Code),
			zap.String("respMsg", resp.Message))
		return
	}

	// 5. 接单成功：乐观把状态推到 generating，避免竞态窗口内被下一轮重复挑中。
	//    softgen 的 app.Run 随后也会把状态置 generating（幂等，无副作用）。
	if uErr := gdb.Table("software_tasks").
		Where("id = ? AND status = ?", task.ID, "pending").
		Update("status", "generating").Error; uErr != nil {
		log.Warn("⚠️ 乐观更新任务状态失败（不阻塞，softgen 侧会兜底置位）",
			zap.String("taskName", task.SoftwareName), zap.Error(uErr))
	}

	log.Info("✅ 任务成功送达印钞机",
		zap.String("taskName", task.SoftwareName),
		zap.String("serverMsg", resp.Message))
}
