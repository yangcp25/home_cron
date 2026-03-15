package db

import (
	// 🌟 核心：给你的基础库取个别名 infradb，防止和当前包名冲突
	infradb "gitlab.hudonggz.cn/yangchunping/go-infra/db"
	"gitlab.hudonggz.cn/yangchunping/go-infra/log"
	"go.uber.org/zap"
)

// Init 实例化底层数据库连接，返回你基建库里的 *infradb.DB 对象
func Init(dbPath string) (*infradb.DB, error) {
	// 直接调用你基建库里的方法
	client, err := infradb.NewSqliteDB(dbPath)
	if err != nil {
		log.Error("实例化本地 SQLite 失败", zap.Error(err))
		return nil, err
	}

	log.Info("✅ 数据库实例化成功", zap.String("path", dbPath))
	return client, nil
}
