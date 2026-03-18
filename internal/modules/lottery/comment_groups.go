package lottery

import (
	"context"

	"tk-business/internal/dao"
	"tk-business/internal/userclient"
)

// loadCommentGroups 获取详情页评论分组数据。
// 读取策略：
// 1. 优先走 user gRPC（便于后续分库分表/缓存统一治理）；
// 2. gRPC 异常时，回退为本地 DB 查询，保证页面不空白。
func (s *Service) loadCommentGroups(ctx context.Context, infoID uint) (userclient.LotteryCommentGroups, error) {
	// 判断条件并进入对应分支逻辑。
	if s.commentClient != nil && s.commentClient.IsEnabled() {
		// 定义并初始化当前变量。
		payload, err := s.commentClient.LotteryCommentGroups(ctx, infoID)
		// 判断条件并进入对应分支逻辑。
		if err == nil {
			// 返回当前处理结果。
			return payload, nil
		}
	}

	// 定义并初始化当前变量。
	systemRows, err := s.dao.ListLotteryComments(infoID, 12, "latest", []string{"official"})
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return userclient.LotteryCommentGroups{}, err
	}
	// 定义并初始化当前变量。
	userRows, err := s.dao.ListLotteryComments(infoID, 12, "latest", []string{"natural", "robot"})
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return userclient.LotteryCommentGroups{}, err
	}
	// 定义并初始化当前变量。
	hotRows, err := s.dao.ListLotteryComments(infoID, 8, "hot", nil)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return userclient.LotteryCommentGroups{}, err
	}
	// 定义并初始化当前变量。
	latestRows, err := s.dao.ListLotteryComments(infoID, 8, "latest", nil)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return userclient.LotteryCommentGroups{}, err
	}
	// 返回当前处理结果。
	return userclient.LotteryCommentGroups{
		// 调用buildCommentPayload完成当前处理。
		SystemComments: buildCommentPayload(systemRows),
		// 调用buildCommentPayload完成当前处理。
		UserComments: buildCommentPayload(userRows),
		// 调用buildCommentPayload完成当前处理。
		HotComments: buildCommentPayload(hotRows),
		// 调用buildCommentPayload完成当前处理。
		LatestComments: buildCommentPayload(latestRows),
		// 处理当前语句逻辑。
	}, nil
}

// buildCommentPayload 将评论聚合行转为前端响应结构。
func buildCommentPayload(rows []dao.LotteryCommentRow) []map[string]interface{} {
	// 定义并初始化当前变量。
	out := make([]map[string]interface{}, 0, len(rows))
	// 循环处理当前数据集合。
	for _, row := range rows {
		// 更新当前变量或字段值。
		out = append(out, map[string]interface{}{
			// 处理当前语句逻辑。
			"id": row.ID,
			// 处理当前语句逻辑。
			"user_id": row.UserID,
			// 处理当前语句逻辑。
			"parent_id": row.ParentID,
			// 处理当前语句逻辑。
			"content": row.Content,
			// 处理当前语句逻辑。
			"likes": row.Likes,
			// 处理当前语句逻辑。
			"created_at": row.CreatedAt,
			// 处理当前语句逻辑。
			"username": row.Username,
			// 处理当前语句逻辑。
			"nickname": row.Nickname,
			// 处理当前语句逻辑。
			"avatar": row.Avatar,
			// 处理当前语句逻辑。
			"user_type": row.UserType,
		})
	}
	// 返回当前处理结果。
	return out
}
