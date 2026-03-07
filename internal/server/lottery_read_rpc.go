package server

import (
	"context"

	tkv1 "tk-proto/tk/v1"
)

// LotteryDashboard 返回开奖看板数据（首页开奖区与开奖现场共用）。
func (s *BusinessServer) LotteryDashboard(_ context.Context, req *tkv1.IDRequest) (*tkv1.JsonDataReply, error) {
	payload, err := s.ctx.LotteryCore.BuildDashboard(uint(req.GetId()))
	if err != nil {
		return &tkv1.JsonDataReply{Code: 40401, Msg: "special lottery not found"}, nil
	}
	return marshalOK(payload)
}

// DrawHistory 返回开奖区历史开奖列表（按彩种维度）。
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

// DrawDetail 返回开奖区开奖详情（图5结构）。
func (s *BusinessServer) DrawDetail(_ context.Context, req *tkv1.IDRequest) (*tkv1.JsonDataReply, error) {
	payload, err := s.ctx.LotteryCore.BuildDrawDetail(uint(req.GetId()))
	if err != nil {
		return &tkv1.JsonDataReply{Code: 40442, Msg: "draw detail not found"}, nil
	}
	return marshalOK(payload)
}

// ListCards 返回彩种图卡列表。
func (s *BusinessServer) ListCards(_ context.Context, req *tkv1.ListCardsRequest) (*tkv1.JsonDataReply, error) {
	items, err := s.ctx.LotteryCore.ListCards(req.GetCategory())
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

// LotteryHistory 返回历史开奖记录。
func (s *BusinessServer) LotteryHistory(_ context.Context, req *tkv1.IDRequest) (*tkv1.JsonDataReply, error) {
	payload, err := s.ctx.LotteryCore.BuildHistory(uint(req.GetId()))
	if err != nil {
		return &tkv1.JsonDataReply{Code: 40451, Msg: "lottery info not found"}, nil
	}
	return marshalOK(payload)
}

// LotteryResults 保持与 LotteryDetail 相同结构，兼容前端旧调用。
func (s *BusinessServer) LotteryResults(ctx context.Context, req *tkv1.IDRequest) (*tkv1.JsonDataReply, error) {
	return s.LotteryDetail(ctx, req)
}
