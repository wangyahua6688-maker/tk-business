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
	if err != nil {
		return nil, err
	}

	// 2) 将 DAO 模型转换为前端固定结构，保证字段命名稳定。
	items := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		items = append(items, map[string]interface{}{
			"id":                  r.ID,
			"special_lottery_id":  r.SpecialLotteryID,
			"category_id":         r.CategoryID,
			"category_tag":        r.CategoryTag,
			"title":               r.Title,
			"issue":               r.Issue,
			"cover_image_url":     r.CoverImageURL,
			"draw_code":           r.DrawCode,
			"normal_draw_result":  r.NormalDrawResult,
			"special_draw_result": r.SpecialDrawResult,
			"draw_result":         r.DrawResult,
			"playback_url":        r.PlaybackURL,
			"draw_at":             r.DrawAt.Format(time.RFC3339),
		})
	}
	return items, nil
}
