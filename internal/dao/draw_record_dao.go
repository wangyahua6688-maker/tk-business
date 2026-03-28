package dao

import (
	"errors"
	"strings"

	common_model "github.com/wangyahua6688-maker/tk-common/models"
	"gorm.io/gorm"
)

// GetLatestDrawRecordBySpecialID 查询指定彩种最新一期开奖记录。
func (d *LotteryDAO) GetLatestDrawRecordBySpecialID(sid uint) (*common_model.WDrawRecord, error) {
	// 声明当前变量。
	var row common_model.WDrawRecord
	// 判断条件并进入对应分支逻辑。
	if err := d.db.Where("special_lottery_id = ? AND status = 1", sid).
		// 调用Order完成当前处理。
		Order("is_current DESC, draw_at DESC, id DESC").
		// 调用First完成当前处理。
		First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		// 返回当前处理结果。
		return nil, err
	}
	// 返回当前处理结果。
	return &row, nil
}

// ListDrawRecordsBySpecialID 查询彩种历史开奖记录。
func (d *LotteryDAO) ListDrawRecordsBySpecialID(sid uint, limit int, orderMode string) ([]common_model.WDrawRecord, error) {
	// 定义并初始化当前变量。
	rows := make([]common_model.WDrawRecord, 0)
	// 定义并初始化当前变量。
	q := d.db.Where("status = 1")
	// 判断条件并进入对应分支逻辑。
	if sid > 0 {
		// 更新当前变量或字段值。
		q = q.Where("special_lottery_id = ?", sid)
	}
	// 判断条件并进入对应分支逻辑。
	if strings.EqualFold(strings.TrimSpace(orderMode), "asc") {
		// 更新当前变量或字段值。
		q = q.Order("draw_at ASC, id ASC")
		// 进入新的代码块进行处理。
	} else {
		// 更新当前变量或字段值。
		q = q.Order("draw_at DESC, id DESC")
	}
	// 判断条件并进入对应分支逻辑。
	if limit > 0 {
		// 更新当前变量或字段值。
		q = q.Limit(limit)
	}
	// 判断条件并进入对应分支逻辑。
	if err := q.Find(&rows).Error; err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 返回当前处理结果。
	return rows, nil
}

// GetDrawRecord 查询单条开奖记录。
func (d *LotteryDAO) GetDrawRecord(id uint) (*common_model.WDrawRecord, error) {
	// 声明当前变量。
	var row common_model.WDrawRecord
	// 判断条件并进入对应分支逻辑。
	if err := d.db.Where("id = ? AND status = 1", id).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		// 返回当前处理结果。
		return nil, err
	}
	// 返回当前处理结果。
	return &row, nil
}
