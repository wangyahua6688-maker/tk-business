package lottery

import (
	"context"
	"fmt"
	"strings"
	"time"

	"tk-shared/models"
)

// BuildLiveScenePage 构建“开奖现场”整页数据。
// 说明：
// 1. 一个接口返回 LiveScenePage 所需核心区块；
// 2. 页面切换彩种时只需重新请求本接口；
// 3. 接口内部复用已有开奖看板、历史开奖与图卡能力。
func (s *Service) BuildLiveScenePage(specialLotteryID uint) (map[string]interface{}, error) {
	// 0) 优先读取 Redis 缓存，命中后直接返回整页数据。
	ctx := context.Background()
	if cached, ok := s.loadLiveSceneCache(ctx, specialLotteryID); ok {
		return cached, nil
	}

	// 1) 查询彩种标签，作为页面顶部切换按钮。
	tabs, err := s.dao.ListSpecialLotteries(12)
	if err != nil {
		return nil, err
	}
	if len(tabs) == 0 {
		return nil, fmt.Errorf("special lottery not found")
	}

	// 2) 选择当前激活彩种：优先使用请求参数，缺省回退第一项。
	activeID := specialLotteryID
	if activeID == 0 {
		activeID = tabs[0].ID
	}
	if !containsSpecialLotteryID(tabs, activeID) {
		activeID = tabs[0].ID
	}

	// 3) 构建开奖看板（期号、倒计时、直播状态、开奖号码）。
	dashboard, err := s.BuildDashboard(activeID)
	if err != nil {
		return nil, err
	}

	// 4) 拉取图卡并过滤出当前彩种，用于“开奖回放”与“推荐”计算。
	allCards, err := s.ListCards("")
	if err != nil {
		return nil, err
	}
	sceneCards := filterCardsBySpecialLottery(allCards, activeID)
	playbackItems := sceneCards
	if len(playbackItems) > 4 {
		playbackItems = playbackItems[:4]
	}

	// 5) 拉取当前彩种历史开奖（主数据来自 tk_draw_record）。
	historyItems := make([]map[string]interface{}, 0)
	history, hisErr := s.BuildDrawHistoryBySpecialID(activeID, "desc", true, 80)
	if hisErr != nil {
		return nil, hisErr
	}
	historyItems = extractHistoryItems(history)

	// 6) 直接返回推荐区块，前端无需再次拼接。
	recommendBlocks := buildSceneRecommendBlocks(historyItems, sceneCards)

	// 7) 组装整页数据，前端只调用一次即可渲染。
	payload := map[string]interface{}{
		"scene_title":      "开奖现场",
		"generated_at":     time.Now().Format(time.RFC3339),
		"tabs":             buildSceneTabsPayload(tabs),
		"active_tab_id":    activeID,
		"dashboard":        dashboard,
		"cards":            sceneCards,
		"playback_items":   playbackItems,
		"history_items":    historyItems,
		"recommend_blocks": recommendBlocks,
	}

	// 8) 写入缓存：同时缓存“请求 ID”与“实际激活 ID”两套键，提升命中率。
	s.saveLiveSceneCache(ctx, specialLotteryID, payload)
	if activeID != specialLotteryID {
		s.saveLiveSceneCache(ctx, activeID, payload)
	}
	return payload, nil
}

// containsSpecialLotteryID 判断彩种 ID 是否存在于切换标签内。
func containsSpecialLotteryID(rows []models.WSpecialLottery, id uint) bool {
	for _, row := range rows {
		if row.ID == id {
			return true
		}
	}
	return false
}

// buildSceneTabsPayload 构建前端切换标签结构。
func buildSceneTabsPayload(rows []models.WSpecialLottery) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]interface{}{
			"id":            row.ID,
			"name":          row.Name,
			"code":          row.Code,
			"current_issue": row.CurrentIssue,
			"next_draw_at":  row.NextDrawAt.Format(time.RFC3339),
		})
	}
	return out
}

// filterCardsBySpecialLottery 过滤当前彩种图卡列表。
func filterCardsBySpecialLottery(cards []map[string]interface{}, specialLotteryID uint) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(cards))
	for _, row := range cards {
		if toUint(row["special_lottery_id"]) == specialLotteryID {
			out = append(out, row)
		}
	}
	return out
}

// extractHistoryItems 从历史开奖 payload 中提取 items。
func extractHistoryItems(history map[string]interface{}) []map[string]interface{} {
	items, ok := history["items"].([]map[string]interface{})
	if ok {
		return items
	}
	// 兼容兜底：类型断言失败时返回空列表，避免前端崩溃。
	return []map[string]interface{}{}
}

