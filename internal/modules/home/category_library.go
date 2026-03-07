package home

import "strings"

// ListCategoryLibrary 分类库查询接口。
// 该接口用于“更多分类”页面，支持按 keyword 搜索并返回总量。
func (s *Service) ListCategoryLibrary(keyword string) (map[string]interface{}, error) {
	rows, err := s.dao.SearchLotteryCategories(keyword, 300)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		// 兼容降级：分类表不可用时，仍然允许前端进行关键字搜索。
		tags, tagErr := s.dao.ListLotteryCategoryTags(300)
		if tagErr != nil {
			return nil, tagErr
		}
		rows = tagsToCategories(filterTagsByKeyword(tags, keyword), false)
	}
	if len(rows) == 0 {
		// 二级降级：标签字段缺失时，使用图纸标题构建分类搜索结果。
		titles, titleErr := s.dao.ListLotteryTitles(300)
		if titleErr != nil {
			return nil, titleErr
		}
		filtered := filterTagsByKeyword(titles, keyword)
		rows = titlesToCategories(filtered, false)
	}
	if len(rows) == 0 {
		// 三级降级：返回默认分类集合，避免分类库空白。
		rows = defaultHomeCategories()
	}

	return map[string]interface{}{
		"keyword": strings.TrimSpace(keyword),
		"items":   buildCategoryPayload(rows),
		"total":   len(rows),
	}, nil
}
