package lottery

import (
	"context"
	"fmt"
	"strings"
	"time"

	common_model "github.com/wangyahua6688-maker/tk-common/models"
)

// BuildLiveScenePage 构建“开奖现场”整页数据。
// 说明：
// 1. 一个接口返回 LiveScenePage 所需核心区块；
// 2. 页面切换彩种时只需重新请求本接口；
// 3. 接口内部复用已有开奖看板、历史开奖与图卡能力。
func (s *Service) BuildLiveScenePage(specialLotteryID uint) (map[string]interface{}, error) {
	now := lotteryNowInEast8()
	// 0) 优先读取 Redis 缓存，命中后直接返回整页数据。
	ctx := context.Background()
	// 判断条件并进入对应分支逻辑。
	if cached, ok := s.loadLiveSceneCache(ctx, specialLotteryID); ok {
		// 返回当前处理结果。
		return cached, nil
	}

	// 1) 查询彩种标签，作为页面顶部切换按钮。
	tabs, err := s.dao.ListSpecialLotteries(12)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 判断条件并进入对应分支逻辑。
	if len(tabs) == 0 {
		return emptyLiveScenePayload(now), nil
	}

	// 2) 选择当前激活彩种：优先使用请求参数，缺省回退第一项。
	activeID := specialLotteryID
	// 判断条件并进入对应分支逻辑。
	if activeID == 0 {
		// 更新当前变量或字段值。
		activeID = tabs[0].ID
	}
	// 判断条件并进入对应分支逻辑。
	if !containsSpecialLotteryID(tabs, activeID) {
		// 更新当前变量或字段值。
		activeID = tabs[0].ID
	}

	// 3) 构建开奖看板（期号、倒计时、直播状态、开奖号码）。
	dashboard, err := s.BuildDashboard(activeID)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}

	// 4) 拉取图卡并过滤出当前彩种，用于“开奖回放”与“推荐”计算。
	allCards, err := s.ListCards("")
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 定义并初始化当前变量。
	sceneCards := filterCardsBySpecialLottery(allCards, activeID)
	// 定义并初始化当前变量。
	playbackItems := sceneCards
	// 判断条件并进入对应分支逻辑。
	if len(playbackItems) > 4 {
		// 更新当前变量或字段值。
		playbackItems = playbackItems[:4]
	}

	// 5) 拉取当前彩种历史开奖（主数据来自 tk_draw_record）。
	history, hisErr := s.BuildDrawHistoryBySpecialID(activeID, "desc", true, 80)
	// 判断条件并进入对应分支逻辑。
	if hisErr != nil {
		// 返回当前处理结果。
		return nil, hisErr
	}
	// 定义并初始化当前变量。
	historyItems := extractHistoryItems(history)

	// 6) 直接返回推荐区块，前端无需再次拼接。
	recommendBlocks := buildSceneRecommendBlocks(historyItems, sceneCards)

	// 7) 组装整页数据，前端只调用一次即可渲染。
	payload := map[string]interface{}{
		// 处理当前语句逻辑。
		"scene_title": "开奖现场",
		// 调用time.Now完成当前处理。
		"generated_at": now.Format(time.RFC3339),
		// 调用buildSceneTabsPayload完成当前处理。
		"tabs": buildSceneTabsPayload(tabs, now),
		// 处理当前语句逻辑。
		"active_tab_id": activeID,
		// 处理当前语句逻辑。
		"dashboard": dashboard,
		// 处理当前语句逻辑。
		"cards": sceneCards,
		// 处理当前语句逻辑。
		"playback_items": playbackItems,
		// 处理当前语句逻辑。
		"history_items": historyItems,
		// 处理当前语句逻辑。
		"recommend_blocks": recommendBlocks,
	}

	// 8) 写入缓存：同时缓存“请求 ID”与“实际激活 ID”两套键，提升命中率。
	s.saveLiveSceneCache(ctx, specialLotteryID, payload)
	// 判断条件并进入对应分支逻辑。
	if activeID != specialLotteryID {
		// 调用s.saveLiveSceneCache完成当前处理。
		s.saveLiveSceneCache(ctx, activeID, payload)
	}
	// 返回当前处理结果。
	return payload, nil
}

// containsSpecialLotteryID 判断彩种 ID 是否存在于切换标签内。
func containsSpecialLotteryID(rows []common_model.WSpecialLottery, id uint) bool {
	// 循环处理当前数据集合。
	for _, row := range rows {
		// 判断条件并进入对应分支逻辑。
		if row.ID == id {
			// 返回当前处理结果。
			return true
		}
	}
	// 返回当前处理结果。
	return false
}

// buildSceneTabsPayload 构建前端切换标签结构。
func buildSceneTabsPayload(rows []common_model.WSpecialLottery, now time.Time) []map[string]interface{} {
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
			"code": row.Code,
			// 处理当前语句逻辑。
			"current_issue": row.CurrentIssue,
			// 统一按“每天固定开奖时刻”输出最近一次未来时间。
			"next_draw_at": normalizeNextDrawAt(row.NextDrawAt, now).Format(time.RFC3339),
		})
	}
	// 返回当前处理结果。
	return out
}

