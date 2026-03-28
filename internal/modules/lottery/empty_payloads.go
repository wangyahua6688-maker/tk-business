package lottery

import "time"

// emptyDashboardPayload 返回空库下的开奖看板结构，保证前端拿到稳定字段。
func emptyDashboardPayload(sid uint) map[string]interface{} {
	return map[string]interface{}{
		"special_lottery": map[string]interface{}{
			"id":            sid,
			"name":          "",
			"code":          "",
			"current_issue": "",
			"next_draw_at":  "",
		},
		"live": map[string]interface{}{
			"show_player": false,
			"has_data":    false,
			"status":      "",
			"stream_url":  "",
		},
		"draw": map[string]interface{}{
			"draw_record_id":        0,
			"special_lottery_id":    sid,
			"issue":                 "",
			"draw_at":               "",
			"normal_numbers":        []int{},
			"special_number":        0,
			"draw_result":           "",
			"playback_url":          "",
			"numbers":               []int{},
			"labels":                []string{},
			"pair_labels":           []string{},
			"color_labels":          []string{},
			"zodiac_labels":         []string{},
			"wuxing_labels":         []string{},
			"special_single_double": "",
			"special_big_small":     "",
			"sum_single_double":     "",
			"sum_big_small":         "",
			"special_code":          "",
			"normal_code":           "",
			"zheng1":                "",
			"zheng2":                "",
			"zheng3":                "",
			"zheng4":                "",
			"zheng5":                "",
			"zheng6":                "",
		},
	}
}

// emptyLiveScenePayload 返回空库下的开奖现场结构。
func emptyLiveScenePayload(now time.Time) map[string]interface{} {
	return map[string]interface{}{
		"scene_title":      "开奖现场",
		"generated_at":     now.Format(time.RFC3339),
		"tabs":             []map[string]interface{}{},
		"active_tab_id":    0,
		"dashboard":        emptyDashboardPayload(0),
		"cards":            []map[string]interface{}{},
		"playback_items":   []map[string]interface{}{},
		"history_items":    []map[string]interface{}{},
		"recommend_blocks": []map[string]interface{}{},
	}
}

// emptyDrawHistoryPayload 返回空库下的历史开奖结构。
func emptyDrawHistoryPayload(specialLotteryID uint, specialLotteryName, order string, showFive bool) map[string]interface{} {
	return map[string]interface{}{
		"special_lottery_id":   specialLotteryID,
		"special_lottery_name": specialLotteryName,
		"year":                 lotteryNowInEast8().Year(),
		"order_mode":           order,
		"show_five":            showFive,
		"items":                []map[string]interface{}{},
	}
}

// emptyDrawDetailPayload 返回空库下的开奖详情结构。
func emptyDrawDetailPayload(recordID uint) map[string]interface{} {
	return map[string]interface{}{
		"id":                    recordID,
		"special_lottery_id":    0,
		"special_lottery_name":  "",
		"issue":                 "",
		"year":                  0,
		"draw_at":               "",
		"draw_time":             "",
		"normal_draw_result":    "",
		"special_draw_result":   "",
		"draw_result":           "",
		"numbers":               []int{},
		"labels":                []string{},
		"pair_labels":           []string{},
		"color_labels":          []string{},
		"zodiac_labels":         []string{},
		"wuxing_labels":         []string{},
		"playback_url":          "",
		"special_single_double": "",
		"special_big_small":     "",
		"sum_single_double":     "",
		"sum_big_small":         "",
		"recommend_six":         "",
		"recommend_four":        "",
		"recommend_one":         "",
		"recommend_ten":         "",
		"special_code":          "",
		"normal_code":           "",
		"zheng1":                "",
		"zheng2":                "",
		"zheng3":                "",
		"zheng4":                "",
		"zheng5":                "",
		"zheng6":                "",
		"result_bundle":         map[string]interface{}{},
	}
}

// emptyLotteryDetailPayload 返回空库下的图纸详情结构。
func emptyLotteryDetailPayload(infoID uint) map[string]interface{} {
	return map[string]interface{}{
		"current": map[string]interface{}{
			"id":                  infoID,
			"special_lottery_id":  0,
			"title":               "",
			"issue":               "",
			"year":                0,
			"draw_issue":          "",
			"draw_year":           0,
			"draw_record_id":      0,
			"detail_image_url":    "",
			"draw_code":           "",
			"normal_draw_result":  "",
			"special_draw_result": "",
			"draw_result":         "",
			"draw_numbers":        []int{},
			"draw_labels":         []string{},
			"color_labels":        []string{},
			"zodiac_labels":       []string{},
			"wuxing_labels":       []string{},
			"playback_url":        "",
			"likes_count":         0,
			"comment_count":       0,
			"favorite_count":      0,
			"read_count":          0,
		},
		"years":             []int{},
		"issues":            []map[string]interface{}{},
		"poll_options":      []map[string]interface{}{},
		"poll_enabled":      false,
		"poll_default_open": false,
		"show_metrics":      false,
		"detail_banners":    []map[string]interface{}{},
		"recommend_items":   []map[string]interface{}{},
		"external_links":    []map[string]interface{}{},
		"system_comments":   []map[string]interface{}{},
		"user_comments":     []map[string]interface{}{},
		"hot_comments":      []map[string]interface{}{},
		"latest_comments":   []map[string]interface{}{},
	}
}
