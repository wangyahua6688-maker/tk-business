package lottery

import (
	"fmt"
	"strings"
	"time"
)

// BuildDrawHistoryBySpecialID 生成开奖区“历史开奖”列表数据。
func (s *Service) BuildDrawHistoryBySpecialID(specialLotteryID uint, orderMode string, showFive bool, limit int) (map[string]interface{}, error) {
	// 1) 查询彩种基础信息，确保返回结构包含彩种名。
	sl, err := s.dao.GetSpecialLottery(specialLotteryID)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 200 {
		limit = 80
	}
	order := normalizeOrderMode(orderMode)

	// 2) 拉取开奖记录列表。
	rows, err := s.dao.ListDrawRecordsBySpecialID(specialLotteryID, limit, order)
	if err != nil {
		return nil, err
	}

	// 3) 转换前端展示结构。
	items := make([]map[string]interface{}, 0, len(rows))
	historyYear := time.Now().Year()
	if len(rows) > 0 {
		historyYear = rows[0].Year
	}
	for _, row := range rows {
		numbers := extractDrawNumbersFromRecord(row)
		pairLabels := extractDrawLabels(row, numbers)
		zodiacLabels, wuxingLabels := extractZodiacAndWuxingLabels(row, numbers)
		displayLabels := zodiacLabels
		if showFive {
			displayLabels = pairLabels
		}
		item := map[string]interface{}{
			"id":                  row.ID,
			"issue":               row.Issue,
			"year":                row.Year,
			"draw_at":             row.DrawAt.Format("2006-01-02"),
			"draw_time":           row.DrawAt.Format("2006-01-02 15:04:05"),
			"normal_draw_result":  row.NormalDrawResult,
			"special_draw_result": row.SpecialDrawResult,
			"draw_result":         row.DrawResult,
			"numbers":             numbers,
			"labels":              displayLabels,
			"zodiac_labels":       zodiacLabels,
		}
		if showFive {
			item["wuxing_labels"] = wuxingLabels
			item["pair_labels"] = pairLabels
		}
		items = append(items, item)
	}

	// 4) 返回历史开奖页聚合结构。
	return map[string]interface{}{
		"special_lottery_id":   specialLotteryID,
		"special_lottery_name": sl.Name,
		"year":                 historyYear,
		"order_mode":           order,
		"show_five":            showFive,
		"items":                items,
	}, nil
}

// BuildDrawDetail 生成开奖区“开奖详情”页数据（图5结构）。
func (s *Service) BuildDrawDetail(recordID uint) (map[string]interface{}, error) {
	// 1) 查询目标开奖记录。
	row, err := s.dao.GetDrawRecord(recordID)
	if err != nil {
		return nil, err
	}
	sl, err := s.dao.GetSpecialLottery(row.SpecialLotteryID)
	if err != nil {
		return nil, err
	}

	// 2) 提取号码与标签。
	numbers := extractDrawNumbersFromRecord(*row)
	labels := extractDrawLabels(*row, numbers)

	// 3) 自动补齐详情字段（后台没填时兜底）。
	stats := deriveRecordStats(numbers)
	specialSingleDouble := pickString(row.SpecialSingleDouble, stats.SpecialSingleDouble)
	specialBigSmall := pickString(row.SpecialBigSmall, stats.SpecialBigSmall)
	sumSingleDouble := pickString(row.SumSingleDouble, stats.SumSingleDouble)
	sumBigSmall := pickString(row.SumBigSmall, stats.SumBigSmall)
	specialCode := pickString(row.SpecialCode, stats.SpecialCode)
	normalCode := pickString(row.NormalCode, stats.NormalCode)

	// 4) 返回开奖详情结构。
	return map[string]interface{}{
		"id":                    row.ID,
		"special_lottery_id":    row.SpecialLotteryID,
		"special_lottery_name":  sl.Name,
		"issue":                 row.Issue,
		"year":                  row.Year,
		"draw_at":               row.DrawAt.Format("2006-01-02"),
		"draw_time":             row.DrawAt.Format("2006-01-02 15:04:05"),
		"normal_draw_result":    row.NormalDrawResult,
		"special_draw_result":   row.SpecialDrawResult,
		"draw_result":           row.DrawResult,
		"numbers":               numbers,
		"labels":                labels,
		"playback_url":          row.PlaybackURL,
		"special_single_double": specialSingleDouble,
		"special_big_small":     specialBigSmall,
		"sum_single_double":     sumSingleDouble,
		"sum_big_small":         sumBigSmall,
		"recommend_six":         row.RecommendSix,
		"recommend_four":        row.RecommendFour,
		"recommend_one":         row.RecommendOne,
		"recommend_ten":         row.RecommendTen,
		"special_code":          specialCode,
		"normal_code":           normalCode,
		"zheng1":                row.Zheng1,
		"zheng2":                row.Zheng2,
		"zheng3":                row.Zheng3,
		"zheng4":                row.Zheng4,
		"zheng5":                row.Zheng5,
		"zheng6":                row.Zheng6,
	}, nil
}

// normalizeOrderMode 标准化历史排序参数。
func normalizeOrderMode(raw string) string {
	mode := strings.ToLower(strings.TrimSpace(raw))
	if mode == "asc" {
		return "asc"
	}
	return "desc"
}

// recordStats 开奖结果自动计算结构。
type recordStats struct {
	SpecialSingleDouble string
	SpecialBigSmall     string
	SumSingleDouble     string
	SumBigSmall         string
	SpecialCode         string
	NormalCode          string
}

// deriveRecordStats 基于开奖号码自动计算详情字段兜底值。
func deriveRecordStats(numbers []int) recordStats {
	if len(numbers) < 7 {
		return recordStats{}
	}
	total := 0
	for _, n := range numbers {
		total += n
	}
	special := numbers[6]
	return recordStats{
		SpecialSingleDouble: oddEvenCN(special),
		SpecialBigSmall:     bigSmallCN(special, 24),
		SumSingleDouble:     oddEvenCN(total),
		SumBigSmall:         bigSmallCN(total, 175),
		SpecialCode:         fmt.Sprintf("%02d", special),
		NormalCode:          joinPaddedInts(numbers[:6]),
	}
}

// oddEvenCN 输出中文单双。
func oddEvenCN(v int) string {
	if v%2 == 0 {
		return "双"
	}
	return "单"
}

// bigSmallCN 输出中文大小（> threshold 为大）。
func bigSmallCN(v, threshold int) string {
	if v > threshold {
		return "大"
	}
	return "小"
}

// pickString 优先使用手工配置值；为空时使用自动值。
func pickString(manual, auto string) string {
	if strings.TrimSpace(manual) != "" {
		return strings.TrimSpace(manual)
	}
	return strings.TrimSpace(auto)
}
