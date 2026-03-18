package lottery

import (
	"context"

	"tk-common/models"
)

// BuildDetail 构建彩票详情页数据。
// 该接口聚合彩种详情核心区块：
// - 当期图纸与开奖号；
// - 年份/期号切换；
// - 投票区；
// - 评论区（系统/网友/热门/最新）；
// - 推荐图纸与外链。
func (s *Service) BuildDetail(ctx context.Context, infoID uint) (map[string]interface{}, error) {
	// 1) 查询当前图纸主记录。
	current, err := s.dao.GetLotteryInfo(infoID)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}

	// 2) 查询同彩种下的历史期号，用于详情页切换。
	list, err := s.dao.ListLotteryInfosBySpecialID(current.SpecialLotteryID, 30)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 3) 查询投票选项（若缺失则自动补默认选项）。
	options, err := s.ensurePollOptions(current.ID)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 4) 查询详情页 banner。
	banners, err := s.dao.ListDetailBanners()
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 5) 查询详情页外链推荐区。
	links, err := s.dao.ListExternalLinks("lottery_detail", 24)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 6) 查询评论分组（优先 user gRPC，失败回退本地）。
	comments, err := s.loadCommentGroups(ctx, current.ID)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}

	// 7) 构建年份与期号切换数据。
	years := make([]int, 0)
	// 定义并初始化当前变量。
	yearSet := map[int]struct{}{}
	// 定义并初始化当前变量。
	issues := make([]map[string]interface{}, 0, len(list))
	// 循环处理当前数据集合。
	for _, row := range list {
		// 判断条件并进入对应分支逻辑。
		if _, ok := yearSet[row.Year]; !ok {
			// 更新当前变量或字段值。
			yearSet[row.Year] = struct{}{}
			// 更新当前变量或字段值。
			years = append(years, row.Year)
		}
		// 更新当前变量或字段值。
		issues = append(issues, map[string]interface{}{"id": row.ID, "issue": row.Issue, "year": row.Year})
	}
	// 年份按降序排序，确保最新年份在前。
	sortYearsDesc(years)

	// 8) 开奖区数据优先读取开奖记录表；缺失时再回退图纸兼容字段。
	drawNumbers := extractDrawNumbers(*current)
	// 定义并初始化当前变量。
	drawLabels := buildSimpleLabels(drawNumbers)
	// 定义并初始化当前变量。
	drawIssue := current.Issue
	// 定义并初始化当前变量。
	drawYear := current.Year
	// 定义并初始化当前变量。
	drawPlaybackURL := current.PlaybackURL
	// 定义并初始化当前变量。
	drawRecordID := uint(0)
	// 判断条件并进入对应分支逻辑。
	if record, recErr := s.dao.GetLatestDrawRecordBySpecialID(current.SpecialLotteryID); recErr == nil {
		// 更新当前变量或字段值。
		drawNumbers = extractDrawNumbersFromRecord(*record)
		// 更新当前变量或字段值。
		drawLabels = extractDrawLabels(*record, drawNumbers)
		// 更新当前变量或字段值。
		drawIssue = record.Issue
		// 更新当前变量或字段值。
		drawYear = record.Year
		// 更新当前变量或字段值。
		drawPlaybackURL = record.PlaybackURL
		// 更新当前变量或字段值。
		drawRecordID = record.ID
	}

	// 9) 计算投票总数与百分比。
	totalVotes := int64(0)
	// 循环处理当前数据集合。
	for _, opt := range options {
		// 更新当前变量或字段值。
		totalVotes += opt.Votes
	}
	// 定义并初始化当前变量。
	poll := buildPollPayload(options, totalVotes)
	// 定义并初始化当前变量。
	pollEnabled := current.PollEnabled == 1 || len(options) > 0

	// 10) 转换 banner 到前端结构。
	bannerPayload := make([]map[string]interface{}, 0, len(banners))
	// 循环处理当前数据集合。
	for _, b := range banners {
		// 更新当前变量或字段值。
		bannerPayload = append(bannerPayload, map[string]interface{}{
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
		})
	}

	// 11) 优先读取后台配置的推荐图纸 IDs。
	recommendInfos := make([]models.WLotteryInfo, 0)
	// 定义并初始化当前变量。
	recommendIDs := parseCSVUintIDs(current.RecommendInfoIDs)
	// 判断条件并进入对应分支逻辑。
	if len(recommendIDs) > 0 {
		// 更新当前变量或字段值。
		recommendInfos, err = s.dao.ListLotteryInfosByIDs(recommendIDs)
		// 判断条件并进入对应分支逻辑。
		if err != nil {
			// 返回当前处理结果。
			return nil, err
		}
		// 更新当前变量或字段值。
		recommendInfos = reorderInfosByIDs(recommendInfos, recommendIDs, current.ID)
	}
	// 12) 若未配置推荐，则回退为同彩种最新图纸。
	if len(recommendInfos) == 0 {
		// 更新当前变量或字段值。
		recommendInfos, err = s.dao.ListRecommendFallback(current.SpecialLotteryID, current.ID, 8)
		// 判断条件并进入对应分支逻辑。
		if err != nil {
			// 返回当前处理结果。
			return nil, err
		}
	}

	// 13) 指标栏按“有数据才显示”策略控制显隐。
	showMetrics := current.LikesCount > 0 || current.CommentCount > 0 || current.FavoriteCount > 0 || current.ReadCount > 0
	// 返回当前处理结果。
	return map[string]interface{}{
		// 进入新的代码块进行处理。
		"current": map[string]interface{}{
			// 处理当前语句逻辑。
			"id": current.ID,
			// 处理当前语句逻辑。
			"special_lottery_id": current.SpecialLotteryID,
			// 处理当前语句逻辑。
			"title": current.Title,
			// 处理当前语句逻辑。
			"issue": current.Issue,
			// 处理当前语句逻辑。
			"year": current.Year,
			// 处理当前语句逻辑。
			"draw_issue": drawIssue,
			// 处理当前语句逻辑。
			"draw_year": drawYear,
			// 处理当前语句逻辑。
			"draw_record_id": drawRecordID,
			// 处理当前语句逻辑。
			"detail_image_url": current.DetailImageURL,
			// 处理当前语句逻辑。
			"draw_code": current.DrawCode,
			// 处理当前语句逻辑。
			"normal_draw_result": current.NormalDrawResult,
			// 处理当前语句逻辑。
			"special_draw_result": current.SpecialDrawResult,
			// 处理当前语句逻辑。
			"draw_result": current.DrawResult,
			// 处理当前语句逻辑。
			"draw_numbers": drawNumbers,
			// 处理当前语句逻辑。
			"draw_labels": drawLabels,
			// 处理当前语句逻辑。
			"playback_url": drawPlaybackURL,
			// 处理当前语句逻辑。
			"likes_count": current.LikesCount,
			// 处理当前语句逻辑。
			"comment_count": current.CommentCount,
			// 处理当前语句逻辑。
			"favorite_count": current.FavoriteCount,
			// 处理当前语句逻辑。
			"read_count": current.ReadCount,
		},
		// 处理当前语句逻辑。
		"years": years,
		// 处理当前语句逻辑。
		"issues": issues,
		// 处理当前语句逻辑。
		"poll_options": poll,
		// 处理当前语句逻辑。
		"poll_enabled": pollEnabled,
		// 处理当前语句逻辑。
		"poll_default_open": current.PollDefaultExpand == 1,
		// 处理当前语句逻辑。
		"show_metrics": showMetrics,
		// 处理当前语句逻辑。
		"detail_banners": bannerPayload,
		// 调用buildRecommendPayload完成当前处理。
		"recommend_items": buildRecommendPayload(recommendInfos),
		// 调用buildExternalLinkPayload完成当前处理。
		"external_links": buildExternalLinkPayload(links),
		// 处理当前语句逻辑。
		"system_comments": comments.SystemComments,
		// 处理当前语句逻辑。
		"user_comments": comments.UserComments,
		// 处理当前语句逻辑。
		"hot_comments": comments.HotComments,
		// 处理当前语句逻辑。
		"latest_comments": comments.LatestComments,
		// 处理当前语句逻辑。
	}, nil
}
