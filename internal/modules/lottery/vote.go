package lottery

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"tk-common/models"

	"gorm.io/gorm"
)

// Vote 投票接口，内置防刷策略：
// 1. 基于设备指纹限频；
// 2. 基于 (lottery_info_id + voter_hash) 唯一约束防重复。
func (s *Service) Vote(infoID, optionID uint, meta VoteMeta) (map[string]interface{}, error) {
	// 0) 确保图纸存在投票选项（自动补默认项，避免“无可投选项”）。
	if _, err := s.ensurePollOptions(infoID); err != nil {
		// 返回当前处理结果。
		return nil, err
	}

	// 1) 先计算投票指纹。
	voterHash := buildVoterHash(meta)
	// 判断条件并进入对应分支逻辑。
	if voterHash == "" {
		// 返回当前处理结果。
		return nil, fmt.Errorf("invalid voter fingerprint")
	}

	// 2) 先过本地令牌桶限频，挡住高频请求。
	if !s.voteLimiter.Allow(voterHash) {
		// 返回当前处理结果。
		return nil, fmt.Errorf("vote too frequent")
	}

	// 3) 校验选项是否存在且属于当前图纸。
	opt, err := s.dao.FindOption(optionID)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 判断条件并进入对应分支逻辑。
	if opt.LotteryInfoID != infoID {
		// 返回当前处理结果。
		return nil, fmt.Errorf("option not matched to lottery info")
	}

	// 4) 查询是否已经投过票（命中即直接返回）。
	if _, err := s.dao.GetVoteRecord(infoID, voterHash); err == nil {
		// 返回当前处理结果。
		return nil, fmt.Errorf("already voted")
		// 进入新的代码块进行处理。
	} else if err != nil && err != gorm.ErrRecordNotFound {
		// 返回当前处理结果。
		return nil, err
	}

	// 5) 事务写入：插入投票记录 + 选项票数自增。
	err = s.dao.WithTx(func(tx *gorm.DB) error {
		// 定义并初始化当前变量。
		record := &models.WLotteryVoteRecord{
			// 处理当前语句逻辑。
			LotteryInfoID: infoID,
			// 处理当前语句逻辑。
			OptionID: optionID,
			// 处理当前语句逻辑。
			VoterHash: voterHash,
			// 处理当前语句逻辑。
			DeviceID: meta.DeviceID,
			// 处理当前语句逻辑。
			ClientIP: meta.ClientIP,
			// 调用trimLen完成当前处理。
			UserAgent: trimLen(meta.UserAgent, 255),
		}
		// 判断条件并进入对应分支逻辑。
		if err := tx.Create(record).Error; err != nil {
			// 返回当前处理结果。
			return err
		}
		// 判断条件并进入对应分支逻辑。
		if err := tx.Model(&models.WLotteryOption{}).
			// 更新当前变量或字段值。
			Where("id = ?", optionID).
			// 调用UpdateColumn完成当前处理。
			UpdateColumn("votes", gorm.Expr("votes + 1")).Error; err != nil {
			// 返回当前处理结果。
			return err
		}
		// 返回当前处理结果。
		return nil
	})
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 6) 对唯一索引冲突进行统一语义映射。
		lower := strings.ToLower(err.Error())
		// 判断条件并进入对应分支逻辑。
		if strings.Contains(lower, "duplicate") || strings.Contains(lower, "unique") {
			// 返回当前处理结果。
			return nil, fmt.Errorf("already voted")
		}
		// 返回当前处理结果。
		return nil, err
	}

	// 7) 返回提交后的最新投票状态。
	return s.GetVoteRecord(infoID, meta)
}

// GetVoteRecord 查询当前设备对指定图纸的投票状态。
func (s *Service) GetVoteRecord(infoID uint, meta VoteMeta) (map[string]interface{}, error) {
	// 0) 先保证图纸有可投票选项（默认 12 生肖）。
	options, err := s.ensurePollOptions(infoID)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 定义并初始化当前变量。
	totalVotes := int64(0)
	// 循环处理当前数据集合。
	for _, opt := range options {
		// 更新当前变量或字段值。
		totalVotes += opt.Votes
	}

	// 1) 用同一指纹策略读取历史投票。
	voterHash := buildVoterHash(meta)
	// 定义并初始化当前变量。
	record, err := s.dao.GetVoteRecord(infoID, voterHash)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 判断条件并进入对应分支逻辑。
		if err == gorm.ErrRecordNotFound {
			// 未投票时返回固定结构，减少前端判空逻辑。
			return map[string]interface{}{
				// 处理当前语句逻辑。
				"voted": false,
				// 处理当前语句逻辑。
				"my_option_id": 0,
				// 调用buildPollPayload完成当前处理。
				"poll_options": buildPollPayload(options, totalVotes),
				// 处理当前语句逻辑。
			}, nil
		}
		// 返回当前处理结果。
		return nil, err
	}

	// 返回当前处理结果。
	return map[string]interface{}{
		// 处理当前语句逻辑。
		"voted": true,
		// 处理当前语句逻辑。
		"my_option_id": record.OptionID,
		// 调用buildPollPayload完成当前处理。
		"poll_options": buildPollPayload(options, totalVotes),
		// 处理当前语句逻辑。
		"voted_at": record.CreatedAt,
		// 处理当前语句逻辑。
	}, nil
}

// buildVoterHash 计算投票指纹。
// 优先使用 DeviceID，缺失时退化为 IP + UA。
func buildVoterHash(meta VoteMeta) string {
	// 定义并初始化当前变量。
	raw := strings.TrimSpace(meta.DeviceID)
	// 判断条件并进入对应分支逻辑。
	if raw == "" {
		// 更新当前变量或字段值。
		raw = strings.TrimSpace(meta.ClientIP) + "|" + strings.TrimSpace(meta.UserAgent)
	}
	// 判断条件并进入对应分支逻辑。
	if strings.TrimSpace(raw) == "" {
		// 返回当前处理结果。
		return ""
	}
	// 定义并初始化当前变量。
	sum := sha256.Sum256([]byte(raw))
	// 返回当前处理结果。
	return hex.EncodeToString(sum[:])
}

// trimLen 安全截断字符串，避免写入过长字段触发 DB 错误。
func trimLen(v string, max int) string {
	// 判断条件并进入对应分支逻辑。
	if len(v) <= max {
		// 返回当前处理结果。
		return v
	}
	// 返回当前处理结果。
	return v[:max]
}
