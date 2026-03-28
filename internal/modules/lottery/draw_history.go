package lottery

import (
	"fmt"
	"strings"
)

// BuildDrawHistoryBySpecialID 生成开奖区“历史开奖”列表数据。
func (s *Service) BuildDrawHistoryBySpecialID(specialLotteryID uint, orderMode string, showFive bool, limit int) (map[string]interface{}, error) {
	// 1) 查询彩种基础信息，确保返回结构包含彩种名。
	sl, err := s.dao.GetSpecialLottery(specialLotteryID)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 判断条件并进入对应分支逻辑。
	if limit <= 0 || limit > 200 {
		// 更新当前变量或字段值。
		limit = 80
	}
	// 定义并初始化当前变量。
	order := normalizeOrderMode(orderMode)
	if sl == nil {
		return emptyDrawHistoryPayload(specialLotteryID, "", order, showFive), nil
	}

	// 2) 拉取开奖记录列表。
	rows, err := s.dao.ListDrawRecordsBySpecialID(specialLotteryID, limit, order)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}

	// 3) 转换前端展示结构。
	items := make([]map[string]interface{}, 0, len(rows))
	// 定义并初始化当前变量。
	historyYear := lotteryNowInEast8().Year()
	// 判断条件并进入对应分支逻辑。
	if len(rows) > 0 {
		// 更新当前变量或字段值。
		historyYear = rows[0].Year
	}
	// 循环处理当前数据集合。
	for _, row := range rows {
		// 定义并初始化当前变量。
		numbers := extractDrawNumbersFromRecord(row)
		// 定义并初始化当前变量。
		stats := deriveRecordStats(numbers)
		// 定义并初始化当前变量。
		pairLabels := extractDrawLabels(row, numbers)
		// 定义并初始化当前变量。
		colorLabels := extractColorLabels(row, numbers)
		// 定义并初始化当前变量。
		zodiacLabels, wuxingLabels := extractZodiacAndWuxingLabels(row, numbers)
		// 定义并初始化当前变量。
		displayLabels := zodiacLabels
		// 判断条件并进入对应分支逻辑。
		if showFive {
			// 更新当前变量或字段值。
			displayLabels = pairLabels
		}
		// 定义并初始化当前变量。
		item := map[string]interface{}{
			// 处理当前语句逻辑。
			"id": row.ID,
			// 处理当前语句逻辑。
			"special_lottery_id": row.SpecialLotteryID,
			// 处理当前语句逻辑。
			"special_lottery_name": sl.Name,
			// 处理当前语句逻辑。
			"issue": row.Issue,
			// 处理当前语句逻辑。
			"year": row.Year,
			// 调用row.DrawAt.Format完成当前处理。
			"draw_at": row.DrawAt.Format("2006-01-02"),
			// 调用row.DrawAt.Format完成当前处理。
			"draw_time": row.DrawAt.Format("2006-01-02 15:04:05"),
			// 处理当前语句逻辑。
			"normal_draw_result": row.NormalDrawResult,
			// 处理当前语句逻辑。
			"special_draw_result": row.SpecialDrawResult,
			// 处理当前语句逻辑。
			"draw_result": row.DrawResult,
			// 处理当前语句逻辑。
			"numbers": numbers,
			// 处理当前语句逻辑。
			"labels": displayLabels,
			// 处理当前语句逻辑。
			"pair_labels": pairLabels,
			// 处理当前语句逻辑。
			"color_labels": colorLabels,
			// 处理当前语句逻辑。
			"zodiac_labels": zodiacLabels,
			// 处理当前语句逻辑。
			"wuxing_labels": wuxingLabels,
			// 处理当前语句逻辑。
			"playback_url": row.PlaybackURL,
			// 处理当前语句逻辑。
			"special_single_double": pickString(row.SpecialSingleDouble, stats.SpecialSingleDouble),
			// 处理当前语句逻辑。
			"special_big_small": pickString(row.SpecialBigSmall, stats.SpecialBigSmall),
			// 处理当前语句逻辑。
			"sum_single_double": pickString(row.SumSingleDouble, stats.SumSingleDouble),
			// 处理当前语句逻辑。
			"sum_big_small": pickString(row.SumBigSmall, stats.SumBigSmall),
			// 处理当前语句逻辑。
			"special_code": pickString(row.SpecialCode, stats.SpecialCode),
			// 处理当前语句逻辑。
			"normal_code": pickString(row.NormalCode, stats.NormalCode),
			// 处理当前语句逻辑。
			"zheng1": pickString(row.Zheng1, stats.ZhengDescriptions[0]),
			// 处理当前语句逻辑。
			"zheng2": pickString(row.Zheng2, stats.ZhengDescriptions[1]),
			// 处理当前语句逻辑。
			"zheng3": pickString(row.Zheng3, stats.ZhengDescriptions[2]),
			// 处理当前语句逻辑。
			"zheng4": pickString(row.Zheng4, stats.ZhengDescriptions[3]),
			// 处理当前语句逻辑。
			"zheng5": pickString(row.Zheng5, stats.ZhengDescriptions[4]),
			// 处理当前语句逻辑。
			"zheng6": pickString(row.Zheng6, stats.ZhengDescriptions[5]),
		}
		// 更新当前变量或字段值。
		items = append(items, item)
	}

	// 4) 返回历史开奖页聚合结构。
	return map[string]interface{}{
		// 处理当前语句逻辑。
		"special_lottery_id": specialLotteryID,
		// 处理当前语句逻辑。
		"special_lottery_name": sl.Name,
		// 处理当前语句逻辑。
		"year": historyYear,
		// 处理当前语句逻辑。
		"order_mode": order,
		// 处理当前语句逻辑。
		"show_five": showFive,
		// 处理当前语句逻辑。
		"items": items,
		// 处理当前语句逻辑。
	}, nil
}

