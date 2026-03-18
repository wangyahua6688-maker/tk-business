package home

import "strings"

// ListCategoryLibrary 分类库查询接口。
// 该接口用于“更多分类”页面，支持按 keyword 搜索并返回总量。
func (s *Service) ListCategoryLibrary(keyword string) (map[string]interface{}, error) {
	// 定义并初始化当前变量。
	rows, err := s.dao.SearchLotteryCategories(keyword, 300)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 判断条件并进入对应分支逻辑。
	if len(rows) == 0 {
		// 兼容降级：分类表不可用时，仍然允许前端进行关键字搜索。
		tags, tagErr := s.dao.ListLotteryCategoryTags(300)
		// 判断条件并进入对应分支逻辑。
		if tagErr != nil {
			// 返回当前处理结果。
			return nil, tagErr
		}
		// 更新当前变量或字段值。
		rows = tagsToCategories(filterTagsByKeyword(tags, keyword), false)
	}
	// 判断条件并进入对应分支逻辑。
	if len(rows) == 0 {
		// 二级降级：标签字段缺失时，使用图纸标题构建分类搜索结果。
		titles, titleErr := s.dao.ListLotteryTitles(300)
		// 判断条件并进入对应分支逻辑。
		if titleErr != nil {
			// 返回当前处理结果。
			return nil, titleErr
		}
		// 定义并初始化当前变量。
		filtered := filterTagsByKeyword(titles, keyword)
		// 更新当前变量或字段值。
		rows = titlesToCategories(filtered, false)
	}
	// 判断条件并进入对应分支逻辑。
	if len(rows) == 0 {
		// 三级降级：返回默认分类集合，避免分类库空白。
		rows = defaultHomeCategories()
	}

	// 返回当前处理结果。
	return map[string]interface{}{
		// 调用strings.TrimSpace完成当前处理。
		"keyword": strings.TrimSpace(keyword),
		// 调用buildCategoryPayload完成当前处理。
		"items": buildCategoryPayload(rows),
		// 调用len完成当前处理。
		"total": len(rows),
		// 处理当前语句逻辑。
	}, nil
}
