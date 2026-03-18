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
	// 处理当前语句逻辑。
	db *gorm.DB
}

// NewHomeDAO 创建HomeDAO实例。
func NewHomeDAO(db *gorm.DB) *HomeDAO {
	// 返回当前处理结果。
	return &HomeDAO{db: db}
}

// ListHomeBanners 查询HomeBanners列表。
func (d *HomeDAO) ListHomeBanners() ([]models.WBanner, error) {
	// 声明当前变量。
	var rows []models.WBanner
	// 定义并初始化当前变量。
	err := d.db.Where("status = 1 AND (position = ? OR FIND_IN_SET(?, positions) > 0)", "home", "home").
		// 调用Order完成当前处理。
		Order("sort ASC, id DESC").Find(&rows).Error
	// 返回当前处理结果。
	return rows, err
}

// ListBroadcasts 查询Broadcasts列表。
func (d *HomeDAO) ListBroadcasts(limit int) ([]models.WBroadcast, error) {
	// 声明当前变量。
	var rows []models.WBroadcast
	// 定义并初始化当前变量。
	err := d.db.Where("status = 1").Order("sort ASC, id DESC").Limit(limit).Find(&rows).Error
	// 返回当前处理结果。
	return rows, err
}

// GetActiveHomePopup 查询首页首屏弹窗（按排序取第一条生效记录）。
func (d *HomeDAO) GetActiveHomePopup() (*models.WHomePopup, error) {
	// 定义并初始化当前变量。
	row := models.WHomePopup{}
	// 定义并初始化当前变量。
	err := d.db.
		// 更新当前变量或字段值。
		Where("status = 1 AND position = ?", "home").
		// 调用Order完成当前处理。
		Order("sort ASC, id DESC").
		// 调用First完成当前处理。
		First(&row).Error
	// 判断条件并进入对应分支逻辑。
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 返回当前处理结果。
		return nil, nil
	}
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 表不存在时直接降级为空，不阻断首页主流程。
		if isTableNotExistsError(err) {
			// 返回当前处理结果。
			return nil, nil
		}
		// 返回当前处理结果。
		return nil, err
	}
	// 返回当前处理结果。
	return &row, nil
}

// ListSpecialLotteries 查询SpecialLotteries列表。
func (d *HomeDAO) ListSpecialLotteries() ([]models.WSpecialLottery, error) {
	// 声明当前变量。
	var rows []models.WSpecialLottery
	// 定义并初始化当前变量。
	err := d.db.Where("status = 1").Order("sort ASC, id ASC").Find(&rows).Error
	// 返回当前处理结果。
	return rows, err
}

