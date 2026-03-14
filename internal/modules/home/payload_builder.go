package home

import (
	"strings"
	"time"

	"tk-common/models"
)

// activeInTimeRange 判断配置对象是否处于生效时间窗口内。
func activeInTimeRange(now time.Time, startAt *time.Time, endAt *time.Time) bool {
	if startAt != nil && now.Before(*startAt) {
		return false
	}
	if endAt != nil && now.After(*endAt) {
		return false
	}
	return true
}

// buildHomeLinksPayload 将外链模型转换为前端结构。
func buildHomeLinksPayload(rows []models.WExternalLink) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]interface{}{
			"id":       row.ID,
			"name":     row.Name,
			"url":      row.URL,
			"position": row.Position,
			"icon_url": row.IconURL,
		})
	}
	return out
}

// buildKingKongPayload 只保留“开奖现场”导航入口。
// 若后台未配置则回退到默认入口，确保前端页面可用。
func buildKingKongPayload(rows []models.WExternalLink) []map[string]interface{} {
	if len(rows) == 0 {
		return []map[string]interface{}{
			{
				"id":        0,
				"name":      "开奖现场",
				"url":       "/home/live-scene",
				"icon_url":  "",
				"group_key": "live_scene",
			},
		}
	}

	selected := rows[0]
	for _, row := range rows {
		if strings.Contains(strings.TrimSpace(row.Name), "开奖") {
			selected = row
			break
		}
	}

	name := strings.TrimSpace(selected.Name)
	if name == "" || !strings.Contains(name, "开奖") {
		name = "开奖现场"
	}
	return []map[string]interface{}{
		{
			"id":        selected.ID,
			"name":      name,
			"url":       selected.URL,
			"icon_url":  selected.IconURL,
			"group_key": selected.GroupKey,
		},
	}
}

// buildCategoryPayload 分类模型 -> 前端分类展示结构。
func buildCategoryPayload(rows []models.WLotteryCategory) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		key := strings.TrimSpace(row.CategoryKey)
		name := strings.TrimSpace(row.Name)
		if key == "" || name == "" {
			continue
		}
		out = append(out, map[string]interface{}{
			"key":          key,
			"name":         name,
			"show_on_home": row.ShowOnHome,
		})
	}
	return out
}

// pickDefaultCategory 选择首页默认激活分类。
func pickDefaultCategory(rows []models.WLotteryCategory) string {
	for _, item := range rows {
		key := strings.TrimSpace(item.CategoryKey)
		if key != "" {
			return key
		}
	}
	return ""
}

// tagsToCategories 将开奖标签列表映射为分类结构（降级路径）。
func tagsToCategories(tags []string, homeOnly bool) []models.WLotteryCategory {
	out := make([]models.WLotteryCategory, 0, len(tags))
	for idx, tag := range tags {
		name := strings.TrimSpace(tag)
		if name == "" {
			continue
		}
		showOnHome := int8(0)
		if !homeOnly || idx < 6 {
			showOnHome = 1
		}
		out = append(out, models.WLotteryCategory{
			CategoryKey: name,
			Name:        name,
			ShowOnHome:  showOnHome,
			Status:      1,
			Sort:        idx + 1,
		})
	}
	return out
}

// titlesToCategories 将图纸标题映射为分类结构（兜底路径）。
func titlesToCategories(titles []string, homeOnly bool) []models.WLotteryCategory {
	out := make([]models.WLotteryCategory, 0, len(titles))
	for idx, title := range titles {
		name := strings.TrimSpace(title)
		if name == "" {
			continue
		}
		showOnHome := int8(0)
		if !homeOnly || idx < 6 {
			showOnHome = 1
		}
		out = append(out, models.WLotteryCategory{
			CategoryKey: name,
			Name:        name,
			ShowOnHome:  showOnHome,
			Status:      1,
			Sort:        idx + 1,
		})
	}
	return out
}

// defaultHomeCategories 首页分类硬编码兜底（数据库未初始化时仍保证可展示）。
func defaultHomeCategories() []models.WLotteryCategory {
	names := []string{"九肖系列", "内幕系列", "四不像系列", "跑狗图系列", "挂牌系列", "更多"}
	out := make([]models.WLotteryCategory, 0, len(names))
	for idx, name := range names {
		out = append(out, models.WLotteryCategory{
			CategoryKey: name,
			Name:        name,
			ShowOnHome:  1,
			Status:      1,
			Sort:        idx + 1,
		})
	}
	return out
}

// filterTagsByKeyword 在降级路径里对标签执行关键字过滤。
func filterTagsByKeyword(tags []string, keyword string) []string {
	kw := strings.ToLower(strings.TrimSpace(keyword))
	if kw == "" {
		return tags
	}
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		name := strings.TrimSpace(tag)
		if name == "" {
			continue
		}
		if strings.Contains(strings.ToLower(name), kw) {
			out = append(out, name)
		}
	}
	return out
}
