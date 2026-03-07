package lottery

import (
	"strings"
	"time"
)

// BuildDashboard 生成首页开奖区/开奖现场顶部看板数据。
// 包含：
// - 当前彩种期号与倒计时；
// - 直播流状态（是否显示播放器）；
// - 当期开奖号码与五行标签。
func (s *Service) BuildDashboard(sid uint) (map[string]interface{}, error) {
	// 1) 读取彩种主配置（包含直播流地址、直播状态、开奖倒计时配置）。
	sl, err := s.dao.GetSpecialLottery(sid)
	if err != nil {
		return nil, err
	}
	// 2) 优先读取当前彩种最新一期开奖记录（开奖区主数据）。
	//    若开奖区表暂未录数，则回退到 tk_lottery_info，避免首页开奖区整块丢失。
	current, err := s.dao.GetLatestDrawRecordBySpecialID(sid)
	drawRecordID := uint(0)
	issue := sl.CurrentIssue
	drawAt := ""
	playbackURL := ""
	numbers := make([]int, 0)
	labels := make([]string, 0)
	normalNumbers := make([]int, 0)
	specialNumber := 0
	if err == nil {
		// 2.1) 命中开奖记录时，走开奖区标准字段。
		drawRecordID = current.ID
		issue = current.Issue
		drawAt = current.DrawAt.Format("2006/01/02 15:04:05")
		playbackURL = current.PlaybackURL
		numbers = extractDrawNumbersFromRecord(*current)
		labels = extractDrawLabels(*current, numbers)
		normalNumbers = splitCSVInts(current.NormalDrawResult)
		specialNumbers := splitCSVInts(current.SpecialDrawResult)
		if len(specialNumbers) > 0 {
			specialNumber = specialNumbers[0]
		}
	} else {
		// 2.2) 回退到图纸表（兼容旧数据/初始化阶段）。
		info, infoErr := s.dao.GetLatestLotteryInfoBySpecialID(sid)
		if infoErr != nil {
			return nil, err
		}
		issue = info.Issue
		drawAt = info.DrawAt.Format("2006/01/02 15:04:05")
		playbackURL = info.PlaybackURL
		numbers = extractDrawNumbers(*info)
		labels = buildPairLabels(numbers)
		normalNumbers = splitCSVInts(info.NormalDrawResult)
		specialNumbers := splitCSVInts(info.SpecialDrawResult)
		if len(specialNumbers) > 0 {
			specialNumber = specialNumbers[0]
		}
	}

	// 3) 若彩种主表 current_issue 为空，兜底使用当前开奖号期号。
	if strings.TrimSpace(issue) == "" {
		issue = sl.CurrentIssue
	}
	// 4) 计算距下期开奖倒计时，最小值钳制为 0。
	countdown := int64(sl.NextDrawAt.Sub(time.Now()).Seconds())
	if countdown < 0 {
		countdown = 0
	}

	// 5) 播放器显示策略：
	// - 流地址非空 且（显式启用直播 或 直播状态=live）即展示播放器。
	// - 避免后台 live_enabled 与 live_status 短暂不一致时整块被隐藏。
	showByConfig := strings.TrimSpace(sl.LiveStreamURL) != "" && (sl.LiveEnabled == 1 || sl.LiveStatus == "live")
	hasLiveData := false
	if showByConfig {
		// 异步探测模式下首次请求可能尚未命中缓存；
		// 为避免首页播放器“完全不显示”，播放器展示由配置开关决定，
		// has_data 仅作为前端状态标识（可用于展示“加载中/无信号”提示）。
		hasLiveData = probeLiveStreamAvailable(sl.LiveStreamURL)
	}

	// 6) 统一输出首页/现场页开奖看板数据结构。
	drawResult := joinPaddedInts(numbers)
	if len(drawResult) == 0 {
		drawResult = "--"
	}

	return map[string]interface{}{
		"special_lottery": map[string]interface{}{
			"id":            sl.ID,
			"name":          sl.Name,
			"code":          sl.Code,
			"current_issue": issue,
			"next_draw_at":  sl.NextDrawAt.Format(time.RFC3339),
			"countdown_sec": countdown,
		},
		"live": map[string]interface{}{
			"show_player": showByConfig,
			"has_data":    hasLiveData,
			"status":      sl.LiveStatus,
			"stream_url":  sl.LiveStreamURL,
		},
		"draw": map[string]interface{}{
			"draw_record_id":     drawRecordID,
			"special_lottery_id": sid,
			"issue":              issue,
			"draw_at":            drawAt,
			"normal_numbers":     normalNumbers,
			"special_number":     specialNumber,
			"draw_result":        drawResult,
			"playback_url":       playbackURL,
			"numbers":            numbers,
			"labels":             labels,
		},
	}, nil
}
