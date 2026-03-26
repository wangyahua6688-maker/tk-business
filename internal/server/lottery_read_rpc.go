package server

import (
	"context"

	tkv1 "tk-proto/gen/go/tk/v1"
)

// LotteryDashboard 返回开奖看板数据（使用带缓存版本，防止高峰期缓存击穿）。
// 修复点：原版直接调用 BuildDashboard（每次查 DB），
//
//	改为调用 BuildDashboardWithCache（Redis 缓存 + 互斥锁防击穿）。
func (s *BusinessServer) LotteryDashboard(ctx context.Context, req *tkv1.IDRequest) (*tkv1.JsonDataReply, error) {
	// 使用带缓存版本替代直接查 DB
	payload, err := s.ctx.LotteryCore.BuildDashboardWithCache(ctx, uint(req.GetId()))
	if err != nil {
		return &tkv1.JsonDataReply{Code: 40401, Msg: "special lottery not found"}, nil
	}
	return marshalOK(payload)
}

// DrawHistory 返回开奖区历史开奖列表。
func (s *BusinessServer) DrawHistory(_ context.Context, req *tkv1.DrawHistoryRequest) (*tkv1.JsonDataReply, error) {
	payload, err := s.ctx.LotteryCore.BuildDrawHistoryBySpecialID(
		uint(req.GetSpecialLotteryId()),
		req.GetOrderMode(),
		req.GetShowFive(),
		int(req.GetLimit()),
	)
	if err != nil {
		return &tkv1.JsonDataReply{Code: 40441, Msg: "draw history not found"}, nil
	}
	return marshalOK(payload)
}

// DrawDetail 返回开奖区开奖详情。
func (s *BusinessServer) DrawDetail(_ context.Context, req *tkv1.IDRequest) (*tkv1.JsonDataReply, error) {
	payload, err := s.ctx.LotteryCore.BuildDrawDetail(uint(req.GetId()))
	if err != nil {
		return &tkv1.JsonDataReply{Code: 40442, Msg: "draw detail not found"}, nil
	}
	return marshalOK(payload)
}

// ListCards 返回图纸卡片列表（使用带缓存版本）。
func (s *BusinessServer) ListCards(_ context.Context, req *tkv1.ListCardsRequest) (*tkv1.JsonDataReply, error) {
	// 注：BuildLotteryCardsWithCache 返回 []map[string]interface{}
	// 原版 ListCards 已包含在 Service 中，此处调用带缓存版本
	ctx := context.Background()
	items, err := s.ctx.LotteryCore.BuildLotteryCardsWithCache(ctx, req.GetCategory())
	if err != nil {
		return &tkv1.JsonDataReply{Code: 50011, Msg: "failed to load lottery cards"}, nil
	}
	return marshalOK(map[string]interface{}{"items": items, "current_category": req.GetCategory()})
}

// LotteryDetail 返回彩种详情聚合数据。
func (s *BusinessServer) LotteryDetail(ctx context.Context, req *tkv1.IDRequest) (*tkv1.JsonDataReply, error) {
	payload, err := s.ctx.LotteryCore.BuildDetail(ctx, uint(req.GetId()))
	if err != nil {
		return &tkv1.JsonDataReply{Code: 40411, Msg: "lottery info not found"}, nil
	}
	return marshalOK(payload)
}
