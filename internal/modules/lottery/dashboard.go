package lottery

import (
	"strings"
	"time"
)

// normalizeNextDrawAt 规范化下一期开奖时间。
// 规则：
// 1) 若配置时间已在未来，直接使用；
// 2) 若配置时间已过期，按“每天同一时刻”顺延到未来最近一次。
func normalizeNextDrawAt(base time.Time, now time.Time) time.Time {
	// 判断条件并进入对应分支逻辑。
	if base.After(now) {
		// 返回当前处理结果。
		return base
	}
	// 定义并初始化当前变量。
	next := base
	// 循环处理当前数据集合。
	for !next.After(now) {
		// 更新当前变量或字段值。
		next = next.Add(24 * time.Hour)
	}
	// 返回当前处理结果。
	return next
}

// BuildDashboard 生成首页开奖区/开奖现场顶部看板数据。
// 包含：
// - 当前彩种期号与倒计时；
// - 直播流状态（是否显示播放器）；
// - 当期开奖号码与五行标签。
func (s *Service) BuildDashboard(sid uint) (map[string]interface{}, error) {
	// 1) 读取彩种主配置（包含直播流地址、直播状态、开奖倒计时配置）。
	sl, err := s.dao.GetSpecialLottery(sid)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 2) 优先读取当前彩种最新一期开奖记录（开奖区主数据）。
	//    若开奖区表暂未录数，则回退到 tk_lottery_info，避免首页开奖区整块丢失。
	current, err := s.dao.GetLatestDrawRecordBySpecialID(sid)
	// 定义并初始化当前变量。
	drawRecordID := uint(0)
	// 声明当前变量。
	var issue string
	// 定义并初始化当前变量。
	drawAt := ""
	// 定义并初始化当前变量。
	playbackURL := ""
	// 声明当前变量。
	var numbers []int
	// 声明当前变量。
	var labels []string
	// 声明当前变量。
	var normalNumbers []int
	// 定义并初始化当前变量。
	specialNumber := 0
	// 判断条件并进入对应分支逻辑。
	if err == nil {
		// 2.1) 命中开奖记录时，走开奖区标准字段。
		drawRecordID = current.ID
		// 更新当前变量或字段值。
		issue = current.Issue
		// 更新当前变量或字段值。
		drawAt = current.DrawAt.Format("2006/01/02 15:04:05")
		// 更新当前变量或字段值。
		playbackURL = current.PlaybackURL
		// 更新当前变量或字段值。
		numbers = extractDrawNumbersFromRecord(*current)
		// 更新当前变量或字段值。
		labels = extractDrawLabels(*current, numbers)
		// 更新当前变量或字段值。
		normalNumbers = splitCSVInts(current.NormalDrawResult)
		// 定义并初始化当前变量。
		specialNumbers := splitCSVInts(current.SpecialDrawResult)
		// 判断条件并进入对应分支逻辑。
		if len(specialNumbers) > 0 {
			// 更新当前变量或字段值。
			specialNumber = specialNumbers[0]
		}
		// 进入新的代码块进行处理。
	} else {
		// 2.2) 回退到图纸表（兼容旧数据/初始化阶段）。
		info, infoErr := s.dao.GetLatestLotteryInfoBySpecialID(sid)
		// 判断条件并进入对应分支逻辑。
		if infoErr != nil {
			// 返回当前处理结果。
			return nil, err
		}
		// 更新当前变量或字段值。
		issue = info.Issue
		// 更新当前变量或字段值。
		drawAt = info.DrawAt.Format("2006/01/02 15:04:05")
		// 更新当前变量或字段值。
		playbackURL = info.PlaybackURL
		// 更新当前变量或字段值。
		numbers = extractDrawNumbers(*info)
		// 更新当前变量或字段值。
		labels = buildPairLabels(numbers)
		// 更新当前变量或字段值。
		normalNumbers = splitCSVInts(info.NormalDrawResult)
		// 定义并初始化当前变量。
		specialNumbers := splitCSVInts(info.SpecialDrawResult)
		// 判断条件并进入对应分支逻辑。
		if len(specialNumbers) > 0 {
			// 更新当前变量或字段值。
			specialNumber = specialNumbers[0]
		}
	}

	// 3) 若彩种主表 current_issue 为空，兜底使用当前开奖号期号。
	if strings.TrimSpace(issue) == "" {
		// 更新当前变量或字段值。
		issue = sl.CurrentIssue
	}
	// 4) 倒计时基准时间兜底：当配置时间已过期时，顺延到未来最近一次（通常为次日同一时刻）。
	nextDrawAt := normalizeNextDrawAt(sl.NextDrawAt, time.Now())
	// 5) 播放器显示策略：
	// - 流地址非空 且（显式启用直播 或 直播状态=live）即展示播放器。
	// - 避免后台 live_enabled 与 live_status 短暂不一致时整块被隐藏。
	showByConfig := strings.TrimSpace(sl.LiveStreamURL) != "" && (sl.LiveEnabled == 1 || sl.LiveStatus == "live")
	// 定义并初始化当前变量。
	hasLiveData := false
	// 判断条件并进入对应分支逻辑。
	if showByConfig {
		// 异步探测模式下首次请求可能尚未命中缓存；
		// 为避免首页播放器“完全不显示”，播放器展示由配置开关决定，
		// has_data 仅作为前端状态标识（可用于展示“加载中/无信号”提示）。
		hasLiveData = probeLiveStreamAvailable(sl.LiveStreamURL)
	}

	// 6) 统一输出首页/现场页开奖看板数据结构。
	drawResult := joinPaddedInts(numbers)
	// 判断条件并进入对应分支逻辑。
	if len(drawResult) == 0 {
		// 更新当前变量或字段值。
		drawResult = "--"
	}

	// 返回当前处理结果。
	return map[string]interface{}{
		// 进入新的代码块进行处理。
		"special_lottery": map[string]interface{}{
			// 处理当前语句逻辑。
			"id": sl.ID,
			// 处理当前语句逻辑。
			"name": sl.Name,
			// 处理当前语句逻辑。
			"code": sl.Code,
			// 处理当前语句逻辑。
			"current_issue": issue,
			// 调用nextDrawAt.Format完成当前处理。
			"next_draw_at": nextDrawAt.Format(time.RFC3339),
		},
		// 进入新的代码块进行处理。
		"live": map[string]interface{}{
			// 处理当前语句逻辑。
			"show_player": showByConfig,
			// 处理当前语句逻辑。
			"has_data": hasLiveData,
			// 处理当前语句逻辑。
			"status": sl.LiveStatus,
			// 处理当前语句逻辑。
			"stream_url": sl.LiveStreamURL,
		},
		// 进入新的代码块进行处理。
		"draw": map[string]interface{}{
			// 处理当前语句逻辑。
			"draw_record_id": drawRecordID,
			// 处理当前语句逻辑。
			"special_lottery_id": sid,
			// 处理当前语句逻辑。
			"issue": issue,
			// 处理当前语句逻辑。
			"draw_at": drawAt,
			// 处理当前语句逻辑。
			"normal_numbers": normalNumbers,
			// 处理当前语句逻辑。
			"special_number": specialNumber,
			// 处理当前语句逻辑。
			"draw_result": drawResult,
			// 处理当前语句逻辑。
			"playback_url": playbackURL,
			// 处理当前语句逻辑。
			"numbers": numbers,
			// 处理当前语句逻辑。
			"labels": labels,
		},
		// 处理当前语句逻辑。
	}, nil
}
