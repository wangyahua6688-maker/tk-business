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
	// 定义并初始化当前变量。
	now := homeNowInEast8()

	// 定义并初始化当前变量。
	banners, err := s.dao.ListHomeBanners()
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 定义并初始化当前变量。
	broadcasts, err := s.dao.ListBroadcasts(8)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 定义并初始化当前变量。
	homePopup, err := s.dao.GetActiveHomePopup()
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 定义并初始化当前变量。
	tabs, err := s.dao.ListSpecialLotteries()
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 定义并初始化当前变量。
	homeLinks, err := s.dao.ListHomeExternalLinks(18)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 定义并初始化当前变量。
	kingKongNavs, err := s.dao.ListHomeKingKongNav(20)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 定义并初始化当前变量。
	categories, err := s.dao.ListLotteryCategories(24, true)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 判断条件并进入对应分支逻辑。
	if len(categories) == 0 {
		// 兼容降级：当分类配置表还没初始化时，从开奖记录标签回推分类。
		tags, tagErr := s.dao.ListLotteryCategoryTags(24)
		// 判断条件并进入对应分支逻辑。
		if tagErr != nil {
			// 返回当前处理结果。
			return nil, tagErr
		}
		// 更新当前变量或字段值。
		categories = tagsToCategories(tags, true)
	}
	// 判断条件并进入对应分支逻辑。
	if len(categories) == 0 {
		// 二级降级：分类标签不可用时，从图纸标题回推分类。
		titles, titleErr := s.dao.ListLotteryTitles(24)
		// 判断条件并进入对应分支逻辑。
		if titleErr != nil {
			// 返回当前处理结果。
			return nil, titleErr
		}
		// 更新当前变量或字段值。
		categories = titlesToCategories(titles, true)
	}
	// 判断条件并进入对应分支逻辑。
	if len(categories) == 0 {
		// 三级降级：数据库无内容时，输出前端可展示的默认分类。
		categories = defaultHomeCategories()
	}

	// 定义并初始化当前变量。
	adBanners := make([]map[string]interface{}, 0)
	// 定义并初始化当前变量。
	officialBanners := make([]map[string]interface{}, 0)
	// 循环处理当前数据集合。
	for _, b := range banners {
		// 判断条件并进入对应分支逻辑。
		if !activeInTimeRange(now, b.StartAt, b.EndAt) {
			// 处理当前语句逻辑。
			continue
		}
		// 定义并初始化当前变量。
		item := map[string]interface{}{
			// 处理当前语句逻辑。
			"id": b.ID,
			// 处理当前语句逻辑。
			"title": b.Title,
			// 处理当前语句逻辑。
			"image_url": b.ImageURL,
			// 处理当前语句逻辑。
			"link_url": b.LinkURL,
			// 处理当前语句逻辑。
			"type": b.Type,
			// 处理当前语句逻辑。
			"positions": b.Positions,
			// 处理当前语句逻辑。
			"jump_type": b.JumpType,
			// 处理当前语句逻辑。
			"jump_post_id": b.JumpPostID,
			// 处理当前语句逻辑。
			"jump_url": b.JumpURL,
			// 处理当前语句逻辑。
			"content_html": b.ContentHTML,
		}
		// 判断条件并进入对应分支逻辑。
		if b.Type == "official" {
			// 更新当前变量或字段值。
			officialBanners = append(officialBanners, item)
			// 进入新的代码块进行处理。
		} else {
			// 更新当前变量或字段值。
			adBanners = append(adBanners, item)
		}
	}

	// 仅返回当前时间窗口内的首页弹窗。
	var popupPayload map[string]interface{}
	// 判断条件并进入对应分支逻辑。
	if homePopup != nil && activeInTimeRange(now, homePopup.StartAt, homePopup.EndAt) {
		// 更新当前变量或字段值。
		popupPayload = map[string]interface{}{
			// 处理当前语句逻辑。
			"id": homePopup.ID,
			// 处理当前语句逻辑。
			"title": homePopup.Title,
			// 处理当前语句逻辑。
			"content": homePopup.Content,
			// 处理当前语句逻辑。
			"image_url": homePopup.ImageURL,
			// 处理当前语句逻辑。
			"button_text": homePopup.ButtonText,
			// 处理当前语句逻辑。
			"button_link": homePopup.ButtonLink,
			// 处理当前语句逻辑。
			"show_once": homePopup.ShowOnce == 1,
		}
	}

	// 定义并初始化当前变量。
	tabPayload := make([]map[string]interface{}, 0, len(tabs))
	// 定义并初始化当前变量。
	activeID := uint(0)
	// 循环处理当前数据集合。
	for idx, t := range tabs {
		// 判断条件并进入对应分支逻辑。
		if idx == 0 {
			// 更新当前变量或字段值。
			activeID = t.ID
		}
		// 更新当前变量或字段值。
		tabPayload = append(tabPayload, map[string]interface{}{
			// 处理当前语句逻辑。
			"id": t.ID,
			// 处理当前语句逻辑。
			"name": t.Name,
			// 处理当前语句逻辑。
			"code": t.Code,
			// 处理当前语句逻辑。
			"current_issue": t.CurrentIssue,
			// 统一按“每天固定开奖时刻”输出最近一次未来时间。
			"next_draw_at": normalizeHomeNextDrawAt(t.NextDrawAt, now).Format(time.RFC3339),
		})
	}

	// 返回当前处理结果。
	return map[string]interface{}{
		// 处理当前语句逻辑。
		"title": "TK 客户端首页",
		// 调用now.Format完成当前处理。
		"server_time": now.Format(time.RFC3339),
		// 处理当前语句逻辑。
		"active_tab_id": activeID,
		// 处理当前语句逻辑。
		"special_lotteries": tabPayload,
		// 进入新的代码块进行处理。
		"banners": map[string]interface{}{
			// 处理当前语句逻辑。
			"ad": adBanners,
			// 处理当前语句逻辑。
			"official": officialBanners,
		},
		// 处理当前语句逻辑。
		"home_popup": popupPayload,
		// 处理当前语句逻辑。
		"broadcasts": broadcasts,
		// 调用buildHomeLinksPayload完成当前处理。
		"external_links": buildHomeLinksPayload(homeLinks),
		// 调用buildKingKongPayload完成当前处理。
		"kingkong_navs": buildKingKongPayload(kingKongNavs),
		// 调用buildCategoryPayload完成当前处理。
		"lottery_categories": buildCategoryPayload(categories),
		// 调用pickDefaultCategory完成当前处理。
		"default_category": pickDefaultCategory(categories),
		// 处理当前语句逻辑。
		"home_layout_version": "v2",
		// 处理当前语句逻辑。
	}, nil
}

// normalizeHomeNextDrawAt 将配置时间归一化为“未来最近一次每日开奖时刻”。
func normalizeHomeNextDrawAt(base time.Time, now time.Time) time.Time {
	loc := now.Location()
	hour, minute, second := base.Clock()
	next := time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		hour,
		minute,
		second,
		0,
		loc,
	)
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}
