package lottery

import common_model "tk-common/models"

// defaultPollOptionNames 返回系统默认的生肖投票项。
// 说明：
// - 当某个图纸尚未配置任何投票项时，自动补齐该默认集合；
// - 与当前业务的“生肖竞猜”一致，固定 12 项。
func defaultPollOptionNames() []string {
	// 返回当前处理结果。
	return []string{"鼠", "牛", "虎", "兔", "龙", "蛇", "马", "羊", "猴", "鸡", "狗", "猪"}
}

// ensurePollOptions 确保某个图纸存在可投票选项。
// 策略：
// 1) 先查现有选项；
// 2) 若为空，自动补默认 12 生肖；
// 3) 再次读取并返回。
func (s *Service) ensurePollOptions(infoID uint) ([]common_model.WLotteryOption, error) {
	// 定义并初始化当前变量。
	options, err := s.dao.ListOptionsByLotteryInfoID(infoID)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 判断条件并进入对应分支逻辑。
	if len(options) > 0 {
		// 返回当前处理结果。
		return options, nil
	}

	// 判断条件并进入对应分支逻辑。
	if err := s.dao.CreateMissingOptions(infoID, defaultPollOptionNames()); err != nil {
		// 返回当前处理结果。
		return nil, err
	}

	// 返回当前处理结果。
	return s.dao.ListOptionsByLotteryInfoID(infoID)
}
