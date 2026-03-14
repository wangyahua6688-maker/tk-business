package dao

import (
	"strings"
	"time"

	"tk-common/models"

	"gorm.io/gorm"
)

// LotteryDAO 开奖与彩种数据访问层。
type LotteryDAO struct {
	db *gorm.DB
}

// LotteryCommentRow 彩种详情评论聚合视图（评论 + 用户信息）。
type LotteryCommentRow struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	ParentID  uint      `json:"parent_id"`
	Content   string    `json:"content"`
	Likes     int64     `json:"likes"`
	CreatedAt time.Time `json:"created_at"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	UserType  string    `json:"user_type"`
}

func NewLotteryDAO(db *gorm.DB) *LotteryDAO {
	return &LotteryDAO{db: db}
}

// ListSpecialLotteries 返回首页/开奖现场需要的彩种切换标签。
func (d *LotteryDAO) ListSpecialLotteries(limit int) ([]models.WSpecialLottery, error) {
	rows := make([]models.WSpecialLottery, 0)
	q := d.db.Where("status = 1").Order("sort ASC, id ASC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&rows).Error
	return rows, err
}

func (d *LotteryDAO) GetSpecialLottery(id uint) (*models.WSpecialLottery, error) {
	var row models.WSpecialLottery
	if err := d.db.Where("id = ? AND status = 1", id).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (d *LotteryDAO) GetLatestLotteryInfoBySpecialID(sid uint) (*models.WLotteryInfo, error) {
	var row models.WLotteryInfo
	if err := d.db.Where("special_lottery_id = ? AND status = 1", sid).
		Order("is_current DESC, draw_at DESC, id DESC").First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (d *LotteryDAO) ListCards(category string, limit int) ([]models.WLotteryInfo, error) {
	var rows []models.WLotteryInfo
	// 1) 标准化分类参数，避免前后端传参带空格导致误筛选。
	category = strings.TrimSpace(category)
	q := d.db.Where("status = 1")
	if category != "" {
		// 2) 优先读取分类配置，拿到 category_key 对应的 name，扩大命中范围。
		var cfg models.WLotteryCategory
		if err := d.db.Select("category_key", "name").Where("status = 1 AND category_key = ?", category).First(&cfg).Error; err == nil {
			tagSet := []string{category}
			if cfg.Name != "" && cfg.Name != category {
				// name 与 key 不一致时，把 name 也加入匹配集合。
				tagSet = append(tagSet, cfg.Name)
			}
			// 优先按 category_tag 命中；兼容旧数据按标题模糊匹配。
			q = q.Where("category_tag IN ? OR title LIKE ?", tagSet, "%"+category+"%")
		} else {
			// 分类配置不存在时，兼容“标题即分类”的降级场景。
			q = q.Where("category_tag = ? OR title LIKE ?", category, "%"+category+"%")
		}
	}
	// 3) 保持统一排序：先 sort，再开奖时间，再主键。
	q = q.Order("sort ASC, draw_at DESC, id DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	// 4) 仅使用 tk_lottery_info 作为图纸唯一数据源。
	err := q.Find(&rows).Error
	if err == nil {
		return rows, nil
	}

	// 兼容旧库：当 tk_lottery_info 尚无 category_tag 列、或 tk_lottery_category 表未建时，
	// 自动降级为“仅按标题匹配”，避免首页分类切换后列表空白。
	if isUnknownColumnError(err) || isTableNotExistsError(err) {
		// 6) 旧库降级路径：不依赖 category_tag，仅按标题匹配。
		compatQ := d.db.Where("status = 1")
		if category != "" {
			compatQ = compatQ.Where("title LIKE ?", "%"+category+"%")
		}
		compatQ = compatQ.Order("sort ASC, draw_at DESC, id DESC")
		if limit > 0 {
			compatQ = compatQ.Limit(limit)
		}
		if compatErr := compatQ.Find(&rows).Error; compatErr != nil {
			return nil, compatErr
		}
		// 7) 降级查询成功后直接返回（不再读取 w_* 表）。
		return rows, nil
	}

	return nil, err
}

func (d *LotteryDAO) GetLotteryInfo(id uint) (*models.WLotteryInfo, error) {
	var row models.WLotteryInfo
	if err := d.db.Where("id = ? AND status = 1", id).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (d *LotteryDAO) ListLotteryInfosBySpecialID(sid uint, limit int) ([]models.WLotteryInfo, error) {
	var rows []models.WLotteryInfo
	err := d.db.Where("special_lottery_id = ? AND status = 1", sid).
		Order("draw_at DESC, id DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (d *LotteryDAO) ListOptionsByLotteryInfoID(infoID uint) ([]models.WLotteryOption, error) {
	var rows []models.WLotteryOption
	err := d.db.Where("lottery_info_id = ?", infoID).Order("sort ASC, id ASC").Find(&rows).Error
	return rows, err
}

// CreateMissingOptions 为指定图纸补齐缺失的竞猜选项（幂等）。
// 说明：
// - 仅插入当前不存在的 option_name，避免重复创建；
// - sort 按 names 传入顺序从 1 开始。
func (d *LotteryDAO) CreateMissingOptions(infoID uint, names []string) error {
	if len(names) == 0 {
		return nil
	}

	// 1) 先读取已有选项名，构建存在集合。
	existingRows := make([]models.WLotteryOption, 0)
	if err := d.db.Select("option_name").Where("lottery_info_id = ?", infoID).Find(&existingRows).Error; err != nil {
		return err
	}
	exists := make(map[string]struct{}, len(existingRows))
	for _, row := range existingRows {
		name := strings.TrimSpace(row.OptionName)
		if name == "" {
			continue
		}
		exists[name] = struct{}{}
	}

	// 2) 仅收集缺失项。
	toCreate := make([]models.WLotteryOption, 0, len(names))
	for idx, raw := range names {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		if _, ok := exists[name]; ok {
			continue
		}
		toCreate = append(toCreate, models.WLotteryOption{
			LotteryInfoID: infoID,
			OptionName:    name,
			Votes:         0,
			Sort:          idx + 1,
		})
	}
	if len(toCreate) == 0 {
		return nil
	}

	// 3) 批量写入缺失选项。
	return d.db.Create(&toCreate).Error
}

func (d *LotteryDAO) ListDetailBanners() ([]models.WBanner, error) {
	var rows []models.WBanner
	err := d.db.Where("status = 1 AND (position = ? OR FIND_IN_SET(?, positions) > 0)", "lottery_detail", "lottery_detail").
		Order("sort ASC, id DESC").Find(&rows).Error
	return rows, err
}

func (d *LotteryDAO) ListExternalLinks(position string, limit int) ([]models.WExternalLink, error) {
	var rows []models.WExternalLink
	q := d.db.Where("status = 1 AND position = ?", position).Order("sort ASC, id DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&rows).Error
	return rows, err
}

func (d *LotteryDAO) ListLotteryInfosByIDs(ids []uint) ([]models.WLotteryInfo, error) {
	if len(ids) == 0 {
		return []models.WLotteryInfo{}, nil
	}
	var rows []models.WLotteryInfo
	err := d.db.Where("id IN ? AND status = 1", ids).Find(&rows).Error
	return rows, err
}

func (d *LotteryDAO) ListRecommendFallback(sid, currentID uint, limit int) ([]models.WLotteryInfo, error) {
	var rows []models.WLotteryInfo
	err := d.db.Where("status = 1 AND special_lottery_id = ? AND id <> ?", sid, currentID).
		Order("draw_at DESC, id DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (d *LotteryDAO) ListLotteryComments(infoID uint, limit int, orderBy string, userTypes []string) ([]LotteryCommentRow, error) {
	rows := make([]LotteryCommentRow, 0)
	q := d.db.Table("tk_comment AS c").
		Select(`c.id, c.user_id, c.parent_id, c.content, c.likes, c.created_at,
			COALESCE(u.username, '') AS username,
			COALESCE(u.nickname, '') AS nickname,
			COALESCE(u.avatar, '') AS avatar,
			COALESCE(u.user_type, 'natural') AS user_type`).
		Joins("LEFT JOIN tk_users AS u ON u.id = c.user_id").
		Where("c.status = 1 AND c.lottery_info_id = ?", infoID)
	if len(userTypes) > 0 {
		q = q.Where("u.user_type IN ?", userTypes)
	}
	switch orderBy {
	case "hot":
		q = q.Order("c.likes DESC, c.id DESC")
	default:
		q = q.Order("c.created_at DESC, c.id DESC")
	}
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Scan(&rows).Error
	return rows, err
}

func (d *LotteryDAO) FindOption(optionID uint) (*models.WLotteryOption, error) {
	var row models.WLotteryOption
	if err := d.db.First(&row, optionID).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (d *LotteryDAO) GetVoteRecord(infoID uint, voterHash string) (*models.WLotteryVoteRecord, error) {
	var row models.WLotteryVoteRecord
	if err := d.db.Where("lottery_info_id = ? AND voter_hash = ?", infoID, voterHash).First(&row).Error; err != nil {
		return nil, err
	}
	return &row, nil
}

func (d *LotteryDAO) WithTx(fn func(tx *gorm.DB) error) error {
	return d.db.Transaction(fn)
}