// filterCardsBySpecialLottery 过滤当前彩种图卡列表。
func filterCardsBySpecialLottery(cards []map[string]interface{}, specialLotteryID uint) []map[string]interface{} {
	// 定义并初始化当前变量。
	out := make([]map[string]interface{}, 0, len(cards))
	// 循环处理当前数据集合。
	for _, row := range cards {
		// 判断条件并进入对应分支逻辑。
		if toUint(row["special_lottery_id"]) == specialLotteryID {
			// 更新当前变量或字段值。
			out = append(out, row)
		}
	}
	// 返回当前处理结果。
	return out
}

// extractHistoryItems 从历史开奖 payload 中提取 items。
func extractHistoryItems(history map[string]interface{}) []map[string]interface{} {
	// 定义并初始化当前变量。
	items, ok := history["items"].([]map[string]interface{})
	// 判断条件并进入对应分支逻辑。
	if ok {
		// 返回当前处理结果。
		return items
	}
	// 兼容兜底：类型断言失败时返回空列表，避免前端崩溃。
	return []map[string]interface{}{}
}

// buildSceneRecommendBlocks 构建“推荐区”展示数据。
func buildSceneRecommendBlocks(historyItems []map[string]interface{}, cards []map[string]interface{}) []map[string]interface{} {
	// 1) 建立 issue -> draw_code 映射，便于补充“十码”数据。
	drawCodeByIssue := make(map[string]string, len(cards))
	// 循环处理当前数据集合。
	for _, item := range cards {
		// 定义并初始化当前变量。
		issue := strings.TrimSpace(toString(item["issue"]))
		// 判断条件并进入对应分支逻辑。
		if issue == "" {
			// 处理当前语句逻辑。
			continue
		}
		// 更新当前变量或字段值。
		drawCodeByIssue[issue] = toString(item["draw_code"])
	}

	// 2) 将历史开奖前三期转换为推荐区块。
	limit := 3
	// 判断条件并进入对应分支逻辑。
	if len(historyItems) < limit {
		// 更新当前变量或字段值。
		limit = len(historyItems)
	}
	// 定义并初始化当前变量。
	blocks := make([]map[string]interface{}, 0, limit)
	// 循环处理当前数据集合。
	for idx := 0; idx < limit; idx++ {
		// 定义并初始化当前变量。
		item := historyItems[idx]
		// 定义并初始化当前变量。
		issue := strings.TrimSpace(toString(item["issue"]))
		// 定义并初始化当前变量。
		numbers := toIntSlice(item["numbers"])
		// 定义并初始化当前变量。
		labels := toStringSlice(item["labels"])
		// 定义并初始化当前变量。
		zodiac := pickZodiac(labels)
		// 定义并初始化当前变量。
		drawCode := splitCSVInts(drawCodeByIssue[issue])
		// 定义并初始化当前变量。
		ten := uniqueNumbers(append(numbers, drawCode...))
		// 判断条件并进入对应分支逻辑。
		if len(ten) > 10 {
			// 更新当前变量或字段值。
			ten = ten[:10]
		}
		// 定义并初始化当前变量。
		yearPrefix, issueNo := splitIssue(issue)
		// 定义并初始化当前变量。
		title := fmt.Sprintf("%s 第%s期推荐", yearPrefix, issueNo)

		// 更新当前变量或字段值。
		blocks = append(blocks, map[string]interface{}{
			// 调用toUint完成当前处理。
			"id": toUint(item["id"]),
			// 处理当前语句逻辑。
			"title": title,
			// 进入新的代码块进行处理。
			"lines": []map[string]interface{}{
				// 调用strings.Join完成当前处理。
				{"label": "六肖", "value": strings.Join(sliceString(zodiac, 6), " "), "hit": idx >= 1, "highlight": false},
				// 调用strings.Join完成当前处理。
				{"label": "四肖", "value": strings.Join(sliceString(zodiac, 4), " "), "hit": idx == 1, "highlight": false},
				// 调用strings.Join完成当前处理。
				{"label": "一肖", "value": strings.Join(sliceString(zodiac, 1), " "), "hit": false, "highlight": false},
				// 调用joinPaddedInts完成当前处理。
				{"label": "十码", "value": joinPaddedInts(ten), "hit": false, "highlight": true},
			},
		})
	}
	// 返回当前处理结果。
	return blocks
}

// pickZodiac 从“生肖/五行”标签中提取生肖片段。
func pickZodiac(labels []string) []string {
	// 定义并初始化当前变量。
	out := make([]string, 0, len(labels))
	// 循环处理当前数据集合。
	for _, label := range labels {
		// 定义并初始化当前变量。
		parts := strings.Split(strings.TrimSpace(label), "/")
		// 定义并初始化当前变量。
		head := strings.TrimSpace(parts[0])
		// 判断条件并进入对应分支逻辑。
		if head != "" {
			// 更新当前变量或字段值。
			out = append(out, head)
		}
	}
	// 返回当前处理结果。
	return out
}

