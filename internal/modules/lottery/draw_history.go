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
		pairLabels := extractDrawLabels(row, numbers)
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
			"zodiac_labels": zodiacLabels,
		}
		// 判断条件并进入对应分支逻辑。
		if showFive {
			// 更新当前变量或字段值。
			item["wuxing_labels"] = wuxingLabels
			// 更新当前变量或字段值。
			item["pair_labels"] = pairLabels
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
	// 定义并初始化当前变量。
	sl, err := s.dao.GetSpecialLottery(row.SpecialLotteryID)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}

	// 2) 提取号码与标签。
	numbers := extractDrawNumbersFromRecord(*row)
	// 定义并初始化当前变量。
	labels := extractDrawLabels(*row, numbers)

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

	// 4) 返回开奖详情结构。
	return map[string]interface{}{
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
		"labels": labels,
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
		"zheng1": row.Zheng1,
		// 处理当前语句逻辑。
		"zheng2": row.Zheng2,
		// 处理当前语句逻辑。
		"zheng3": row.Zheng3,
		// 处理当前语句逻辑。
		"zheng4": row.Zheng4,
		// 处理当前语句逻辑。
		"zheng5": row.Zheng5,
		// 处理当前语句逻辑。
		"zheng6": row.Zheng6,
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
	// 循环处理当前数据集合。
	for _, n := range numbers {
		// 更新当前变量或字段值。
		total += n
	}
	// 定义并初始化当前变量。
	special := numbers[6]
	// 返回当前处理结果。
	return recordStats{
		// 调用oddEvenCN完成当前处理。
		SpecialSingleDouble: oddEvenCN(special),
		// 调用bigSmallCN完成当前处理。
		SpecialBigSmall: bigSmallCN(special, 24),
		// 调用oddEvenCN完成当前处理。
		SumSingleDouble: oddEvenCN(total),
		// 调用bigSmallCN完成当前处理。
		SumBigSmall: bigSmallCN(total, 175),
		// 调用fmt.Sprintf完成当前处理。
		SpecialCode: fmt.Sprintf("%02d", special),
		// 调用joinPaddedInts完成当前处理。
		NormalCode: joinPaddedInts(numbers[:6]),
	}
}

// oddEvenCN 输出中文单双。
func oddEvenCN(v int) string {
	// 判断条件并进入对应分支逻辑。
	if v%2 == 0 {
		return "双"
	}
	return "单"
}

// bigSmallCN 输出中文大小（> threshold 为大）。
func bigSmallCN(v, threshold int) string {
	// 判断条件并进入对应分支逻辑。
	if v > threshold {
		return "大"
	}
	return "小"
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
