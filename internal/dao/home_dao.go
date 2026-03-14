package dao

import (
	"errors"
	"strings"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"tk-common/models"
)

// HomeDAO 首页数据访问层。
type HomeDAO struct {
	db *gorm.DB
}

func NewHomeDAO(db *gorm.DB) *HomeDAO {
	return &HomeDAO{db: db}
}

func (d *HomeDAO) ListHomeBanners() ([]models.WBanner, error) {
	var rows []models.WBanner
	err := d.db.Where("status = 1 AND (position = ? OR FIND_IN_SET(?, positions) > 0)", "home", "home").
		Order("sort ASC, id DESC").Find(&rows).Error
	return rows, err
}

func (d *HomeDAO) ListBroadcasts(limit int) ([]models.WBroadcast, error) {
	var rows []models.WBroadcast
	err := d.db.Where("status = 1").Order("sort ASC, id DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

// GetActiveHomePopup 查询首页首屏弹窗（按排序取第一条生效记录）。
func (d *HomeDAO) GetActiveHomePopup() (*models.WHomePopup, error) {
	row := models.WHomePopup{}
	err := d.db.
		Where("status = 1 AND position = ?", "home").
		Order("sort ASC, id DESC").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		// 表不存在时直接降级为空，不阻断首页主流程。
		if isTableNotExistsError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (d *HomeDAO) ListSpecialLotteries() ([]models.WSpecialLottery, error) {
	var rows []models.WSpecialLottery
	err := d.db.Where("status = 1").Order("sort ASC, id ASC").Find(&rows).Error
	return rows, err
}

func (d *HomeDAO) ListHomeExternalLinks(limit int) ([]models.WExternalLink, error) {
	// 仅使用 tk_external_link 作为外链唯一数据源。
	rows := make([]models.WExternalLink, 0)
	q := d.db.Where("status = 1").Order("sort ASC, id DESC")
	// limit > 0 时才截断，limit <= 0 表示不限制。
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (d *HomeDAO) ListHomeKingKongNav(limit int) ([]models.WExternalLink, error) {
	var rows []models.WExternalLink
	q := d.db.Where("status = 1 AND position IN ?", []string{"home_kingkong", "home_nav"}).
		Order("sort ASC, id DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&rows).Error
	return rows, err
}

func (d *HomeDAO) ListLotteryCategories(limit int, homeOnly bool) ([]models.WLotteryCategory, error) {
	rows := make([]models.WLotteryCategory, 0)
	q := d.db.Model(&models.WLotteryCategory{}).Where("status = 1")
	if homeOnly {
		q = q.Where("show_on_home = 1")
	}
	q = q.Order("sort ASC, id ASC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&rows).Error
	if isTableNotExistsError(err) {
		// 只保留 tk_* 读路径；tk_lottery_category 缺失时返回空，让上层走 tags/titles/default 兜底。
		return []models.WLotteryCategory{}, nil
	}
	return rows, err
}

func (d *HomeDAO) SearchLotteryCategories(keyword string, limit int) ([]models.WLotteryCategory, error) {
	rows := make([]models.WLotteryCategory, 0)
	q := d.db.Model(&models.WLotteryCategory{}).Where("status = 1")
	if kw := strings.TrimSpace(keyword); kw != "" {
		like := "%" + kw + "%"
		q = q.Where("name LIKE ? OR category_key LIKE ? OR search_keywords LIKE ?", like, like, like)
	}
	q = q.Order("sort ASC, id ASC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&rows).Error
	if isTableNotExistsError(err) {
		// 只保留 tk_* 读路径；缺表时返回空列表，由上层按默认分类兜底。
		return []models.WLotteryCategory{}, nil
	}
	return rows, err
}

// ListLotteryCategoryTags 回退接口：
// 当 tk_lottery_category 尚未配置时，从开奖内容里提取分类标签。
func (d *HomeDAO) ListLotteryCategoryTags(limit int) ([]string, error) {
	rows := make([]string, 0)
	q := d.db.Model(&models.WLotteryInfo{}).
		Where("status = 1 AND category_tag <> ''").
		Distinct("category_tag").
		Order("category_tag ASC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Pluck("category_tag", &rows).Error
	if isUnknownColumnError(err) || isTableNotExistsError(err) {
		return []string{}, nil
	}
	return rows, err
}

// ListLotteryTitles 回退接口：
// 当分类表与标签字段都不可用时，使用图纸标题构建分类候选。
func (d *HomeDAO) ListLotteryTitles(limit int) ([]string, error) {
	rows := make([]string, 0)
	q := d.db.Model(&models.WLotteryInfo{}).
		Where("status = 1 AND title <> ''").
		Order("draw_at DESC, id DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Pluck("title", &rows).Error
	if isUnknownColumnError(err) || isTableNotExistsError(err) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	return uniqNonEmptyStrings(rows), nil
}

// uniqNonEmptyStrings 去重并过滤空字符串，保持原始顺序。
func uniqNonEmptyStrings(input []string) []string {
	seen := make(map[string]struct{}, len(input))
	out := make([]string, 0, len(input))
	for _, raw := range input {
		v := strings.TrimSpace(raw)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func isTableNotExistsError(err error) bool {
	if err == nil {
		return false
	}
	var me *mysqlDriver.MySQLError
	if errors.As(err, &me) && me.Number == 1146 {
		return true
	}
	return false
}

func isUnknownColumnError(err error) bool {
	if err == nil {
		return false
	}
	var me *mysqlDriver.MySQLError
	if errors.As(err, &me) && me.Number == 1054 {
		return true
	}
	return false
}
