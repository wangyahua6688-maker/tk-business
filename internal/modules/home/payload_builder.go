package home

import (
	"strings"
	"time"

	common_model "tk-common/models"
)

// activeInTimeRange 判断配置对象是否处于生效时间窗口内。
func activeInTimeRange(now time.Time, startAt *time.Time, endAt *time.Time) bool {
	// 判断条件并进入对应分支逻辑。
	if startAt != nil && now.Before(*startAt) {
		// 返回当前处理结果。
		return false
	}
	// 判断条件并进入对应分支逻辑。
	if endAt != nil && now.After(*endAt) {
		// 返回当前处理结果。
		return false
	}
	// 返回当前处理结果。
	return true
}

// buildHomeLinksPayload 将外链模型转换为前端结构。
func buildHomeLinksPayload(rows []common_model.WExternalLink) []map[string]interface{} {
	// 定义并初始化当前变量。
	out := make([]map[string]interface{}, 0, len(rows))
	// 循环处理当前数据集合。
	for _, row := range rows {
		// 更新当前变量或字段值。
		out = append(out, map[string]interface{}{
			// 处理当前语句逻辑。
			"id": row.ID,
			// 处理当前语句逻辑。
			"name": row.Name,
			// 处理当前语句逻辑。
			"url": row.URL,
			// 处理当前语句逻辑。
			"position": row.Position,
			// 处理当前语句逻辑。
			"icon_url": row.IconURL,
		})
	}
	// 返回当前处理结果。
	return out
}

// buildKingKongPayload 只保留“开奖现场”导航入口。
// 若后台未配置则回退到默认入口，确保前端页面可用。
func buildKingKongPayload(rows []common_model.WExternalLink) []map[string]interface{} {
	// 判断条件并进入对应分支逻辑。
	if len(rows) == 0 {
		// 返回当前处理结果。
		return []map[string]interface{}{
			{
				// 处理当前语句逻辑。
				"id": 0,
				// 处理当前语句逻辑。
				"name": "开奖现场",
				// 处理当前语句逻辑。
				"url": "/home/live-scene",
				// 处理当前语句逻辑。
				"icon_url": "",
				// 处理当前语句逻辑。
				"group_key": "live_scene",
			},
		}
	}

	// 定义并初始化当前变量。
	selected := rows[0]
	// 循环处理当前数据集合。
	for _, row := range rows {
		// 判断条件并进入对应分支逻辑。
		if strings.Contains(strings.TrimSpace(row.Name), "开奖") {
			// 更新当前变量或字段值。
			selected = row
			// 处理当前语句逻辑。
			break
		}
	}

	// 定义并初始化当前变量。
	name := strings.TrimSpace(selected.Name)
	// 判断条件并进入对应分支逻辑。
	if name == "" || !strings.Contains(name, "开奖") {
		// 更新当前变量或字段值。
		name = "开奖现场"
	}
	// 返回当前处理结果。
	return []map[string]interface{}{
		{
			// 处理当前语句逻辑。
			"id": selected.ID,
			// 处理当前语句逻辑。
			"name": name,
			// 处理当前语句逻辑。
			"url": selected.URL,
			// 处理当前语句逻辑。
			"icon_url": selected.IconURL,
			// 处理当前语句逻辑。
			"group_key": selected.GroupKey,
		},
	}
}

// buildCategoryPayload 分类模型 -> 前端分类展示结构。
func buildCategoryPayload(rows []common_model.WLotteryCategory) []map[string]interface{} {
	// 定义并初始化当前变量。
	out := make([]map[string]interface{}, 0, len(rows))
	// 循环处理当前数据集合。
	for _, row := range rows {
		// 定义并初始化当前变量。
		key := strings.TrimSpace(row.CategoryKey)
		// 定义并初始化当前变量。
		name := strings.TrimSpace(row.Name)
		// 判断条件并进入对应分支逻辑。
		if key == "" || name == "" {
			// 处理当前语句逻辑。
			continue
		}
		// 更新当前变量或字段值。
		out = append(out, map[string]interface{}{
			// 处理当前语句逻辑。
			"key": key,
			// 处理当前语句逻辑。
			"name": name,
			// 处理当前语句逻辑。
			"show_on_home": row.ShowOnHome,
		})
	}
	// 返回当前处理结果。
	return out
}

// pickDefaultCategory 选择首页默认激活分类。
func pickDefaultCategory(rows []common_model.WLotteryCategory) string {
	// 循环处理当前数据集合。
	for _, item := range rows {
		// 定义并初始化当前变量。
		key := strings.TrimSpace(item.CategoryKey)
		// 判断条件并进入对应分支逻辑。
		if key != "" {
			// 返回当前处理结果。
			return key
		}
	}
	// 返回当前处理结果。
	return ""
}