// BuildDrawDetail 生成开奖区“开奖详情”页数据（图5结构）。
func (s *Service) BuildDrawDetail(recordID uint) (map[string]interface{}, error) {
	// 1) 查询目标开奖记录。
	row, err := s.dao.GetDrawRecord(recordID)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	if row == nil {
		return emptyDrawDetailPayload(recordID), nil
	}
	// 定义并初始化当前变量。
	sl, err := s.dao.GetSpecialLottery(row.SpecialLotteryID)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	specialLotteryName := ""
	if sl != nil {
		specialLotteryName = sl.Name
	}

	// 2) 提取号码与标签。
	numbers := extractDrawNumbersFromRecord(*row)
	// 定义并初始化当前变量。
	labels := extractDrawLabels(*row, numbers)
	// 定义并初始化当前变量。
	colorLabels := extractColorLabels(*row, numbers)
	// 定义并初始化当前变量。
	zodiacLabels, wuxingLabels := extractZodiacAndWuxingLabels(*row, numbers)
	// 定义并初始化当前变量。
	resultBundle := buildDrawResultBundleView(numbers)

	// 3) 自动补齐详情字段（后台没填时兜底）。
	stats := deriveRecordStats(numbers)
	// 定义并初始化当前变量。
	specialSingleDouble := pickString(row.SpecialSingleDouble, stats.SpecialSingleDouble)
	// 定义并初始化当前变量。
	specialBigSmall := pickString(row.SpecialBigSmall, stats.SpecialBigSmall)
	// 定义并初始化当前变量。
	sumSingleDouble := pickString(row.SumSingleDouble, stats.SumSingleDouble)
	// 定义并初始化当前变量。
	sumBigSmall := pickString(row.SumBigSmall, stats.SumBigSmall)
	// 定义并初始化当前变量。
	specialCode := pickString(row.SpecialCode, stats.SpecialCode)
	// 定义并初始化当前变量。
	normalCode := pickString(row.NormalCode, stats.NormalCode)
	// 定义并初始化当前变量。
	zheng1 := pickString(row.Zheng1, stats.ZhengDescriptions[0])
	// 定义并初始化当前变量。
	zheng2 := pickString(row.Zheng2, stats.ZhengDescriptions[1])
	// 定义并初始化当前变量。
	zheng3 := pickString(row.Zheng3, stats.ZhengDescriptions[2])
	// 定义并初始化当前变量。
	zheng4 := pickString(row.Zheng4, stats.ZhengDescriptions[3])
	// 定义并初始化当前变量。
	zheng5 := pickString(row.Zheng5, stats.ZhengDescriptions[4])
	// 定义并初始化当前变量。
	zheng6 := pickString(row.Zheng6, stats.ZhengDescriptions[5])

	// 4) 返回开奖详情结构。
	return map[string]interface{}{
		// 处理当前语句逻辑。
		"id": row.ID,
		// 处理当前语句逻辑。
		"special_lottery_id": row.SpecialLotteryID,
		// 处理当前语句逻辑。
		"special_lottery_name": specialLotteryName,
		// 处理当前语句逻辑。
		"issue": row.Issue,
		// 处理当前语句逻辑。
		"year": row.Year,
		// 调用row.DrawAt.Format完成当前处理。
		"draw_at": row.DrawAt.Format("2006-01-02"),
		// 调用row.DrawAt.Format完成当前处理。
		"draw_time": row.DrawAt.Format("2006-01-02 15:04:05"),
		// 处理当前语句逻辑。
		"normal_draw_result": row.NormalDrawResult,
		// 处理当前语句逻辑。
		"special_draw_result": row.SpecialDrawResult,
		// 处理当前语句逻辑。
		"draw_result": row.DrawResult,
		// 处理当前语句逻辑。
		"numbers": numbers,
		// 处理当前语句逻辑。
		"labels": labels,
		// 处理当前语句逻辑。
		"pair_labels": labels,
		// 处理当前语句逻辑。
		"color_labels": colorLabels,
		// 处理当前语句逻辑。
		"zodiac_labels": zodiacLabels,
		// 处理当前语句逻辑。
		"wuxing_labels": wuxingLabels,
		// 处理当前语句逻辑。
		"playback_url": row.PlaybackURL,
		// 处理当前语句逻辑。
		"special_single_double": specialSingleDouble,
		// 处理当前语句逻辑。
		"special_big_small": specialBigSmall,
		// 处理当前语句逻辑。
		"sum_single_double": sumSingleDouble,
		// 处理当前语句逻辑。
		"sum_big_small": sumBigSmall,
		// 处理当前语句逻辑。
		"recommend_six": row.RecommendSix,
		// 处理当前语句逻辑。
		"recommend_four": row.RecommendFour,
		// 处理当前语句逻辑。
		"recommend_one": row.RecommendOne,
		// 处理当前语句逻辑。
		"recommend_ten": row.RecommendTen,
		// 处理当前语句逻辑。
		"special_code": specialCode,
		// 处理当前语句逻辑。
		"normal_code": normalCode,
		// 处理当前语句逻辑。
		"zheng1": zheng1,
		// 处理当前语句逻辑。
		"zheng2": zheng2,
		// 处理当前语句逻辑。
		"zheng3": zheng3,
		// 处理当前语句逻辑。
		"zheng4": zheng4,
		// 处理当前语句逻辑。
		"zheng5": zheng5,
		// 处理当前语句逻辑。
		"zheng6": zheng6,
		// 处理当前语句逻辑。
		"result_bundle": resultBundle,
		// 处理当前语句逻辑。
	}, nil
}

