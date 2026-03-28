package lottery

import (
	"strconv"
	"strings"

	common_model "github.com/wangyahua6688-maker/tk-common/models"
)

// compiledNumberDetailView 是面向接口输出的号码规则明细。
type compiledNumberDetailView struct {
	Number          int    `json:"number"`
	Position        int    `json:"position"`
	ColorWave       string `json:"color_wave"`
	BigSmall        string `json:"big_small"`
	SingleDouble    string `json:"single_double"`
	SumSingleDouble string `json:"sum_single_double"`
	TailBigSmall    string `json:"tail_big_small"`
	Zodiac          string `json:"zodiac"`
	Wuxing          string `json:"wuxing"`
	Beast           string `json:"beast"`
	TailLabel       string `json:"tail_label"`
}

// compileNumberDetailView 按六合彩固定规则编译单个号码的玩法属性。
func compileNumberDetailView(num, position int) compiledNumberDetailView {
	zodiac := common_model.DrawResultZodiacMap[num]
	return compiledNumberDetailView{
		Number:          num,
		Position:        position,
		ColorWave:       common_model.DrawResultColorWaveMap[num],
		BigSmall:        resultSpecialBigSmall(num),
		SingleDouble:    resultSpecialSingleDouble(num),
		SumSingleDouble: resultSpecialSumSingleDouble(num),
		TailBigSmall:    resultSpecialTailBigSmall(num),
		Zodiac:          zodiac,
		Wuxing:          common_model.DrawResultWuxingMap[num],
		Beast:           common_model.DrawResultBeastMap[zodiac],
		TailLabel:       strconv.Itoa(num%10) + "尾",
	}
}

// resultSpecialBigSmall 按特码/正码规则返回大小。
func resultSpecialBigSmall(num int) string {
	if num == 49 {
		return "和"
	}
	if num >= 25 {
		return "大"
	}
	return "小"
}

// resultSpecialSingleDouble 按特码/正码规则返回单双。
func resultSpecialSingleDouble(num int) string {
	if num == 49 {
		return "和"
	}
	if num%2 == 0 {
		return "双"
	}
	return "单"
}

// resultSpecialSumSingleDouble 按特码/正码规则返回合数单双。
func resultSpecialSumSingleDouble(num int) string {
	if num == 49 {
		return "和"
	}
	sum := num/10 + num%10
	if sum%2 == 0 {
		return "合双"
	}
	return "合单"
}

// resultSpecialTailBigSmall 按特码尾数规则返回尾大尾小。
func resultSpecialTailBigSmall(num int) string {
	if num == 49 {
		return "和"
	}
	if num%10 <= 4 {
		return "尾小"
	}
	return "尾大"
}

// resultCountAsOdd 七码统计里 49 按单计。
func resultCountAsOdd(num int) bool {
	if num == 49 {
		return true
	}
	return num%2 == 1
}

// resultCountAsBig 七码统计里 49 按大计。
func resultCountAsBig(num int) bool {
	if num == 49 {
		return true
	}
	return num >= 25
}

// resultHalfWaveColorSize 按特码半波规则输出“波色+大小”。
func resultHalfWaveColorSize(detail compiledNumberDetailView) string {
	if detail.Number == 49 {
		return "和局"
	}
	return strings.TrimSuffix(detail.ColorWave, "波") + detail.BigSmall
}

// resultHalfWaveColorParity 按特码半波规则输出“波色+单双”。
func resultHalfWaveColorParity(detail compiledNumberDetailView) string {
	if detail.Number == 49 {
		return "和局"
	}
	return strings.TrimSuffix(detail.ColorWave, "波") + detail.SingleDouble
}

// composeZhengDescriptionView 输出正码1-6的人类可读描述。
func composeZhengDescriptionView(detail compiledNumberDetailView) string {
	return strings.Join([]string{
		detail.BigSmall,
		detail.SingleDouble,
		detail.ColorWave,
		detail.SumSingleDouble,
		detail.TailBigSmall,
		detail.Zodiac,
		detail.Wuxing,
	}, ",")
}

// collectOrderedSetView 按固定顺序输出命中集合。
func collectOrderedSetView(order []string, set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for _, label := range order {
		if _, ok := set[label]; ok {
			out = append(out, label)
		}
	}
	return out
}

// diffOrderedSetView 按固定顺序输出未命中集合。
func diffOrderedSetView(order []string, set map[string]struct{}) []string {
	out := make([]string, 0, len(order))
	for _, label := range order {
		if _, ok := set[label]; !ok {
			out = append(out, label)
		}
	}
	return out
}

