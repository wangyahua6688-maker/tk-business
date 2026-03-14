package lottery

import "tk-common/models"

// defaultPollOptionNames 返回系统默认的生肖投票项。
// 说明：
// - 当某个图纸尚未配置任何投票项时，自动补齐该默认集合；
// - 与当前业务的“生肖竞猜”一致，固定 12 项。
func defaultPollOptionNames() []string {
	return []string{"鼠", "牛", "虎", "兔", "龙", "蛇", "马", "羊", "猴", "鸡", "狗", "猪"}
}

// ensurePollOptions 确保某个图纸存在可投票选项。
// 策略：
// 1) 先查现有选项；
// 2) 若为空，自动补默认 12 生肖；
// 3) 再次读取并返回。
func (s *Service) ensurePollOptions(infoID uint) ([]models.WLotteryOption, error) {
	options, err := s.dao.ListOptionsByLotteryInfoID(infoID)
	if err != nil {
		return nil, err
	}
	if len(options) > 0 {
		return options, nil
	}

	if err := s.dao.CreateMissingOptions(infoID, defaultPollOptionNames()); err != nil {
		return nil, err
	}

	return s.dao.ListOptionsByLotteryInfoID(infoID)
}
