package dao

import (
	"strings"

	"tk-common/models"
)

// GetLatestDrawRecordBySpecialID 查询指定彩种最新一期开奖记录。
func (d *LotteryDAO) GetLatestDrawRecordBySpecialID(sid uint) (*models.WDrawRecord, error) {
	var row models.WDrawRecord
	if err := d.db.Where("special_lottery_id = ? AND status = 1", sid).
		Order("is_current DESC, draw_at DESC, id DESC").
		First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

// ListDrawRecordsBySpecialID 查询彩种历史开奖记录。
func (d *LotteryDAO) ListDrawRecordsBySpecialID(sid uint, limit int, orderMode string) ([]models.WDrawRecord, error) {
	rows := make([]models.WDrawRecord, 0)
	q := d.db.Where("status = 1")
	if sid > 0 {
		q = q.Where("special_lottery_id = ?", sid)
	}
	if strings.EqualFold(strings.TrimSpace(orderMode), "asc") {
		q = q.Order("draw_at ASC, id ASC")
	} else {
		q = q.Order("draw_at DESC, id DESC")
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// GetDrawRecord 查询单条开奖记录。
func (d *LotteryDAO) GetDrawRecord(id uint) (*models.WDrawRecord, error) {
	var row models.WDrawRecord
	if err := d.db.Where("id = ? AND status = 1", id).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}
