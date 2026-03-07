package home

import "time"

// BuildOverview 构建客户端首页聚合数据。
// 返回内容包括：
// - 顶部 banner（广告/官方通知）；
// - 广播消息；
// - 澳彩/港彩等开奖切换标签；
// - 外链区与金刚导航；
// - 首页可见的彩种分类。
func (s *Service) BuildOverview() (map[string]interface{}, error) {
	now := time.Now()

	banners, err := s.dao.ListHomeBanners()
	if err != nil {
		return nil, err
	}
	broadcasts, err := s.dao.ListBroadcasts(8)
	if err != nil {
		return nil, err
	}
	homePopup, err := s.dao.GetActiveHomePopup()
	if err != nil {
		return nil, err
	}
	tabs, err := s.dao.ListSpecialLotteries()
	if err != nil {
		return nil, err
	}
	homeLinks, err := s.dao.ListHomeExternalLinks(18)
	if err != nil {
		return nil, err
	}
	kingKongNavs, err := s.dao.ListHomeKingKongNav(20)
	if err != nil {
		return nil, err
	}
	categories, err := s.dao.ListLotteryCategories(24, true)
	if err != nil {
		return nil, err
	}
	if len(categories) == 0 {
		// 兼容降级：当分类配置表还没初始化时，从开奖记录标签回推分类。
		tags, tagErr := s.dao.ListLotteryCategoryTags(24)
		if tagErr != nil {
			return nil, tagErr
		}
		categories = tagsToCategories(tags, true)
	}
	if len(categories) == 0 {
		// 二级降级：分类标签不可用时，从图纸标题回推分类。
		titles, titleErr := s.dao.ListLotteryTitles(24)
		if titleErr != nil {
			return nil, titleErr
		}
		categories = titlesToCategories(titles, true)
	}
	if len(categories) == 0 {
		// 三级降级：数据库无内容时，输出前端可展示的默认分类。
		categories = defaultHomeCategories()
	}

	adBanners := make([]map[string]interface{}, 0)
	officialBanners := make([]map[string]interface{}, 0)
	for _, b := range banners {
		if !activeInTimeRange(now, b.StartAt, b.EndAt) {
			continue
		}
		item := map[string]interface{}{
			"id":           b.ID,
			"title":        b.Title,
			"image_url":    b.ImageURL,
			"link_url":     b.LinkURL,
			"type":         b.Type,
			"positions":    b.Positions,
			"jump_type":    b.JumpType,
			"jump_post_id": b.JumpPostID,
			"jump_url":     b.JumpURL,
			"content_html": b.ContentHTML,
		}
		if b.Type == "official" {
			officialBanners = append(officialBanners, item)
		} else {
			adBanners = append(adBanners, item)
		}
	}

	// 仅返回当前时间窗口内的首页弹窗。
	var popupPayload map[string]interface{}
	if homePopup != nil && activeInTimeRange(now, homePopup.StartAt, homePopup.EndAt) {
		popupPayload = map[string]interface{}{
			"id":          homePopup.ID,
			"title":       homePopup.Title,
			"content":     homePopup.Content,
			"image_url":   homePopup.ImageURL,
			"button_text": homePopup.ButtonText,
			"button_link": homePopup.ButtonLink,
			"show_once":   homePopup.ShowOnce == 1,
		}
	}

	tabPayload := make([]map[string]interface{}, 0, len(tabs))
	activeID := uint(0)
	for idx, t := range tabs {
		if idx == 0 {
			activeID = t.ID
		}
		tabPayload = append(tabPayload, map[string]interface{}{
			"id":            t.ID,
			"name":          t.Name,
			"code":          t.Code,
			"current_issue": t.CurrentIssue,
			"next_draw_at":  t.NextDrawAt.Format(time.RFC3339),
		})
	}

	return map[string]interface{}{
		"title":             "TK 客户端首页",
		"server_time":       now.Format(time.RFC3339),
		"active_tab_id":     activeID,
		"special_lotteries": tabPayload,
		"banners": map[string]interface{}{
			"ad":       adBanners,
			"official": officialBanners,
		},
		"home_popup":          popupPayload,
		"broadcasts":          broadcasts,
		"external_links":      buildHomeLinksPayload(homeLinks),
		"kingkong_navs":       buildKingKongPayload(kingKongNavs),
		"lottery_categories":  buildCategoryPayload(categories),
		"default_category":    pickDefaultCategory(categories),
		"home_layout_version": "v2",
	}, nil
}
