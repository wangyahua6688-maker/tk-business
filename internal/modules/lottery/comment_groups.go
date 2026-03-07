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
	if s.commentClient != nil && s.commentClient.IsEnabled() {
		payload, err := s.commentClient.LotteryCommentGroups(ctx, infoID)
		if err == nil {
			return payload, nil
		}
	}

	systemRows, err := s.dao.ListLotteryComments(infoID, 12, "latest", []string{"official"})
	if err != nil {
		return userclient.LotteryCommentGroups{}, err
	}
	userRows, err := s.dao.ListLotteryComments(infoID, 12, "latest", []string{"natural", "robot"})
	if err != nil {
		return userclient.LotteryCommentGroups{}, err
	}
	hotRows, err := s.dao.ListLotteryComments(infoID, 8, "hot", nil)
	if err != nil {
		return userclient.LotteryCommentGroups{}, err
	}
	latestRows, err := s.dao.ListLotteryComments(infoID, 8, "latest", nil)
	if err != nil {
		return userclient.LotteryCommentGroups{}, err
	}
	return userclient.LotteryCommentGroups{
		SystemComments: buildCommentPayload(systemRows),
		UserComments:   buildCommentPayload(userRows),
		HotComments:    buildCommentPayload(hotRows),
		LatestComments: buildCommentPayload(latestRows),
	}, nil
}

// buildCommentPayload 将评论聚合行转为前端响应结构。
func buildCommentPayload(rows []dao.LotteryCommentRow) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		out = append(out, map[string]interface{}{
			"id":         row.ID,
			"user_id":    row.UserID,
			"parent_id":  row.ParentID,
			"content":    row.Content,
			"likes":      row.Likes,
			"created_at": row.CreatedAt,
			"username":   row.Username,
			"nickname":   row.Nickname,
			"avatar":     row.Avatar,
			"user_type":  row.UserType,
		})
	}
	return out
}