// buildDrawResultBundleView 按开奖号码生成完整玩法结果结构，供接口直接复用。
func buildDrawResultBundleView(numbers []int) map[string]interface{} {
	if len(numbers) != 7 {
		return map[string]interface{}{}
	}

	details := make([]compiledNumberDetailView, 0, len(numbers))
	normalNumbers := make([]int, 0, 6)
	colorLabels := make([]string, 0, len(numbers))
	zodiacLabels := make([]string, 0, len(numbers))
	wuxingLabels := make([]string, 0, len(numbers))
	pairLabels := make([]string, 0, len(numbers))
	totalSum := 0
	oddCount, evenCount, bigCount, smallCount := 0, 0, 0, 0
	appearedZodiacSet := map[string]struct{}{}
	appearedTailSet := map[string]struct{}{}
	appearedWuxingSet := map[string]struct{}{}
	homeBeastHitSet := map[string]struct{}{}
	wildBeastHitSet := map[string]struct{}{}
	zhengDescriptions := [6]string{}

	for idx, num := range numbers {
		detail := compileNumberDetailView(num, idx+1)
		details = append(details, detail)
		totalSum += num
		colorLabels = append(colorLabels, detail.ColorWave)
		zodiacLabels = append(zodiacLabels, detail.Zodiac)
		wuxingLabels = append(wuxingLabels, detail.Wuxing)
		pairLabels = append(pairLabels, detail.Zodiac+"/"+detail.Wuxing)
		appearedZodiacSet[detail.Zodiac] = struct{}{}
		appearedTailSet[detail.TailLabel] = struct{}{}
		appearedWuxingSet[detail.Wuxing] = struct{}{}
		if detail.Beast == "家畜" {
			homeBeastHitSet[detail.Zodiac] = struct{}{}
		}
		if detail.Beast == "野兽" {
			wildBeastHitSet[detail.Zodiac] = struct{}{}
		}
		if resultCountAsOdd(num) {
			oddCount++
		} else {
			evenCount++
		}
		if resultCountAsBig(num) {
			bigCount++
		} else {
			smallCount++
		}
		if idx < 6 {
			normalNumbers = append(normalNumbers, num)
			zhengDescriptions[idx] = composeZhengDescriptionView(detail)
		}
	}

	special := details[6]
	totalBigSmall := "小"
	if totalSum >= 175 {
		totalBigSmall = "大"
	}
	totalSingleDouble := "双"
	if totalSum%2 == 1 {
		totalSingleDouble = "单"
	}

	appearedZodiacs := collectOrderedSetView(common_model.DrawResultZodiacOrder, appearedZodiacSet)
	missedZodiacs := diffOrderedSetView(common_model.DrawResultZodiacOrder, appearedZodiacSet)
	appearedTails := collectOrderedSetView(common_model.DrawResultTailOrder, appearedTailSet)
	missedTails := diffOrderedSetView(common_model.DrawResultTailOrder, appearedTailSet)
	appearedWuxings := collectOrderedSetView(common_model.DrawResultWuxingOrder, appearedWuxingSet)
	homeBeastZodiacs := collectOrderedSetView(common_model.DrawResultZodiacOrder, homeBeastHitSet)
	wildBeastZodiacs := collectOrderedSetView(common_model.DrawResultZodiacOrder, wildBeastHitSet)

	return map[string]interface{}{
		"labels": map[string]interface{}{
			"pair_labels":   pairLabels,
			"color_labels":  colorLabels,
			"zodiac_labels": zodiacLabels,
			"wuxing_labels": wuxingLabels,
		},
		"special_result": map[string]interface{}{
			"special_number":            special.Number,
			"special_color_wave":        special.ColorWave,
			"special_big_small":         special.BigSmall,
			"special_single_double":     special.SingleDouble,
			"special_sum_single_double": special.SumSingleDouble,
			"special_tail_big_small":    special.TailBigSmall,
			"special_zodiac":            special.Zodiac,
			"special_wuxing":            special.Wuxing,
			"special_home_beast":        special.Beast,
			"half_wave_color_size":      resultHalfWaveColorSize(special),
			"half_wave_color_parity":    resultHalfWaveColorParity(special),
			"payload": map[string]interface{}{
				"special":                special,
				"two_sides":              []string{special.BigSmall, special.SingleDouble},
				"half_wave_color_size":   resultHalfWaveColorSize(special),
				"half_wave_color_parity": resultHalfWaveColorParity(special),
			},
		},
		"regular_result": map[string]interface{}{
			"normal_numbers":      normalNumbers,
			"total_sum":           totalSum,
			"total_big_small":     totalBigSmall,
			"total_single_double": totalSingleDouble,
			"zheng1":              zhengDescriptions[0],
			"zheng2":              zhengDescriptions[1],
			"zheng3":              zhengDescriptions[2],
			"zheng4":              zhengDescriptions[3],
			"zheng5":              zhengDescriptions[4],
			"zheng6":              zhengDescriptions[5],
			"positions":           details[:6],
		},
		"count_result": map[string]interface{}{
			"total_sum":             totalSum,
			"odd_count":             oddCount,
			"even_count":            evenCount,
			"big_count":             bigCount,
			"small_count":           smallCount,
			"distinct_zodiac_count": len(appearedZodiacs),
			"distinct_tail_count":   len(appearedTails),
			"distinct_wuxing_count": len(appearedWuxings),
			"appeared_zodiacs":      appearedZodiacs,
			"missed_zodiacs":        missedZodiacs,
			"appeared_tails":        appearedTails,
			"missed_tails":          missedTails,
			"appeared_wuxings":      appearedWuxings,
		},
		"zodiac_tail_result": map[string]interface{}{
			"special_zodiac":     special.Zodiac,
			"special_home_beast": special.Beast,
			"special_wuxing":     special.Wuxing,
			"hit_zodiacs":        appearedZodiacs,
			"miss_zodiacs":       missedZodiacs,
			"hit_tails":          appearedTails,
			"miss_tails":         missedTails,
			"home_beast_zodiacs": homeBeastZodiacs,
			"wild_beast_zodiacs": wildBeastZodiacs,
		},
		"combo_result": map[string]interface{}{
			"normal_numbers": normalNumbers,
			"all_numbers":    numbers,
			"special_number": special.Number,
		},
	}
}
