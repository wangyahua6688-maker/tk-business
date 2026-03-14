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
	if err != nil {
		return nil, err
	}

	// 2) 查询同彩种下的历史期号，用于详情页切换。
	list, err := s.dao.ListLotteryInfosBySpecialID(current.SpecialLotteryID, 30)
	if err != nil {
		return nil, err
	}
	// 3) 查询投票选项（若缺失则自动补默认选项）。
	options, err := s.ensurePollOptions(current.ID)
	if err != nil {
		return nil, err
	}
	// 4) 查询详情页 banner。
	banners, err := s.dao.ListDetailBanners()
	if err != nil {
		return nil, err
	}
	// 5) 查询详情页外链推荐区。
	links, err := s.dao.ListExternalLinks("lottery_detail", 24)
	if err != nil {
		return nil, err
	}
	// 6) 查询评论分组（优先 user gRPC，失败回退本地）。
	comments, err := s.loadCommentGroups(ctx, current.ID)
	if err != nil {
		return nil, err
	}

	// 7) 构建年份与期号切换数据。
	years := make([]int, 0)
	yearSet := map[int]struct{}{}
	issues := make([]map[string]interface{}, 0, len(list))
	for _, row := range list {
		if _, ok := yearSet[row.Year]; !ok {
			yearSet[row.Year] = struct{}{}
			years = append(years, row.Year)
		}
		issues = append(issues, map[string]interface{}{"id": row.ID, "issue": row.Issue, "year": row.Year})
	}
	// 年份按降序排序，确保最新年份在前。
	sortYearsDesc(years)

	// 8) 开奖区数据优先读取开奖记录表；缺失时再回退图纸兼容字段。
	drawNumbers := extractDrawNumbers(*current)
	drawLabels := buildSimpleLabels(drawNumbers)
	drawIssue := current.Issue
	drawYear := current.Year
	drawPlaybackURL := current.PlaybackURL
	drawRecordID := uint(0)
	if record, recErr := s.dao.GetLatestDrawRecordBySpecialID(current.SpecialLotteryID); recErr == nil {
		drawNumbers = extractDrawNumbersFromRecord(*record)
		drawLabels = extractDrawLabels(*record, drawNumbers)
		drawIssue = record.Issue
		drawYear = record.Year
		drawPlaybackURL = record.PlaybackURL
		drawRecordID = record.ID
	}

	// 9) 计算投票总数与百分比。
	totalVotes := int64(0)
	for _, opt := range options {
		totalVotes += opt.Votes
	}
	poll := buildPollPayload(options, totalVotes)
	pollEnabled := current.PollEnabled == 1 || len(options) > 0

	// 10) 转换 banner 到前端结构。
	bannerPayload := make([]map[string]interface{}, 0, len(banners))
	for _, b := range banners {
		bannerPayload = append(bannerPayload, map[string]interface{}{
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
		})
	}

	// 11) 优先读取后台配置的推荐图纸 IDs。
	recommendInfos := make([]models.WLotteryInfo, 0)
	recommendIDs := parseCSVUintIDs(current.RecommendInfoIDs)
	if len(recommendIDs) > 0 {
		recommendInfos, err = s.dao.ListLotteryInfosByIDs(recommendIDs)
		if err != nil {
			return nil, err
		}
		recommendInfos = reorderInfosByIDs(recommendInfos, recommendIDs, current.ID)
	}
	// 12) 若未配置推荐，则回退为同彩种最新图纸。
	if len(recommendInfos) == 0 {
		recommendInfos, err = s.dao.ListRecommendFallback(current.SpecialLotteryID, current.ID, 8)
		if err != nil {
			return nil, err
		}
	}

	// 13) 指标栏按“有数据才显示”策略控制显隐。
	showMetrics := current.LikesCount > 0 || current.CommentCount > 0 || current.FavoriteCount > 0 || current.ReadCount > 0
	return map[string]interface{}{
		"current": map[string]interface{}{
			"id":                  current.ID,
			"special_lottery_id":  current.SpecialLotteryID,
			"title":               current.Title,
			"issue":               current.Issue,
			"year":                current.Year,
			"draw_issue":          drawIssue,
			"draw_year":           drawYear,
			"draw_record_id":      drawRecordID,
			"detail_image_url":    current.DetailImageURL,
			"draw_code":           current.DrawCode,
			"normal_draw_result":  current.NormalDrawResult,
			"special_draw_result": current.SpecialDrawResult,
			"draw_result":         current.DrawResult,
			"draw_numbers":        drawNumbers,
			"draw_labels":         drawLabels,
			"playback_url":        drawPlaybackURL,
			"likes_count":         current.LikesCount,
			"comment_count":       current.CommentCount,
			"favorite_count":      current.FavoriteCount,
			"read_count":          current.ReadCount,
		},
		"years":             years,
		"issues":            issues,
		"poll_options":      poll,
		"poll_enabled":      pollEnabled,
		"poll_default_open": current.PollDefaultExpand == 1,
		"show_metrics":      showMetrics,
		"detail_banners":    bannerPayload,
		"recommend_items":   buildRecommendPayload(recommendInfos),
		"external_links":    buildExternalLinkPayload(links),
		"system_comments":   comments.SystemComments,
		"user_comments":     comments.UserComments,
		"hot_comments":      comments.HotComments,
		"latest_comments":   comments.LatestComments,
	}, nil
}
