package database

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewMySQL 创建 tk-business 的 MySQL 连接。
func NewMySQL(dsn string) (*gorm.DB, error) {
	// 使用 Warn 级别日志：保留慢 SQL/异常信息，避免调试日志过量影响性能。
	// DSN 由 etc/business.yaml 注入，便于多环境切换。
	return gorm.Open(mysql.Open(dsn), &gorm.Config{
		// 调用logger.Default.LogMode完成当前处理。
		Logger: logger.Default.LogMode(logger.Warn),
	})
}