// tagsToCategories 将开奖标签列表映射为分类结构（降级路径）。
func tagsToCategories(tags []string, homeOnly bool) []common_model.WLotteryCategory {
	// 定义并初始化当前变量。
	out := make([]common_model.WLotteryCategory, 0, len(tags))
	// 循环处理当前数据集合。
	for idx, tag := range tags {
		// 定义并初始化当前变量。
		name := strings.TrimSpace(tag)
		// 判断条件并进入对应分支逻辑。
		if name == "" {
			// 处理当前语句逻辑。
			continue
		}
		// 定义并初始化当前变量。
		showOnHome := int8(0)
		// 判断条件并进入对应分支逻辑。
		if !homeOnly || idx < 6 {
			// 更新当前变量或字段值。
			showOnHome = 1
		}
		// 更新当前变量或字段值。
		out = append(out, common_model.WLotteryCategory{
			// 处理当前语句逻辑。
			CategoryKey: name,
			// 处理当前语句逻辑。
			Name: name,
			// 处理当前语句逻辑。
			ShowOnHome: showOnHome,
			// 处理当前语句逻辑。
			Status: 1,
			// 处理当前语句逻辑。
			Sort: idx + 1,
		})
	}
	// 返回当前处理结果。
	return out
}

// titlesToCategories 将图纸标题映射为分类结构（兜底路径）。
func titlesToCategories(titles []string, homeOnly bool) []common_model.WLotteryCategory {
	// 定义并初始化当前变量。
	out := make([]common_model.WLotteryCategory, 0, len(titles))
	// 循环处理当前数据集合。
	for idx, title := range titles {
		// 定义并初始化当前变量。
		name := strings.TrimSpace(title)
		// 判断条件并进入对应分支逻辑。
		if name == "" {
			// 处理当前语句逻辑。
			continue
		}
		// 定义并初始化当前变量。
		showOnHome := int8(0)
		// 判断条件并进入对应分支逻辑。
		if !homeOnly || idx < 6 {
			// 更新当前变量或字段值。
			showOnHome = 1
		}
		// 更新当前变量或字段值。
		out = append(out, common_model.WLotteryCategory{
			// 处理当前语句逻辑。
			CategoryKey: name,
			// 处理当前语句逻辑。
			Name: name,
			// 处理当前语句逻辑。
			ShowOnHome: showOnHome,
			// 处理当前语句逻辑。
			Status: 1,
			// 处理当前语句逻辑。
			Sort: idx + 1,
		})
	}
	// 返回当前处理结果。
	return out
}

// defaultHomeCategories 首页分类硬编码兜底（数据库未初始化时仍保证可展示）。
func defaultHomeCategories() []common_model.WLotteryCategory {
	// 定义并初始化当前变量。
	names := []string{"九肖系列", "内幕系列", "四不像系列", "跑狗图系列", "挂牌系列", "更多"}
	// 定义并初始化当前变量。
	out := make([]common_model.WLotteryCategory, 0, len(names))
	// 循环处理当前数据集合。
	for idx, name := range names {
		// 更新当前变量或字段值。
		out = append(out, common_model.WLotteryCategory{
			// 处理当前语句逻辑。
			CategoryKey: name,
			// 处理当前语句逻辑。
			Name: name,
			// 处理当前语句逻辑。
			ShowOnHome: 1,
			// 处理当前语句逻辑。
			Status: 1,
			// 处理当前语句逻辑。
			Sort: idx + 1,
		})
	}
	// 返回当前处理结果。
	return out
}

// filterTagsByKeyword 在降级路径里对标签执行关键字过滤。
func filterTagsByKeyword(tags []string, keyword string) []string {
	// 定义并初始化当前变量。
	kw := strings.ToLower(strings.TrimSpace(keyword))
	// 判断条件并进入对应分支逻辑。
	if kw == "" {
		// 返回当前处理结果。
		return tags
	}
	// 定义并初始化当前变量。
	out := make([]string, 0, len(tags))
	// 循环处理当前数据集合。
	for _, tag := range tags {
		// 定义并初始化当前变量。
		name := strings.TrimSpace(tag)
		// 判断条件并进入对应分支逻辑。
		if name == "" {
			// 处理当前语句逻辑。
			continue
		}
		// 判断条件并进入对应分支逻辑。
		if strings.Contains(strings.ToLower(name), kw) {
			// 更新当前变量或字段值。
			out = append(out, name)
		}
	}
	// 返回当前处理结果。
	return out
}
