package dao

import (
	"strings"
	"time"

	"tk-common/models"

	"gorm.io/gorm"
)

// LotteryDAO 开奖与彩种数据访问层。
type LotteryDAO struct {
	// 处理当前语句逻辑。
	db *gorm.DB
}

// LotteryCommentRow 彩种详情评论聚合视图（评论 + 用户信息）。
type LotteryCommentRow struct {
	// 处理当前语句逻辑。
	ID uint `json:"id"`
	// 处理当前语句逻辑。
	UserID uint `json:"user_id"`
	// 处理当前语句逻辑。
	ParentID uint `json:"parent_id"`
	// 处理当前语句逻辑。
	Content string `json:"content"`
	// 处理当前语句逻辑。
	Likes int64 `json:"likes"`
	// 处理当前语句逻辑。
	CreatedAt time.Time `json:"created_at"`
	// 处理当前语句逻辑。
	Username string `json:"username"`
	// 处理当前语句逻辑。
	Nickname string `json:"nickname"`
	// 处理当前语句逻辑。
	Avatar string `json:"avatar"`
	// 处理当前语句逻辑。
	UserType string `json:"user_type"`
}

// NewLotteryDAO 创建LotteryDAO实例。
func NewLotteryDAO(db *gorm.DB) *LotteryDAO {
	// 返回当前处理结果。
	return &LotteryDAO{db: db}
}

