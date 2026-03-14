package lottery

import (
	"strconv"
	"strings"

	"tk-common/models"
)

// splitCSVInts 将逗号分隔号码字符串转为整型数组。
func splitCSVInts(raw string) []int {
	parts := strings.FieldsFunc(strings.TrimSpace(raw), func(r rune) bool {
		return r == ',' || r == '|' || r == '/' || r == ' ' || r == '\t' || r == '\n'
	})
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err == nil {
			out = append(out, v)
		}
	}
	return out
}

// extractDrawNumbers 优先使用“普通号+特别号”字段，回退到兼容字段 draw_result。
func extractDrawNumbers(info models.WLotteryInfo) []int {
	normal := splitCSVInts(info.NormalDrawResult)
	special := splitCSVInts(info.SpecialDrawResult)
	if len(normal) == 6 && len(special) == 1 {
		return append(normal, special[0])
	}
	return splitCSVInts(info.DrawResult)
}

// extractDrawNumbersFromRecord 从开奖记录提取 6+1 开奖号码。
func extractDrawNumbersFromRecord(record models.WDrawRecord) []int {
	normal := splitCSVInts(record.NormalDrawResult)
	special := splitCSVInts(record.SpecialDrawResult)
	if len(normal) == 6 && len(special) == 1 {
		return append(normal, special[0])
	}
	return splitCSVInts(record.DrawResult)
}

// buildSimpleLabels 按号码生成占位五行标签（后续可替换为真实映射规则）。
func buildSimpleLabels(numbers []int) []string {
	labels := make([]string, 0, len(numbers))
	elements := []string{"金", "木", "水", "火", "土"}
	for _, n := range numbers {
		labels = append(labels, elements[n%len(elements)])
	}
	return labels
}

// extractDrawLabels 优先读取开奖记录标签；标签缺失时自动生成“生肖/五行”占位标签。
func extractDrawLabels(record models.WDrawRecord, numbers []int) []string {
	labels := splitCSVLabels(record.DrawLabels)
	if len(labels) == len(numbers) && len(labels) > 0 {
		return labels
	}
	return buildPairLabels(numbers)
}

// extractZodiacAndWuxingLabels 提取“属相/五行”两组标签。
func extractZodiacAndWuxingLabels(record models.WDrawRecord, numbers []int) ([]string, []string) {
	// 1) 优先使用独立字段，避免前端每次拆分“属相/五行”组合串。
	zodiac := splitCSVLabels(record.ZodiacLabels)
	wuxing := splitCSVLabels(record.WuxingLabels)
	if len(zodiac) == len(numbers) && len(wuxing) == len(numbers) && len(zodiac) > 0 {
		return zodiac, wuxing
	}

	// 2) 兼容旧数据：从 draw_labels（生肖/五行）拆分。
	paired := extractDrawLabels(record, numbers)
	if len(paired) == 0 {
		return []string{}, []string{}
	}
	zodiac = make([]string, 0, len(paired))
	wuxing = make([]string, 0, len(paired))
	for _, item := range paired {
		z, w := splitPairLabel(item)
		zodiac = append(zodiac, z)
		wuxing = append(wuxing, w)
	}
	return zodiac, wuxing
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

// buildPairLabels 生成“生肖/五行”标签，用于开奖记录标签缺失时的展示兜底。
func buildPairLabels(numbers []int) []string {
	zodiacs := []string{"鼠", "牛", "虎", "兔", "龙", "蛇", "马", "羊", "猴", "鸡", "狗", "猪"}
	elements := []string{"金", "木", "水", "火", "土"}
	out := make([]string, 0, len(numbers))
	for _, n := range numbers {
		zodiac := zodiacs[(n-1)%len(zodiacs)]
		element := elements[(n-1)%len(elements)]
		out = append(out, zodiac+"/"+element)
	}
	return out
}

// splitPairLabel 将“生肖/五行”组合标签拆成两个值。
func splitPairLabel(raw string) (string, string) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", ""
	}
	parts := strings.SplitN(value, "/", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return value, ""
}

// sortYearsDesc 对年份进行降序排序。
func sortYearsDesc(years []int) {
	for i := 0; i < len(years)-1; i++ {
		for j := i + 1; j < len(years); j++ {
			if years[j] > years[i] {
				years[i], years[j] = years[j], years[i]
			}
		}
	}
}

// buildPollPayload 计算投票百分比并输出投票列表结构。
func buildPollPayload(options []models.WLotteryOption, totalVotes int64) []map[string]interface{} {
	poll := make([]map[string]interface{}, 0, len(options))
	for _, opt := range options {
		percent := 0.0
		if totalVotes > 0 {
			percent = float64(opt.Votes) * 100 / float64(totalVotes)
		}
		poll = append(poll, map[string]interface{}{
			"id":      opt.ID,
			"name":    opt.OptionName,
			"votes":   opt.Votes,
			"percent": percent,
		})
	}
	return poll
}

// parseCSVUintIDs 解析“推荐图纸ID列表”字段。
func parseCSVUintIDs(raw string) []uint {
	parts := strings.Split(strings.TrimSpace(raw), ",")
	out := make([]uint, 0, len(parts))
	seen := map[uint]struct{}{}
	for _, item := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(item))
		if err != nil || v <= 0 {
			continue
		}
		id := uint(v)
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// reorderInfosByIDs 按配置 ID 顺序重排推荐图纸，并过滤当前图纸自身。
func reorderInfosByIDs(rows []models.WLotteryInfo, orderedIDs []uint, currentID uint) []models.WLotteryInfo {
	rowMap := make(map[uint]models.WLotteryInfo, len(rows))
	for _, row := range rows {
		if row.ID == currentID {
			continue
		}
		rowMap[row.ID] = row
	}
	out := make([]models.WLotteryInfo, 0, len(rows))
	for _, id := range orderedIDs {
		if row, ok := rowMap[id]; ok {
			out = append(out, row)
			delete(rowMap, id)
		}
	}
	return out
}

// buildRecommendPayload 组装推荐图纸返回结构。
func buildRecommendPayload(rows []models.WLotteryInfo) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]interface{}{
			"id":              row.ID,
			"title":           row.Title,
			"issue":           row.Issue,
			"cover_image_url": row.CoverImageURL,
		})
	}
	return out
}

// buildExternalLinkPayload 组装详情页外链列表结构。
func buildExternalLinkPayload(rows []models.WExternalLink) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]interface{}{
			"id":   row.ID,
			"name": row.Name,
			"url":  row.URL,
		})
	}
	return out
}
