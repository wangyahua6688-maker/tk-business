package server

import (
	"context"

	tkv1 "github.com/wangyahua6688-maker/tk-proto/gen/go/tk/v1"
	"tk-business/internal/svc"
)

// lotteryReadService 定义开奖与图纸读取接口。
type lotteryReadService interface {
	BuildDashboardWithCache(ctx context.Context, sid uint) (map[string]interface{}, error)
	BuildDrawHistoryBySpecialID(specialLotteryID uint, orderMode string, showFive bool, limit int) (map[string]interface{}, error)
	BuildDrawDetail(recordID uint) (map[string]interface{}, error)
	BuildLotteryCardsWithCache(ctx context.Context, category string) ([]map[string]interface{}, error)
	BuildDetail(ctx context.Context, infoID uint) (map[string]interface{}, error)
	BuildHistory(infoID uint) (map[string]interface{}, error)
}

// LotteryReadRPC 负责开奖、图纸读取相关 RPC。
type LotteryReadRPC struct {
	lotteryReadSvc lotteryReadService
}

// LotteryReadRPCDeps 定义开奖读取模块依赖。
type LotteryReadRPCDeps struct {
	LotteryReadService lotteryReadService
}

// NewLotteryReadRPC 根据服务上下文创建开奖读取模块 RPC。
func NewLotteryReadRPC(ctx *svc.ServiceContext) *LotteryReadRPC {
	return NewLotteryReadRPCWithDeps(LotteryReadRPCDeps{
		LotteryReadService: ctx.LotteryService,
	})
}

// NewLotteryReadRPCWithDeps 使用显式依赖创建开奖读取模块 RPC。
func NewLotteryReadRPCWithDeps(deps LotteryReadRPCDeps) *LotteryReadRPC {
	return &LotteryReadRPC{lotteryReadSvc: deps.LotteryReadService}
}

// LotteryDashboard 返回开奖看板数据（使用带缓存版本，防止高峰期缓存击穿）。
func (l *LotteryReadRPC) LotteryDashboard(ctx context.Context, req *tkv1.IDRequest) (*tkv1.LotteryDashboardReply, error) {
	payload, err := l.lotteryReadSvc.BuildDashboardWithCache(ctx, uint(req.GetId()))
	if err != nil {
		return &tkv1.LotteryDashboardReply{Code: 40401, Msg: "special lottery not found"}, nil
	}
	data := &tkv1.LotteryDashboardData{}
	if err := decodePayloadToProto(payload, data); err != nil {
		return &tkv1.LotteryDashboardReply{Code: 50099, Msg: "marshal response failed"}, nil
	}
	return &tkv1.LotteryDashboardReply{Code: 0, Msg: "ok", Data: data}, nil
}

// DrawHistory 返回开奖区历史开奖列表。
func (l *LotteryReadRPC) DrawHistory(_ context.Context, req *tkv1.DrawHistoryRequest) (*tkv1.LotteryHistoryReply, error) {
	payload, err := l.lotteryReadSvc.BuildDrawHistoryBySpecialID(
		uint(req.GetSpecialLotteryId()),
		req.GetOrderMode(),
		req.GetShowFive(),
		int(req.GetLimit()),
	)
	if err != nil {
		return &tkv1.LotteryHistoryReply{Code: 40441, Msg: "draw history not found"}, nil
	}
	data := &tkv1.LotteryHistoryData{}
	if err := decodePayloadToProto(payload, data); err != nil {
		return &tkv1.LotteryHistoryReply{Code: 50099, Msg: "marshal response failed"}, nil
	}
	return &tkv1.LotteryHistoryReply{Code: 0, Msg: "ok", Data: data}, nil
}

// DrawDetail 返回开奖区开奖详情。
func (l *LotteryReadRPC) DrawDetail(_ context.Context, req *tkv1.IDRequest) (*tkv1.LotteryDrawDetailReply, error) {
	payload, err := l.lotteryReadSvc.BuildDrawDetail(uint(req.GetId()))
	if err != nil {
		return &tkv1.LotteryDrawDetailReply{Code: 40442, Msg: "draw detail not found"}, nil
	}
	data := &tkv1.LotteryDrawDetailData{}
	if err := decodePayloadToProto(payload, data); err != nil {
		return &tkv1.LotteryDrawDetailReply{Code: 50099, Msg: "marshal response failed"}, nil
	}
	return &tkv1.LotteryDrawDetailReply{Code: 0, Msg: "ok", Data: data}, nil
}

// ListCards 返回图纸卡片列表（使用带缓存版本）。
func (l *LotteryReadRPC) ListCards(ctx context.Context, req *tkv1.ListCardsRequest) (*tkv1.JsonDataReply, error) {
	// 注：BuildLotteryCardsWithCache 返回 []map[string]interface{}
	// 原版 ListCards 已包含在 Service 中，此处调用带缓存版本
	items, err := l.lotteryReadSvc.BuildLotteryCardsWithCache(ctx, req.GetCategory())
	if err != nil {
		return &tkv1.JsonDataReply{Code: 50011, Msg: "failed to load lottery cards"}, nil
	}
	return marshalOK(map[string]interface{}{"items": items, "current_category": req.GetCategory()})
}

// LotteryDetail 返回彩种详情聚合数据。
func (l *LotteryReadRPC) LotteryDetail(ctx context.Context, req *tkv1.IDRequest) (*tkv1.LotteryDetailReply, error) {
	payload, err := l.lotteryReadSvc.BuildDetail(ctx, uint(req.GetId()))
	if err != nil {
		return &tkv1.LotteryDetailReply{Code: 40411, Msg: "lottery info not found"}, nil
	}
	data := &tkv1.LotteryDetailData{}
	if err := decodePayloadToProto(payload, data); err != nil {
		return &tkv1.LotteryDetailReply{Code: 50099, Msg: "marshal response failed"}, nil
	}
	return &tkv1.LotteryDetailReply{Code: 0, Msg: "ok", Data: data}, nil
}

// LotteryHistory 返回彩种历史开奖数据。
func (l *LotteryReadRPC) LotteryHistory(_ context.Context, req *tkv1.IDRequest) (*tkv1.LotteryHistoryReply, error) {
	payload, err := l.lotteryReadSvc.BuildHistory(uint(req.GetId()))
	if err != nil {
		return &tkv1.LotteryHistoryReply{Code: 40412, Msg: "lottery history not found"}, nil
	}
	data := &tkv1.LotteryHistoryData{}
	if err := decodePayloadToProto(payload, data); err != nil {
		return &tkv1.LotteryHistoryReply{Code: 50099, Msg: "marshal response failed"}, nil
	}
	return &tkv1.LotteryHistoryReply{Code: 0, Msg: "ok", Data: data}, nil
}

// LotteryResults 返回彩种结果聚合数据。
func (l *LotteryReadRPC) LotteryResults(ctx context.Context, req *tkv1.IDRequest) (*tkv1.LotteryDetailReply, error) {
	payload, err := l.lotteryReadSvc.BuildDetail(ctx, uint(req.GetId()))
	if err != nil {
		return &tkv1.LotteryDetailReply{Code: 40413, Msg: "lottery results not found"}, nil
	}
	data := &tkv1.LotteryDetailData{}
	if err := decodePayloadToProto(payload, data); err != nil {
		return &tkv1.LotteryDetailReply{Code: 50099, Msg: "marshal response failed"}, nil
	}
	return &tkv1.LotteryDetailReply{Code: 0, Msg: "ok", Data: data}, nil
}
