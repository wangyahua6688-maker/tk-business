package server

import (
	"context"

	tkv1 "tk-proto/tk/v1"
)

// LiveScenePage 返回“开奖现场”整页所需数据。
// 约定：
// - req.id = special_lottery_id（可选，0 表示后端自动选择默认彩种）；
// - 前端改为单接口调用，减少瀑布请求。
func (s *BusinessServer) LiveScenePage(_ context.Context, req *tkv1.IDRequest) (*tkv1.JsonDataReply, error) {
	// 1) 读取请求中的彩种 ID。
	sid := uint(req.GetId())
	// 2) 调用业务层聚合整页数据。
	payload, err := s.ctx.LotteryCore.BuildLiveScenePage(sid)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 3) 聚合失败时返回统一业务错误码，便于前端提示。
		return &tkv1.JsonDataReply{Code: 50061, Msg: "failed to build live scene page"}, nil
	}
	// 4) 序列化后返回。
	return marshalOK(payload)
}