// ListSpecialLotteries 返回首页/开奖现场需要的彩种切换标签。
func (d *LotteryDAO) ListSpecialLotteries(limit int) ([]models.WSpecialLottery, error) {
	// 定义并初始化当前变量。
	rows := make([]models.WSpecialLottery, 0)
	// 定义并初始化当前变量。
	q := d.db.Where("status = 1").Order("sort ASC, id ASC")
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

// GetSpecialLottery 获取SpecialLottery。
func (d *LotteryDAO) GetSpecialLottery(id uint) (*models.WSpecialLottery, error) {
	// 声明当前变量。
	var row models.WSpecialLottery
	// 判断条件并进入对应分支逻辑。
	if err := d.db.Where("id = ? AND status = 1", id).First(&row).Error; err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 返回当前处理结果。
	return &row, nil
}

// GetLatestLotteryInfoBySpecialID 获取LatestLotteryInfoBySpecialID。
func (d *LotteryDAO) GetLatestLotteryInfoBySpecialID(sid uint) (*models.WLotteryInfo, error) {
	// 声明当前变量。
	var row models.WLotteryInfo
	// 判断条件并进入对应分支逻辑。
	if err := d.db.Where("special_lottery_id = ? AND status = 1", sid).
		// 调用Order完成当前处理。
		Order("is_current DESC, draw_at DESC, id DESC").First(&row).Error; err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 返回当前处理结果。
	return &row, nil
}

// ListCards 查询Cards列表。
func (d *LotteryDAO) ListCards(category string, limit int) ([]models.WLotteryInfo, error) {
	// 声明当前变量。
	var rows []models.WLotteryInfo
	// 1) 标准化分类参数，避免前后端传参带空格导致误筛选。
	category = strings.TrimSpace(category)
	// 定义并初始化当前变量。
	q := d.db.Where("status = 1")
	// 判断条件并进入对应分支逻辑。
	if category != "" {
		// 2) 优先读取分类配置，拿到 category_key 对应的 name，扩大命中范围。
		var cfg models.WLotteryCategory
		// 判断条件并进入对应分支逻辑。
		if err := d.db.Select("category_key", "name").Where("status = 1 AND category_key = ?", category).First(&cfg).Error; err == nil {
			// 定义并初始化当前变量。
			tagSet := []string{category}
			// 判断条件并进入对应分支逻辑。
			if cfg.Name != "" && cfg.Name != category {
				// name 与 key 不一致时，把 name 也加入匹配集合。
				tagSet = append(tagSet, cfg.Name)
			}
			// 优先按 category_tag 命中；兼容旧数据按标题模糊匹配。
			q = q.Where("category_tag IN ? OR title LIKE ?", tagSet, "%"+category+"%")
			// 进入新的代码块进行处理。
		} else {
			// 分类配置不存在时，兼容“标题即分类”的降级场景。
			q = q.Where("category_tag = ? OR title LIKE ?", category, "%"+category+"%")
		}
	}
	// 3) 保持统一排序：先 sort，再开奖时间，再主键。
	q = q.Order("sort ASC, draw_at DESC, id DESC")
	// 判断条件并进入对应分支逻辑。
	if limit > 0 {
		// 更新当前变量或字段值。
		q = q.Limit(limit)
	}
	// 4) 仅使用 tk_lottery_info 作为图纸唯一数据源。
	err := q.Find(&rows).Error
	// 判断条件并进入对应分支逻辑。
	if err == nil {
		// 返回当前处理结果。
		return rows, nil
	}

	// 兼容旧库：当 tk_lottery_info 尚无 category_tag 列、或 tk_lottery_category 表未建时，
	// 自动降级为“仅按标题匹配”，避免首页分类切换后列表空白。
	if isUnknownColumnError(err) || isTableNotExistsError(err) {
		// 6) 旧库降级路径：不依赖 category_tag，仅按标题匹配。
		compatQ := d.db.Where("status = 1")
		// 判断条件并进入对应分支逻辑。
		if category != "" {
			// 更新当前变量或字段值。
			compatQ = compatQ.Where("title LIKE ?", "%"+category+"%")
		}
		// 更新当前变量或字段值。
		compatQ = compatQ.Order("sort ASC, draw_at DESC, id DESC")
		// 判断条件并进入对应分支逻辑。
		if limit > 0 {
			// 更新当前变量或字段值。
			compatQ = compatQ.Limit(limit)
		}
		// 判断条件并进入对应分支逻辑。
		if compatErr := compatQ.Find(&rows).Error; compatErr != nil {
			// 返回当前处理结果。
			return nil, compatErr
		}
		// 7) 降级查询成功后直接返回（不再读取 w_* 表）。
		return rows, nil
	}

	// 返回当前处理结果。
	return nil, err
}

// GetLotteryInfo 获取LotteryInfo。
func (d *LotteryDAO) GetLotteryInfo(id uint) (*models.WLotteryInfo, error) {
	// 声明当前变量。
	var row models.WLotteryInfo
	// 判断条件并进入对应分支逻辑。
	if err := d.db.Where("id = ? AND status = 1", id).First(&row).Error; err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 返回当前处理结果。
	return &row, nil
}

// ListLotteryInfosBySpecialID 查询LotteryInfosBySpecialID列表。
func (d *LotteryDAO) ListLotteryInfosBySpecialID(sid uint, limit int) ([]models.WLotteryInfo, error) {
	// 声明当前变量。
	var rows []models.WLotteryInfo
	// 定义并初始化当前变量。
	err := d.db.Where("special_lottery_id = ? AND status = 1", sid).
		// 调用Order完成当前处理。
		Order("draw_at DESC, id DESC").Limit(limit).Find(&rows).Error
	// 返回当前处理结果。
	return rows, err
}

// ListOptionsByLotteryInfoID 查询OptionsByLotteryInfoID列表。
func (d *LotteryDAO) ListOptionsByLotteryInfoID(infoID uint) ([]models.WLotteryOption, error) {
	// 声明当前变量。
	var rows []models.WLotteryOption
	// 定义并初始化当前变量。
	err := d.db.Where("lottery_info_id = ?", infoID).Order("sort ASC, id ASC").Find(&rows).Error
	// 返回当前处理结果。
	return rows, err
}

// CreateMissingOptions 为指定图纸补齐缺失的竞猜选项（幂等）。
// 说明：
// - 仅插入当前不存在的 option_name，避免重复创建；
// - sort 按 names 传入顺序从 1 开始。
func (d *LotteryDAO) CreateMissingOptions(infoID uint, names []string) error {
	// 判断条件并进入对应分支逻辑。
	if len(names) == 0 {
		// 返回当前处理结果。
		return nil
	}

	// 1) 先读取已有选项名，构建存在集合。
	existingRows := make([]models.WLotteryOption, 0)
	// 判断条件并进入对应分支逻辑。
	if err := d.db.Select("option_name").Where("lottery_info_id = ?", infoID).Find(&existingRows).Error; err != nil {
		// 返回当前处理结果。
		return err
	}
	// 定义并初始化当前变量。
	exists := make(map[string]struct{}, len(existingRows))
	// 循环处理当前数据集合。
	for _, row := range existingRows {
		// 定义并初始化当前变量。
		name := strings.TrimSpace(row.OptionName)
		// 判断条件并进入对应分支逻辑。
		if name == "" {
			// 处理当前语句逻辑。
			continue
		}
		// 更新当前变量或字段值。
		exists[name] = struct{}{}
	}

	// 2) 仅收集缺失项。
	toCreate := make([]models.WLotteryOption, 0, len(names))
	// 循环处理当前数据集合。
	for idx, raw := range names {
		// 定义并初始化当前变量。
		name := strings.TrimSpace(raw)
		// 判断条件并进入对应分支逻辑。
		if name == "" {
			// 处理当前语句逻辑。
			continue
		}
		// 判断条件并进入对应分支逻辑。
		if _, ok := exists[name]; ok {
			// 处理当前语句逻辑。
			continue
		}
		// 更新当前变量或字段值。
		toCreate = append(toCreate, models.WLotteryOption{
			// 处理当前语句逻辑。
			LotteryInfoID: infoID,
			// 处理当前语句逻辑。
			OptionName: name,
			// 处理当前语句逻辑。
			Votes: 0,
			// 处理当前语句逻辑。
			Sort: idx + 1,
		})
	}
	// 判断条件并进入对应分支逻辑。
	if len(toCreate) == 0 {
		// 返回当前处理结果。
		return nil
	}

	// 3) 批量写入缺失选项。
	return d.db.Create(&toCreate).Error
}

// ListDetailBanners 查询DetailBanners列表。
func (d *LotteryDAO) ListDetailBanners() ([]models.WBanner, error) {
	// 声明当前变量。
	var rows []models.WBanner
	// 定义并初始化当前变量。
	err := d.db.Where("status = 1 AND (position = ? OR FIND_IN_SET(?, positions) > 0)", "lottery_detail", "lottery_detail").
		// 调用Order完成当前处理。
		Order("sort ASC, id DESC").Find(&rows).Error
	// 返回当前处理结果。
	return rows, err
}

// ListExternalLinks 查询ExternalLinks列表。
func (d *LotteryDAO) ListExternalLinks(position string, limit int) ([]models.WExternalLink, error) {
	// 声明当前变量。
	var rows []models.WExternalLink
	// 定义并初始化当前变量。
	q := d.db.Where("status = 1 AND position = ?", position).Order("sort ASC, id DESC")
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

// ListLotteryInfosByIDs 查询LotteryInfosByIDs列表。
func (d *LotteryDAO) ListLotteryInfosByIDs(ids []uint) ([]models.WLotteryInfo, error) {
	// 判断条件并进入对应分支逻辑。
	if len(ids) == 0 {
		// 返回当前处理结果。
		return []models.WLotteryInfo{}, nil
	}
	// 声明当前变量。
	var rows []models.WLotteryInfo
	// 定义并初始化当前变量。
	err := d.db.Where("id IN ? AND status = 1", ids).Find(&rows).Error
	// 返回当前处理结果。
	return rows, err
}

// ListRecommendFallback 查询RecommendFallback列表。
func (d *LotteryDAO) ListRecommendFallback(sid, currentID uint, limit int) ([]models.WLotteryInfo, error) {
	// 声明当前变量。
	var rows []models.WLotteryInfo
	// 定义并初始化当前变量。
	err := d.db.Where("status = 1 AND special_lottery_id = ? AND id <> ?", sid, currentID).
		// 调用Order完成当前处理。
		Order("draw_at DESC, id DESC").Limit(limit).Find(&rows).Error
	// 返回当前处理结果。
	return rows, err
}

// ListLotteryComments 查询LotteryComments列表。
func (d *LotteryDAO) ListLotteryComments(infoID uint, limit int, orderBy string, userTypes []string) ([]LotteryCommentRow, error) {
	// 定义并初始化当前变量。
	rows := make([]LotteryCommentRow, 0)
	// 定义并初始化当前变量。
	q := d.db.Table("tk_comment AS c").
		// 调用Select完成当前处理。
		Select(`c.id, c.user_id, c.parent_id, c.content, c.likes, c.created_at,
			COALESCE(u.username, '') AS username,
			COALESCE(u.nickname, '') AS nickname,
			COALESCE(u.avatar, '') AS avatar,
			COALESCE(u.user_type, 'natural') AS user_type`).
		// 更新当前变量或字段值。
		Joins("LEFT JOIN tk_users AS u ON u.id = c.user_id").
		// 更新当前变量或字段值。
		Where("c.status = 1 AND c.lottery_info_id = ?", infoID)
	// 判断条件并进入对应分支逻辑。
	if len(userTypes) > 0 {
		// 更新当前变量或字段值。
		q = q.Where("u.user_type IN ?", userTypes)
	}
	// 根据表达式进入多分支处理。
	switch orderBy {
	case "hot":
		// 更新当前变量或字段值。
		q = q.Order("c.likes DESC, c.id DESC")
	default:
		// 更新当前变量或字段值。
		q = q.Order("c.created_at DESC, c.id DESC")
	}
	// 判断条件并进入对应分支逻辑。
	if limit > 0 {
		// 更新当前变量或字段值。
		q = q.Limit(limit)
	}
	// 定义并初始化当前变量。
	err := q.Scan(&rows).Error
	// 返回当前处理结果。
	return rows, err
}

// FindOption 查找Option。
func (d *LotteryDAO) FindOption(optionID uint) (*models.WLotteryOption, error) {
	// 声明当前变量。
	var row models.WLotteryOption
	// 判断条件并进入对应分支逻辑。
	if err := d.db.First(&row, optionID).Error; err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 返回当前处理结果。
	return &row, nil
}

// GetVoteRecord 获取VoteRecord。
func (d *LotteryDAO) GetVoteRecord(infoID uint, voterHash string) (*models.WLotteryVoteRecord, error) {
	// 声明当前变量。
	var row models.WLotteryVoteRecord
	// 判断条件并进入对应分支逻辑。
	if err := d.db.Where("lottery_info_id = ? AND voter_hash = ?", infoID, voterHash).First(&row).Error; err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 返回当前处理结果。
	return &row, nil
}

// WithTx 基于Tx执行操作。
func (d *LotteryDAO) WithTx(fn func(tx *gorm.DB) error) error {
	// 返回当前处理结果。
	return d.db.Transaction(fn)
}
