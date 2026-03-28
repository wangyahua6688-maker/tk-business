package lottery

import (
	"strconv"
	"strings"

	common_model "github.com/wangyahua6688-maker/tk-common/models"
)

// splitCSVInts 将逗号分隔号码字符串转为整型数组。
func splitCSVInts(raw string) []int {
	// 定义并初始化当前变量。
	parts := strings.FieldsFunc(strings.TrimSpace(raw), func(r rune) bool {
		// 返回当前处理结果。
		return r == ',' || r == '|' || r == '/' || r == ' ' || r == '\t' || r == '\n'
	})
	// 定义并初始化当前变量。
	out := make([]int, 0, len(parts))
	// 循环处理当前数据集合。
	for _, p := range parts {
		// 定义并初始化当前变量。
		v, err := strconv.Atoi(strings.TrimSpace(p))
		// 判断条件并进入对应分支逻辑。
		if err == nil {
			// 更新当前变量或字段值。
			out = append(out, v)
		}
	}
	// 返回当前处理结果。
	return out
}

// extractDrawNumbers 优先使用“普通号+特别号”字段，回退到兼容字段 draw_result。
func extractDrawNumbers(info common_model.WLotteryInfo) []int {
	// 定义并初始化当前变量。
	normal := splitCSVInts(info.NormalDrawResult)
	// 定义并初始化当前变量。
	special := splitCSVInts(info.SpecialDrawResult)
	// 判断条件并进入对应分支逻辑。
	if len(normal) == 6 && len(special) == 1 {
		// 返回当前处理结果。
		return append(normal, special[0])
	}
	// 返回当前处理结果。
	return splitCSVInts(info.DrawResult)
}

// extractDrawNumbersFromRecord 从开奖记录提取 6+1 开奖号码。
func extractDrawNumbersFromRecord(record common_model.WDrawRecord) []int {
	// 定义并初始化当前变量。
	normal := splitCSVInts(record.NormalDrawResult)
	// 定义并初始化当前变量。
	special := splitCSVInts(record.SpecialDrawResult)
	// 判断条件并进入对应分支逻辑。
	if len(normal) == 6 && len(special) == 1 {
		// 返回当前处理结果。
		return append(normal, special[0])
	}
	// 返回当前处理结果。
	return splitCSVInts(record.DrawResult)
}

// buildSimpleLabels 基于官方号码映射生成“生肖/五行”组合标签。
func buildSimpleLabels(numbers []int) []string {
	return buildPairLabels(numbers)
}

// extractDrawLabels 优先读取开奖记录标签；标签缺失时自动生成“生肖/五行”标签。
func extractDrawLabels(record common_model.WDrawRecord, numbers []int) []string {
	labels := splitCSVLabels(record.DrawLabels)
	if len(labels) == len(numbers) && len(labels) > 0 {
		return labels
	}
	return buildPairLabels(numbers)
}

// extractColorLabels 提取波色标签。
func extractColorLabels(record common_model.WDrawRecord, numbers []int) []string {
	labels := splitCSVLabels(record.ColorLabels)
	if len(labels) == len(numbers) && len(labels) > 0 {
		return labels
	}
	return buildColorLabels(numbers)
}

// extractZodiacAndWuxingLabels 提取“属相/五行”两组标签。
func extractZodiacAndWuxingLabels(record common_model.WDrawRecord, numbers []int) ([]string, []string) {
	zodiac := splitCSVLabels(record.ZodiacLabels)
	wuxing := splitCSVLabels(record.WuxingLabels)
	if len(zodiac) == len(numbers) && len(wuxing) == len(numbers) && len(zodiac) > 0 {
		return zodiac, wuxing
	}
	return buildZodiacAndWuxingLabels(numbers)
}

// splitCSVLabels 解析开奖记录标签串（逗号/空格/换行分隔）。
func splitCSVLabels(raw string) []string {
	parts := strings.FieldsFunc(strings.TrimSpace(raw), func(r rune) bool {
		return r == ',' || r == '|' || r == ';' || r == '\n' || r == '\r' || r == '\t'
	})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return out
}

// buildColorLabels 生成波色标签，用于旧数据或主表字段缺失时的展示兜底。
func buildColorLabels(numbers []int) []string {
	out := make([]string, 0, len(numbers))
	for _, n := range numbers {
		out = append(out, common_model.DrawResultColorWaveMap[n])
	}
	return out
}

// buildZodiacAndWuxingLabels 基于官方号码映射生成“属相/五行”两组标签。
func buildZodiacAndWuxingLabels(numbers []int) ([]string, []string) {
	zodiac := make([]string, 0, len(numbers))
	wuxing := make([]string, 0, len(numbers))
	for _, n := range numbers {
		zodiac = append(zodiac, common_model.DrawResultZodiacMap[n])
		wuxing = append(wuxing, common_model.DrawResultWuxingMap[n])
	}
	return zodiac, wuxing
}

// buildPairLabels 生成“生肖/五行”标签，用于开奖记录标签缺失时的展示兜底。
func buildPairLabels(numbers []int) []string {
	zodiac, wuxing := buildZodiacAndWuxingLabels(numbers)
	out := make([]string, 0, len(numbers))
	for idx := range numbers {
		label := zodiac[idx]
		if wuxing[idx] != "" {
			label = strings.TrimSpace(label + "/" + wuxing[idx])
		}
		out = append(out, label)
	}
	return out
}

