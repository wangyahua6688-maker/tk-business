package lottery

// BuildHistory 生成开奖历史列表。
func (s *Service) BuildHistory(infoID uint) (map[string]interface{}, error) {
	// 1) 兼容旧接口：先从图纸 ID 反查所属彩种。
	current, err := s.dao.GetLotteryInfo(infoID)
	if err != nil {
		return nil, err
	}
	// 2) 历史开奖主数据来自 tk_draw_record（开奖区独立表）。
	payload, err := s.BuildDrawHistoryBySpecialID(current.SpecialLotteryID, "desc", true, 80)
	if err != nil {
		return nil, err
	}
	// 3) 保留旧字段，避免旧前端联调中断。
	payload["lottery_info_id"] = current.ID
	payload["title"] = current.Title
	return payload, nil
}