// splitIssue 将 "2026-024" 拆为年和期号。
func splitIssue(issue string) (string, string) {
	// 定义并初始化当前变量。
	raw := strings.TrimSpace(issue)
	// 判断条件并进入对应分支逻辑。
	if raw == "" {
		// 返回当前处理结果。
		return "", ""
	}
	// 定义并初始化当前变量。
	parts := strings.Split(raw, "-")
	// 判断条件并进入对应分支逻辑。
	if len(parts) < 2 {
		// 返回当前处理结果。
		return raw, raw
	}
	// 返回当前处理结果。
	return parts[0], parts[1]
}

// joinPaddedInts 将数字按两位补零并拼接为字符串。
func joinPaddedInts(nums []int) string {
	// 判断条件并进入对应分支逻辑。
	if len(nums) == 0 {
		// 返回当前处理结果。
		return ""
	}
	// 定义并初始化当前变量。
	out := make([]string, 0, len(nums))
	// 循环处理当前数据集合。
	for _, n := range nums {
		// 更新当前变量或字段值。
		out = append(out, fmt.Sprintf("%02d", n))
	}
	// 返回当前处理结果。
	return strings.Join(out, " ")
}

// uniqueNumbers 去重并保持原顺序。
func uniqueNumbers(input []int) []int {
	// 定义并初始化当前变量。
	seen := map[int]struct{}{}
	// 定义并初始化当前变量。
	out := make([]int, 0, len(input))
	// 循环处理当前数据集合。
	for _, n := range input {
		// 判断条件并进入对应分支逻辑。
		if _, ok := seen[n]; ok {
			// 处理当前语句逻辑。
			continue
		}
		// 更新当前变量或字段值。
		seen[n] = struct{}{}
		// 更新当前变量或字段值。
		out = append(out, n)
	}
	// 返回当前处理结果。
	return out
}

// toUint 将 interface{} 安全转换为 uint。
func toUint(v interface{}) uint {
	// 根据表达式进入多分支处理。
	switch x := v.(type) {
	case uint:
		// 返回当前处理结果。
		return x
	case uint64:
		// 返回当前处理结果。
		return uint(x)
	case uint32:
		// 返回当前处理结果。
		return uint(x)
	case int:
		// 判断条件并进入对应分支逻辑。
		if x > 0 {
			// 返回当前处理结果。
			return uint(x)
		}
	case int64:
		// 判断条件并进入对应分支逻辑。
		if x > 0 {
			// 返回当前处理结果。
			return uint(x)
		}
	case float64:
		// 判断条件并进入对应分支逻辑。
		if x > 0 {
			// 返回当前处理结果。
			return uint(x)
		}
	}
	// 返回当前处理结果。
	return 0
}

// toString 将 interface{} 转换为字符串。
func toString(v interface{}) string {
	// 判断条件并进入对应分支逻辑。
	if v == nil {
		// 返回当前处理结果。
		return ""
	}
	// 判断条件并进入对应分支逻辑。
	if s, ok := v.(string); ok {
		// 返回当前处理结果。
		return s
	}
	// 返回当前处理结果。
	return fmt.Sprintf("%v", v)
}

// toIntSlice 将 interface{} 转换为 []int。
func toIntSlice(v interface{}) []int {
	// 根据表达式进入多分支处理。
	switch rows := v.(type) {
	case []int:
		// 返回当前处理结果。
		return rows
	case []interface{}:
		// 定义并初始化当前变量。
		out := make([]int, 0, len(rows))
		// 循环处理当前数据集合。
		for _, item := range rows {
			// 根据表达式进入多分支处理。
			switch n := item.(type) {
			case int:
				// 更新当前变量或字段值。
				out = append(out, n)
			case int32:
				// 更新当前变量或字段值。
				out = append(out, int(n))
			case int64:
				// 更新当前变量或字段值。
				out = append(out, int(n))
			case float64:
				// 更新当前变量或字段值。
				out = append(out, int(n))
			}
		}
		// 返回当前处理结果。
		return out
	default:
		// 返回当前处理结果。
		return []int{}
	}
}

// toStringSlice 将 interface{} 转换为 []string。
func toStringSlice(v interface{}) []string {
	// 根据表达式进入多分支处理。
	switch rows := v.(type) {
	case []string:
		// 返回当前处理结果。
		return rows
	case []interface{}:
		// 定义并初始化当前变量。
		out := make([]string, 0, len(rows))
		// 循环处理当前数据集合。
		for _, item := range rows {
			// 更新当前变量或字段值。
			out = append(out, strings.TrimSpace(toString(item)))
		}
		// 返回当前处理结果。
		return out
	default:
		// 返回当前处理结果。
		return []string{}
	}
}

// sliceString 按上限截断字符串列表。
func sliceString(input []string, limit int) []string {
	// 判断条件并进入对应分支逻辑。
	if limit <= 0 {
		// 返回当前处理结果。
		return []string{}
	}
	// 判断条件并进入对应分支逻辑。
	if len(input) <= limit {
		// 返回当前处理结果。
		return input
	}
	// 返回当前处理结果。
	return input[:limit]
}
