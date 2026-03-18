package lottery

import (
	"strings"
	"time"
)

// ListCards 返回彩种列表卡片数据。
// category 为空时返回默认分类集合；不为空时按分类过滤。
func (s *Service) ListCards(category string) ([]map[string]interface{}, error) {
	// 1) 去掉前后空格，避免分类参数出现不可见字符导致空结果。
	rows, err := s.dao.ListCards(strings.TrimSpace(category), 36)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}

	// 2) 将 DAO 模型转换为前端固定结构，保证字段命名稳定。
	items := make([]map[string]interface{}, 0, len(rows))
	// 循环处理当前数据集合。
	for _, r := range rows {
		// 更新当前变量或字段值。
		items = append(items, map[string]interface{}{
			// 处理当前语句逻辑。
			"id": r.ID,
			// 处理当前语句逻辑。
			"special_lottery_id": r.SpecialLotteryID,
			// 处理当前语句逻辑。
			"category_id": r.CategoryID,
			// 处理当前语句逻辑。
			"category_tag": r.CategoryTag,
			// 处理当前语句逻辑。
			"title": r.Title,
			// 处理当前语句逻辑。
			"issue": r.Issue,
			// 处理当前语句逻辑。
			"cover_image_url": r.CoverImageURL,
			// 处理当前语句逻辑。
			"draw_code": r.DrawCode,
			// 处理当前语句逻辑。
			"normal_draw_result": r.NormalDrawResult,
			// 处理当前语句逻辑。
			"special_draw_result": r.SpecialDrawResult,
			// 处理当前语句逻辑。
			"draw_result": r.DrawResult,
			// 处理当前语句逻辑。
			"playback_url": r.PlaybackURL,
			// 调用r.DrawAt.Format完成当前处理。
			"draw_at": r.DrawAt.Format(time.RFC3339),
		})
	}
	// 返回当前处理结果。
	return items, nil
}