// ListHomeExternalLinks 查询HomeExternalLinks列表。
func (d *HomeDAO) ListHomeExternalLinks(limit int) ([]models.WExternalLink, error) {
	// 仅使用 tk_external_link 作为外链唯一数据源。
	rows := make([]models.WExternalLink, 0)
	// 定义并初始化当前变量。
	q := d.db.Where("status = 1").Order("sort ASC, id DESC")
	// limit > 0 时才截断，limit <= 0 表示不限制。
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

// ListHomeKingKongNav 查询HomeKingKongNav列表。
func (d *HomeDAO) ListHomeKingKongNav(limit int) ([]models.WExternalLink, error) {
	// 声明当前变量。
	var rows []models.WExternalLink
	// 定义并初始化当前变量。
	q := d.db.Where("status = 1 AND position IN ?", []string{"home_kingkong", "home_nav"}).
		// 调用Order完成当前处理。
		Order("sort ASC, id DESC")
	// 判断条件并进入对应分支逻辑。
	if limit > 0 {
		// 更新当前变量或字段值。
		q = q.Limit(limit)
	}
	// 定义并初始化当前变量。
	err := q.Find(&rows).Error
	// 返回当前处理结果。
	return rows, err
}

// ListLotteryCategories 查询LotteryCategories列表。
func (d *HomeDAO) ListLotteryCategories(limit int, homeOnly bool) ([]models.WLotteryCategory, error) {
	// 定义并初始化当前变量。
	rows := make([]models.WLotteryCategory, 0)
	// 定义并初始化当前变量。
	q := d.db.Model(&models.WLotteryCategory{}).Where("status = 1")
	// 判断条件并进入对应分支逻辑。
	if homeOnly {
		// 更新当前变量或字段值。
		q = q.Where("show_on_home = 1")
	}
	// 更新当前变量或字段值。
	q = q.Order("sort ASC, id ASC")
	// 判断条件并进入对应分支逻辑。
	if limit > 0 {
		// 更新当前变量或字段值。
		q = q.Limit(limit)
	}
	// 定义并初始化当前变量。
	err := q.Find(&rows).Error
	// 判断条件并进入对应分支逻辑。
	if isTableNotExistsError(err) {
		// 只保留 tk_* 读路径；tk_lottery_category 缺失时返回空，让上层走 tags/titles/default 兜底。
		return []models.WLotteryCategory{}, nil
	}
	// 返回当前处理结果。
	return rows, err
}

// SearchLotteryCategories 处理SearchLotteryCategories相关逻辑。
func (d *HomeDAO) SearchLotteryCategories(keyword string, limit int) ([]models.WLotteryCategory, error) {
	// 定义并初始化当前变量。
	rows := make([]models.WLotteryCategory, 0)
	// 定义并初始化当前变量。
	q := d.db.Model(&models.WLotteryCategory{}).Where("status = 1")
	// 判断条件并进入对应分支逻辑。
	if kw := strings.TrimSpace(keyword); kw != "" {
		// 定义并初始化当前变量。
		like := "%" + kw + "%"
		// 更新当前变量或字段值。
		q = q.Where("name LIKE ? OR category_key LIKE ? OR search_keywords LIKE ?", like, like, like)
	}
	// 更新当前变量或字段值。
	q = q.Order("sort ASC, id ASC")
	// 判断条件并进入对应分支逻辑。
	if limit > 0 {
		// 更新当前变量或字段值。
		q = q.Limit(limit)
	}
	// 定义并初始化当前变量。
	err := q.Find(&rows).Error
	// 判断条件并进入对应分支逻辑。
	if isTableNotExistsError(err) {
		// 只保留 tk_* 读路径；缺表时返回空列表，由上层按默认分类兜底。
		return []models.WLotteryCategory{}, nil
	}
	// 返回当前处理结果。
	return rows, err
}

// ListLotteryCategoryTags 回退接口：
// 当 tk_lottery_category 尚未配置时，从开奖内容里提取分类标签。
func (d *HomeDAO) ListLotteryCategoryTags(limit int) ([]string, error) {
	// 定义并初始化当前变量。
	rows := make([]string, 0)
	// 定义并初始化当前变量。
	q := d.db.Model(&models.WLotteryInfo{}).
		// 更新当前变量或字段值。
		Where("status = 1 AND category_tag <> ''").
		// 调用Distinct完成当前处理。
		Distinct("category_tag").
		// 调用Order完成当前处理。
		Order("category_tag ASC")
	// 判断条件并进入对应分支逻辑。
	if limit > 0 {
		// 更新当前变量或字段值。
		q = q.Limit(limit)
	}
	// 定义并初始化当前变量。
	err := q.Pluck("category_tag", &rows).Error
	// 判断条件并进入对应分支逻辑。
	if isUnknownColumnError(err) || isTableNotExistsError(err) {
		// 返回当前处理结果。
		return []string{}, nil
	}
	// 返回当前处理结果。
	return rows, err
}

// ListLotteryTitles 回退接口：
// 当分类表与标签字段都不可用时，使用图纸标题构建分类候选。
func (d *HomeDAO) ListLotteryTitles(limit int) ([]string, error) {
	// 定义并初始化当前变量。
	rows := make([]string, 0)
	// 定义并初始化当前变量。
	q := d.db.Model(&models.WLotteryInfo{}).
		// 更新当前变量或字段值。
		Where("status = 1 AND title <> ''").
		// 调用Order完成当前处理。
		Order("draw_at DESC, id DESC")
	// 判断条件并进入对应分支逻辑。
	if limit > 0 {
		// 更新当前变量或字段值。
		q = q.Limit(limit)
	}
	// 定义并初始化当前变量。
	err := q.Pluck("title", &rows).Error
	// 判断条件并进入对应分支逻辑。
	if isUnknownColumnError(err) || isTableNotExistsError(err) {
		// 返回当前处理结果。
		return []string{}, nil
	}
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 返回当前处理结果。
	return uniqNonEmptyStrings(rows), nil
}

// uniqNonEmptyStrings 去重并过滤空字符串，保持原始顺序。
func uniqNonEmptyStrings(input []string) []string {
	// 定义并初始化当前变量。
	seen := make(map[string]struct{}, len(input))
	// 定义并初始化当前变量。
	out := make([]string, 0, len(input))
	// 循环处理当前数据集合。
	for _, raw := range input {
		// 定义并初始化当前变量。
		v := strings.TrimSpace(raw)
		// 判断条件并进入对应分支逻辑。
		if v == "" {
			// 处理当前语句逻辑。
			continue
		}
		// 判断条件并进入对应分支逻辑。
		if _, ok := seen[v]; ok {
			// 处理当前语句逻辑。
			continue
		}
		// 更新当前变量或字段值。
		seen[v] = struct{}{}
		// 更新当前变量或字段值。
		out = append(out, v)
	}
	// 返回当前处理结果。
	return out
}

// isTableNotExistsError 判断TableNotExistsError是否成立。
func isTableNotExistsError(err error) bool {
	// 判断条件并进入对应分支逻辑。
	if err == nil {
		// 返回当前处理结果。
		return false
	}
	// 声明当前变量。
	var me *mysqlDriver.MySQLError
	// 判断条件并进入对应分支逻辑。
	if errors.As(err, &me) && me.Number == 1146 {
		// 返回当前处理结果。
		return true
	}
	// 返回当前处理结果。
	return false
}

// isUnknownColumnError 判断UnknownColumnError是否成立。
func isUnknownColumnError(err error) bool {
	// 判断条件并进入对应分支逻辑。
	if err == nil {
		// 返回当前处理结果。
		return false
	}
	// 声明当前变量。
	var me *mysqlDriver.MySQLError
	// 判断条件并进入对应分支逻辑。
	if errors.As(err, &me) && me.Number == 1054 {
		// 返回当前处理结果。
		return true
	}
	// 返回当前处理结果。
	return false
}
