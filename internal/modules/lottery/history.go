package lottery

// BuildHistory 生成开奖历史列表。
func (s *Service) BuildHistory(infoID uint) (map[string]interface{}, error) {
	// 1) 兼容旧接口：先从图纸 ID 反查所属彩种。
	current, err := s.dao.GetLotteryInfo(infoID)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 2) 历史开奖主数据来自 tk_draw_record（开奖区独立表）。
	payload, err := s.BuildDrawHistoryBySpecialID(current.SpecialLotteryID, "desc", true, 80)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 3) 保留旧字段，避免旧前端联调中断。
	payload["lottery_info_id"] = current.ID
	// 更新当前变量或字段值。
	payload["title"] = current.Title
	// 返回当前处理结果。
	return payload, nil
}
