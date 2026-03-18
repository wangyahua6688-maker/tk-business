package server

import (
	"context"

	tkv1 "tk-proto/tk/v1"
)

// LotteryDashboard 返回开奖看板数据（首页开奖区与开奖现场共用）。
func (s *BusinessServer) LotteryDashboard(_ context.Context, req *tkv1.IDRequest) (*tkv1.JsonDataReply, error) {
	// 定义并初始化当前变量。
	payload, err := s.ctx.LotteryCore.BuildDashboard(uint(req.GetId()))
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return &tkv1.JsonDataReply{Code: 40401, Msg: "special lottery not found"}, nil
	}
	// 返回当前处理结果。
	return marshalOK(payload)
}

// DrawHistory 返回开奖区历史开奖列表（按彩种维度）。
func (s *BusinessServer) DrawHistory(_ context.Context, req *tkv1.DrawHistoryRequest) (*tkv1.JsonDataReply, error) {
	// 定义并初始化当前变量。
	payload, err := s.ctx.LotteryCore.BuildDrawHistoryBySpecialID(
		// 调用uint完成当前处理。
		uint(req.GetSpecialLotteryId()),
		// 调用req.GetOrderMode完成当前处理。
		req.GetOrderMode(),
		// 调用req.GetShowFive完成当前处理。
		req.GetShowFive(),
		// 调用int完成当前处理。
		int(req.GetLimit()),
	)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return &tkv1.JsonDataReply{Code: 40441, Msg: "draw history not found"}, nil
	}
	// 返回当前处理结果。
	return marshalOK(payload)
}

// DrawDetail 返回开奖区开奖详情（图5结构）。
func (s *BusinessServer) DrawDetail(_ context.Context, req *tkv1.IDRequest) (*tkv1.JsonDataReply, error) {
	// 定义并初始化当前变量。
	payload, err := s.ctx.LotteryCore.BuildDrawDetail(uint(req.GetId()))
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return &tkv1.JsonDataReply{Code: 40442, Msg: "draw detail not found"}, nil
	}
	// 返回当前处理结果。
	return marshalOK(payload)
}

// ListCards 返回彩种图卡列表。
func (s *BusinessServer) ListCards(_ context.Context, req *tkv1.ListCardsRequest) (*tkv1.JsonDataReply, error) {
	// 定义并初始化当前变量。
	items, err := s.ctx.LotteryCore.ListCards(req.GetCategory())
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return &tkv1.JsonDataReply{Code: 50011, Msg: "failed to load lottery cards"}, nil
	}
	// 返回当前处理结果。
	return marshalOK(map[string]interface{}{"items": items, "current_category": req.GetCategory()})
}

// LotteryDetail 返回彩种详情聚合数据。
func (s *BusinessServer) LotteryDetail(ctx context.Context, req *tkv1.IDRequest) (*tkv1.JsonDataReply, error) {
	// 定义并初始化当前变量。
	payload, err := s.ctx.LotteryCore.BuildDetail(ctx, uint(req.GetId()))
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return &tkv1.JsonDataReply{Code: 40411, Msg: "lottery info not found"}, nil
	}
	// 返回当前处理结果。
	return marshalOK(payload)
}

// LotteryHistory 返回历史开奖记录。
func (s *BusinessServer) LotteryHistory(_ context.Context, req *tkv1.IDRequest) (*tkv1.JsonDataReply, error) {
	// 定义并初始化当前变量。
	payload, err := s.ctx.LotteryCore.BuildHistory(uint(req.GetId()))
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return &tkv1.JsonDataReply{Code: 40451, Msg: "lottery info not found"}, nil
	}
	// 返回当前处理结果。
	return marshalOK(payload)
}

// LotteryResults 保持与 LotteryDetail 相同结构，兼容前端旧调用。
func (s *BusinessServer) LotteryResults(ctx context.Context, req *tkv1.IDRequest) (*tkv1.JsonDataReply, error) {
	// 返回当前处理结果。
	return s.LotteryDetail(ctx, req)
}