// buildSceneRecommendBlocks 构建“推荐区”展示数据。
func buildSceneRecommendBlocks(historyItems []map[string]interface{}, cards []map[string]interface{}) []map[string]interface{} {
	// 1) 建立 issue -> draw_code 映射，便于补充“十码”数据。
	drawCodeByIssue := make(map[string]string, len(cards))
	for _, item := range cards {
		issue := strings.TrimSpace(toString(item["issue"]))
		if issue == "" {
			continue
		}
		drawCodeByIssue[issue] = toString(item["draw_code"])
	}

	// 2) 将历史开奖前三期转换为推荐区块。
	limit := 3
	if len(historyItems) < limit {
		limit = len(historyItems)
	}
	blocks := make([]map[string]interface{}, 0, limit)
	for idx := 0; idx < limit; idx++ {
		item := historyItems[idx]
		issue := strings.TrimSpace(toString(item["issue"]))
		numbers := toIntSlice(item["numbers"])
		labels := toStringSlice(item["labels"])
		zodiac := pickZodiac(labels)
		drawCode := splitCSVInts(drawCodeByIssue[issue])
		ten := uniqueNumbers(append(numbers, drawCode...))
		if len(ten) > 10 {
			ten = ten[:10]
		}
		yearPrefix, issueNo := splitIssue(issue)
		title := fmt.Sprintf("%s 第%s期推荐", yearPrefix, issueNo)

		blocks = append(blocks, map[string]interface{}{
			"id":    toUint(item["id"]),
			"title": title,
			"lines": []map[string]interface{}{
				{"label": "六肖", "value": strings.Join(sliceString(zodiac, 6), " "), "hit": idx >= 1, "highlight": false},
				{"label": "四肖", "value": strings.Join(sliceString(zodiac, 4), " "), "hit": idx == 1, "highlight": false},
				{"label": "一肖", "value": strings.Join(sliceString(zodiac, 1), " "), "hit": false, "highlight": false},
				{"label": "十码", "value": joinPaddedInts(ten), "hit": false, "highlight": true},
			},
		})
	}
	return blocks
}

// pickZodiac 从“生肖/五行”标签中提取生肖片段。
func pickZodiac(labels []string) []string {
	out := make([]string, 0, len(labels))
	for _, label := range labels {
		parts := strings.Split(strings.TrimSpace(label), "/")
		head := strings.TrimSpace(parts[0])
		if head != "" {
			out = append(out, head)
		}
	}
	return out
}

// splitIssue 将 "2026-024" 拆为年和期号。
func splitIssue(issue string) (string, string) {
	raw := strings.TrimSpace(issue)
	if raw == "" {
		return "", ""
	}
	parts := strings.Split(raw, "-")
	if len(parts) < 2 {
		return raw, raw
	}
	return parts[0], parts[1]
}

// joinPaddedInts 将数字按两位补零并拼接为字符串。
func joinPaddedInts(nums []int) string {
	if len(nums) == 0 {
		return ""
	}
	out := make([]string, 0, len(nums))
	for _, n := range nums {
		out = append(out, fmt.Sprintf("%02d", n))
	}
	return strings.Join(out, " ")
}

// uniqueNumbers 去重并保持原顺序。
func uniqueNumbers(input []int) []int {
	seen := map[int]struct{}{}
	out := make([]int, 0, len(input))
	for _, n := range input {
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out
}

// toUint 将 interface{} 安全转换为 uint。
func toUint(v interface{}) uint {
	switch x := v.(type) {
	case uint:
		return x
	case uint64:
		return uint(x)
	case uint32:
		return uint(x)
	case int:
		if x > 0 {
			return uint(x)
		}
	case int64:
		if x > 0 {
			return uint(x)
		}
	case float64:
		if x > 0 {
			return uint(x)
		}
	}
	return 0
}

// toString 将 interface{} 转换为字符串。
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// toIntSlice 将 interface{} 转换为 []int。
func toIntSlice(v interface{}) []int {
	switch rows := v.(type) {
	case []int:
		return rows
	case []interface{}:
		out := make([]int, 0, len(rows))
		for _, item := range rows {
			switch n := item.(type) {
			case int:
				out = append(out, n)
			case int32:
				out = append(out, int(n))
			case int64:
				out = append(out, int(n))
			case float64:
				out = append(out, int(n))
			}
		}
		return out
	default:
		return []int{}
	}
}

// toStringSlice 将 interface{} 转换为 []string。
func toStringSlice(v interface{}) []string {
	switch rows := v.(type) {
	case []string:
		return rows
	case []interface{}:
		out := make([]string, 0, len(rows))
		for _, item := range rows {
			out = append(out, strings.TrimSpace(toString(item)))
		}
		return out
	default:
		return []string{}
	}
}

// sliceString 按上限截断字符串列表。
func sliceString(input []string, limit int) []string {
	if limit <= 0 {
		return []string{}
	}
	if len(input) <= limit {
		return input
	}
	return input[:limit]
}