// splitPairLabel 将“生肖/五行”组合标签拆成两个值。
func splitPairLabel(raw string) (string, string) {
	// 定义并初始化当前变量。
	value := strings.TrimSpace(raw)
	// 判断条件并进入对应分支逻辑。
	if value == "" {
		// 返回当前处理结果。
		return "", ""
	}
	// 定义并初始化当前变量。
	parts := strings.SplitN(value, "/", 2)
	// 判断条件并进入对应分支逻辑。
	if len(parts) == 2 {
		// 返回当前处理结果。
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	// 返回当前处理结果。
	return value, ""
}

// sortYearsDesc 对年份进行降序排序。
func sortYearsDesc(years []int) {
	// 循环处理当前数据集合。
	for i := 0; i < len(years)-1; i++ {
		// 循环处理当前数据集合。
		for j := i + 1; j < len(years); j++ {
			// 判断条件并进入对应分支逻辑。
			if years[j] > years[i] {
				// 更新当前变量或字段值。
				years[i], years[j] = years[j], years[i]
			}
		}
	}
}

// buildPollPayload 计算投票百分比并输出投票列表结构。
func buildPollPayload(options []common_model.WLotteryOption, totalVotes int64) []map[string]interface{} {
	// 定义并初始化当前变量。
	poll := make([]map[string]interface{}, 0, len(options))
	// 循环处理当前数据集合。
	for _, opt := range options {
		// 定义并初始化当前变量。
		percent := 0.0
		// 判断条件并进入对应分支逻辑。
		if totalVotes > 0 {
			// 更新当前变量或字段值。
			percent = float64(opt.Votes) * 100 / float64(totalVotes)
		}
		// 更新当前变量或字段值。
		poll = append(poll, map[string]interface{}{
			// 处理当前语句逻辑。
			"id": opt.ID,
			// 处理当前语句逻辑。
			"name": opt.OptionName,
			// 处理当前语句逻辑。
			"votes": opt.Votes,
			// 处理当前语句逻辑。
			"percent": percent,
		})
	}
	// 返回当前处理结果。
	return poll
}

// parseCSVUintIDs 解析“推荐图纸ID列表”字段。
func parseCSVUintIDs(raw string) []uint {
	// 定义并初始化当前变量。
	parts := strings.Split(strings.TrimSpace(raw), ",")
	// 定义并初始化当前变量。
	out := make([]uint, 0, len(parts))
	// 定义并初始化当前变量。
	seen := map[uint]struct{}{}
	// 循环处理当前数据集合。
	for _, item := range parts {
		// 定义并初始化当前变量。
		v, err := strconv.Atoi(strings.TrimSpace(item))
		// 判断条件并进入对应分支逻辑。
		if err != nil || v <= 0 {
			// 处理当前语句逻辑。
			continue
		}
		// 定义并初始化当前变量。
		id := uint(v)
		// 判断条件并进入对应分支逻辑。
		if _, ok := seen[id]; ok {
			// 处理当前语句逻辑。
			continue
		}
		// 更新当前变量或字段值。
		seen[id] = struct{}{}
		// 更新当前变量或字段值。
		out = append(out, id)
	}
	// 返回当前处理结果。
	return out
}

// reorderInfosByIDs 按配置 ID 顺序重排推荐图纸，并过滤当前图纸自身。
func reorderInfosByIDs(rows []common_model.WLotteryInfo, orderedIDs []uint, currentID uint) []common_model.WLotteryInfo {
	// 定义并初始化当前变量。
	rowMap := make(map[uint]common_model.WLotteryInfo, len(rows))
	// 循环处理当前数据集合。
	for _, row := range rows {
		// 判断条件并进入对应分支逻辑。
		if row.ID == currentID {
			// 处理当前语句逻辑。
			continue
		}
		// 更新当前变量或字段值。
		rowMap[row.ID] = row
	}
	// 定义并初始化当前变量。
	out := make([]common_model.WLotteryInfo, 0, len(rows))
	// 循环处理当前数据集合。
	for _, id := range orderedIDs {
		// 判断条件并进入对应分支逻辑。
		if row, ok := rowMap[id]; ok {
			// 更新当前变量或字段值。
			out = append(out, row)
			// 调用delete完成当前处理。
			delete(rowMap, id)
		}
	}
	// 返回当前处理结果。
	return out
}

// buildRecommendPayload 组装推荐图纸返回结构。
func buildRecommendPayload(rows []common_model.WLotteryInfo) []map[string]interface{} {
	// 定义并初始化当前变量。
	out := make([]map[string]interface{}, 0, len(rows))
	// 循环处理当前数据集合。
	for _, row := range rows {
		// 更新当前变量或字段值。
		out = append(out, map[string]interface{}{
			// 处理当前语句逻辑。
			"id": row.ID,
			// 处理当前语句逻辑。
			"title": row.Title,
			// 处理当前语句逻辑。
			"issue": row.Issue,
			// 处理当前语句逻辑。
			"cover_image_url": row.CoverImageURL,
		})
	}
	// 返回当前处理结果。
	return out
}

// buildExternalLinkPayload 组装详情页外链列表结构。
func buildExternalLinkPayload(rows []common_model.WExternalLink) []map[string]interface{} {
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
			"url": row.URL,
		})
	}
	// 返回当前处理结果。
	return out
}