// normalizeOrderMode 标准化历史排序参数。
func normalizeOrderMode(raw string) string {
	// 定义并初始化当前变量。
	mode := strings.ToLower(strings.TrimSpace(raw))
	// 判断条件并进入对应分支逻辑。
	if mode == "asc" {
		return "asc"
	}
	return "desc"
}

// recordStats 开奖结果自动计算结构。
type recordStats struct {
	// 处理当前语句逻辑。
	SpecialSingleDouble string
	// 处理当前语句逻辑。
	SpecialBigSmall string
	// 处理当前语句逻辑。
	SumSingleDouble string
	// 处理当前语句逻辑。
	SumBigSmall string
	// 处理当前语句逻辑。
	SpecialCode string
	// 处理当前语句逻辑。
	NormalCode string
	// 处理当前语句逻辑。
	ZhengDescriptions [6]string
}

// deriveRecordStats 基于开奖号码自动计算详情字段兜底值。
func deriveRecordStats(numbers []int) recordStats {
	// 判断条件并进入对应分支逻辑。
	if len(numbers) < 7 {
		// 返回当前处理结果。
		return recordStats{}
	}
	// 定义并初始化当前变量。
	total := 0
	// 定义并初始化当前变量。
	zhengDescriptions := [6]string{}
	// 循环处理当前数据集合。
	for idx, n := range numbers {
		// 更新当前变量或字段值。
		total += n
		// 前 6 个位置按正码玩法规则输出兜底描述。
		if idx < 6 {
			zhengDescriptions[idx] = composeZhengDescriptionView(compileNumberDetailView(n, idx+1))
		}
	}
	// 定义并初始化当前变量。
	special := numbers[6]
	// 返回当前处理结果。
	return recordStats{
		// 严格按特码规则处理 49 和局。
		SpecialSingleDouble: resultSpecialSingleDouble(special),
		// 严格按特码规则处理 49 和局。
		SpecialBigSmall: resultSpecialBigSmall(special),
		// 总和单双按总分奇偶。
		SumSingleDouble: totalSingleDoubleCN(total),
		// 总和大小按 >=175 为大。
		SumBigSmall: totalBigSmallCN(total),
		// 调用fmt.Sprintf完成当前处理。
		SpecialCode: fmt.Sprintf("%d", special),
		// 与后台主表 normal_code 口径保持一致，使用逗号分隔原始数字。
		NormalCode: joinIntCSV(numbers[:6]),
		// 正码位置玩法描述兜底。
		ZhengDescriptions: zhengDescriptions,
	}
}

// totalSingleDoubleCN 输出总分单双。
func totalSingleDoubleCN(v int) string {
	// 判断条件并进入对应分支逻辑。
	if v%2 == 0 {
		return "双"
	}
	return "单"
}

// totalBigSmallCN 输出总分大小（>=175 为大）。
func totalBigSmallCN(v int) string {
	// 判断条件并进入对应分支逻辑。
	if v >= 175 {
		return "大"
	}
	return "小"
}

// joinIntCSV 按逗号拼接原始数字串。
func joinIntCSV(nums []int) string {
	parts := make([]string, 0, len(nums))
	for _, num := range nums {
		parts = append(parts, fmt.Sprintf("%d", num))
	}
	return strings.Join(parts, ",")
}

// pickString 优先使用手工配置值；为空时使用自动值。
func pickString(manual, auto string) string {
	// 判断条件并进入对应分支逻辑。
	if strings.TrimSpace(manual) != "" {
		// 返回当前处理结果。
		return strings.TrimSpace(manual)
	}
	// 返回当前处理结果。
	return strings.TrimSpace(auto)
}
