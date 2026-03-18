package lottery

import (
	"strconv"
	"strings"

	"tk-common/models"
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
func extractDrawNumbers(info models.WLotteryInfo) []int {
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
func extractDrawNumbersFromRecord(record models.WDrawRecord) []int {
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

// buildSimpleLabels 按号码生成占位五行标签（后续可替换为真实映射规则）。
func buildSimpleLabels(numbers []int) []string {
	// 定义并初始化当前变量。
	labels := make([]string, 0, len(numbers))
	// 定义并初始化当前变量。
	elements := []string{"金", "木", "水", "火", "土"}
	// 循环处理当前数据集合。
	for _, n := range numbers {
		// 更新当前变量或字段值。
		labels = append(labels, elements[n%len(elements)])
	}
	// 返回当前处理结果。
	return labels
}

// extractDrawLabels 优先读取开奖记录标签；标签缺失时自动生成“生肖/五行”占位标签。
func extractDrawLabels(record models.WDrawRecord, numbers []int) []string {
	// 定义并初始化当前变量。
	labels := splitCSVLabels(record.DrawLabels)
	// 判断条件并进入对应分支逻辑。
	if len(labels) == len(numbers) && len(labels) > 0 {
		// 返回当前处理结果。
		return labels
	}
	// 返回当前处理结果。
	return buildPairLabels(numbers)
}

// extractZodiacAndWuxingLabels 提取“属相/五行”两组标签。
func extractZodiacAndWuxingLabels(record models.WDrawRecord, numbers []int) ([]string, []string) {
	// 1) 优先使用独立字段，避免前端每次拆分“属相/五行”组合串。
	zodiac := splitCSVLabels(record.ZodiacLabels)
	// 定义并初始化当前变量。
	wuxing := splitCSVLabels(record.WuxingLabels)
	// 判断条件并进入对应分支逻辑。
	if len(zodiac) == len(numbers) && len(wuxing) == len(numbers) && len(zodiac) > 0 {
		// 返回当前处理结果。
		return zodiac, wuxing
	}

	// 2) 兼容旧数据：从 draw_labels（生肖/五行）拆分。
	paired := extractDrawLabels(record, numbers)
	// 判断条件并进入对应分支逻辑。
	if len(paired) == 0 {
		// 返回当前处理结果。
		return []string{}, []string{}
	}
	// 更新当前变量或字段值。
	zodiac = make([]string, 0, len(paired))
	// 更新当前变量或字段值。
	wuxing = make([]string, 0, len(paired))
	// 循环处理当前数据集合。
	for _, item := range paired {
		// 定义并初始化当前变量。
		z, w := splitPairLabel(item)
		// 更新当前变量或字段值。
		zodiac = append(zodiac, z)
		// 更新当前变量或字段值。
		wuxing = append(wuxing, w)
	}
	// 返回当前处理结果。
	return zodiac, wuxing
}

// splitCSVLabels 解析开奖记录标签串（逗号/空格/换行分隔）。
func splitCSVLabels(raw string) []string {
	// 定义并初始化当前变量。
	parts := strings.FieldsFunc(strings.TrimSpace(raw), func(r rune) bool {
		// 返回当前处理结果。
		return r == ',' || r == '|' || r == ';' || r == '\n' || r == '\r' || r == '\t'
	})
	// 定义并初始化当前变量。
	out := make([]string, 0, len(parts))
	// 循环处理当前数据集合。
	for _, p := range parts {
		// 定义并初始化当前变量。
		v := strings.TrimSpace(p)
		// 判断条件并进入对应分支逻辑。
		if v == "" {
			// 处理当前语句逻辑。
			continue
		}
		// 更新当前变量或字段值。
		out = append(out, v)
	}
	// 返回当前处理结果。
	return out
}

// buildPairLabels 生成“生肖/五行”标签，用于开奖记录标签缺失时的展示兜底。
func buildPairLabels(numbers []int) []string {
	// 定义并初始化当前变量。
	zodiacs := []string{"鼠", "牛", "虎", "兔", "龙", "蛇", "马", "羊", "猴", "鸡", "狗", "猪"}
	// 定义并初始化当前变量。
	elements := []string{"金", "木", "水", "火", "土"}
	// 定义并初始化当前变量。
	out := make([]string, 0, len(numbers))
	// 循环处理当前数据集合。
	for _, n := range numbers {
		// 定义并初始化当前变量。
		zodiac := zodiacs[(n-1)%len(zodiacs)]
		// 定义并初始化当前变量。
		element := elements[(n-1)%len(elements)]
		// 更新当前变量或字段值。
		out = append(out, zodiac+"/"+element)
	}
	// 返回当前处理结果。
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
func buildPollPayload(options []models.WLotteryOption, totalVotes int64) []map[string]interface{} {
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
func reorderInfosByIDs(rows []models.WLotteryInfo, orderedIDs []uint, currentID uint) []models.WLotteryInfo {
	// 定义并初始化当前变量。
	rowMap := make(map[uint]models.WLotteryInfo, len(rows))
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
	out := make([]models.WLotteryInfo, 0, len(rows))
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
func buildRecommendPayload(rows []models.WLotteryInfo) []map[string]interface{} {
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
func buildExternalLinkPayload(rows []models.WExternalLink) []map[string]interface{} {
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
