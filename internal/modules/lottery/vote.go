package lottery

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"tk-shared/models"

	"gorm.io/gorm"
)

// Vote 投票接口，内置防刷策略：
// 1. 基于设备指纹限频；
// 2. 基于 (lottery_info_id + voter_hash) 唯一约束防重复。
func (s *Service) Vote(infoID, optionID uint, meta VoteMeta) (map[string]interface{}, error) {
	// 0) 确保图纸存在投票选项（自动补默认项，避免“无可投选项”）。
	if _, err := s.ensurePollOptions(infoID); err != nil {
		return nil, err
	}

	// 1) 先计算投票指纹。
	voterHash := buildVoterHash(meta)
	if voterHash == "" {
		return nil, fmt.Errorf("invalid voter fingerprint")
	}

	// 2) 先过本地令牌桶限频，挡住高频请求。
	if !s.voteLimiter.Allow(voterHash) {
		return nil, fmt.Errorf("vote too frequent")
	}

	// 3) 校验选项是否存在且属于当前图纸。
	opt, err := s.dao.FindOption(optionID)
	if err != nil {
		return nil, err
	}
	if opt.LotteryInfoID != infoID {
		return nil, fmt.Errorf("option not matched to lottery info")
	}

	// 4) 查询是否已经投过票（命中即直接返回）。
	if _, err := s.dao.GetVoteRecord(infoID, voterHash); err == nil {
		return nil, fmt.Errorf("already voted")
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// 5) 事务写入：插入投票记录 + 选项票数自增。
	err = s.dao.WithTx(func(tx *gorm.DB) error {
		record := &models.WLotteryVoteRecord{
			LotteryInfoID: infoID,
			OptionID:      optionID,
			VoterHash:     voterHash,
			DeviceID:      meta.DeviceID,
			ClientIP:      meta.ClientIP,
			UserAgent:     trimLen(meta.UserAgent, 255),
		}
		if err := tx.Create(record).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.WLotteryOption{}).
			Where("id = ?", optionID).
			UpdateColumn("votes", gorm.Expr("votes + 1")).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		// 6) 对唯一索引冲突进行统一语义映射。
		lower := strings.ToLower(err.Error())
		if strings.Contains(lower, "duplicate") || strings.Contains(lower, "unique") {
			return nil, fmt.Errorf("already voted")
		}
		return nil, err
	}

	// 7) 返回提交后的最新投票状态。
	return s.GetVoteRecord(infoID, meta)
}

// GetVoteRecord 查询当前设备对指定图纸的投票状态。
func (s *Service) GetVoteRecord(infoID uint, meta VoteMeta) (map[string]interface{}, error) {
	// 0) 先保证图纸有可投票选项（默认 12 生肖）。
	options, err := s.ensurePollOptions(infoID)
	if err != nil {
		return nil, err
	}
	totalVotes := int64(0)
	for _, opt := range options {
		totalVotes += opt.Votes
	}

	// 1) 用同一指纹策略读取历史投票。
	voterHash := buildVoterHash(meta)
	record, err := s.dao.GetVoteRecord(infoID, voterHash)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 未投票时返回固定结构，减少前端判空逻辑。
			return map[string]interface{}{
				"voted":        false,
				"my_option_id": 0,
				"poll_options": buildPollPayload(options, totalVotes),
			}, nil
		}
		return nil, err
	}

	return map[string]interface{}{
		"voted":        true,
		"my_option_id": record.OptionID,
		"poll_options": buildPollPayload(options, totalVotes),
		"voted_at":     record.CreatedAt,
	}, nil
}

// buildVoterHash 计算投票指纹。
// 优先使用 DeviceID，缺失时退化为 IP + UA。
func buildVoterHash(meta VoteMeta) string {
	raw := strings.TrimSpace(meta.DeviceID)
	if raw == "" {
		raw = strings.TrimSpace(meta.ClientIP) + "|" + strings.TrimSpace(meta.UserAgent)
	}
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// trimLen 安全截断字符串，避免写入过长字段触发 DB 错误。
func trimLen(v string, max int) string {
	if len(v) <= max {
		return v
	}
	return v[:max]
}
