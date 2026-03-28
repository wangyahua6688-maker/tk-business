package server

import (
	"context"
	"encoding/json"

	tkv1 "github.com/wangyahua6688-maker/tk-proto/gen/go/tk/v1"
	"tk-business/internal/svc"
)

// BusinessServer 业务域 gRPC 服务实现。
// 该层负责协议适配，不承载核心业务逻辑。
type BusinessServer struct {
	// 处理当前语句逻辑。
	tkv1.UnimplementedBusinessServiceServer
	homeRPC        *HomeRPC
	lotteryReadRPC *LotteryReadRPC
	voteRPC        *VoteRPC
	liveSceneRPC   *LiveSceneRPC
}

// NewBusinessServer 构建业务服务端实例。
func NewBusinessServer(ctx *svc.ServiceContext) *BusinessServer {
	return &BusinessServer{
		homeRPC:        NewHomeRPC(ctx),
		lotteryReadRPC: NewLotteryReadRPC(ctx),
		voteRPC:        NewVoteRPC(ctx),
		liveSceneRPC:   NewLiveSceneRPC(ctx),
	}
}

// HomeOverview 转发到首页模块 RPC。
func (s *BusinessServer) HomeOverview(ctx context.Context, req *tkv1.HomeOverviewRequest) (*tkv1.JsonDataReply, error) {
	return s.homeRPC.HomeOverview(ctx, req)
}

// LotteryCategories 转发到首页模块 RPC。
func (s *BusinessServer) LotteryCategories(ctx context.Context, req *tkv1.CategoryLibraryRequest) (*tkv1.JsonDataReply, error) {
	return s.homeRPC.LotteryCategories(ctx, req)
}

// LiveScenePage 转发到彩票模块 RPC。
func (s *BusinessServer) LiveScenePage(ctx context.Context, req *tkv1.IDRequest) (*tkv1.JsonDataReply, error) {
	return s.liveSceneRPC.LiveScenePage(ctx, req)
}

// LotteryDashboard 转发到彩票模块 RPC。
func (s *BusinessServer) LotteryDashboard(ctx context.Context, req *tkv1.IDRequest) (*tkv1.LotteryDashboardReply, error) {
	return s.lotteryReadRPC.LotteryDashboard(ctx, req)
}

// DrawHistory 转发到彩票模块 RPC。
func (s *BusinessServer) DrawHistory(ctx context.Context, req *tkv1.DrawHistoryRequest) (*tkv1.LotteryHistoryReply, error) {
	return s.lotteryReadRPC.DrawHistory(ctx, req)
}

// DrawDetail 转发到彩票模块 RPC。
func (s *BusinessServer) DrawDetail(ctx context.Context, req *tkv1.IDRequest) (*tkv1.LotteryDrawDetailReply, error) {
	return s.lotteryReadRPC.DrawDetail(ctx, req)
}

// ListCards 转发到彩票模块 RPC。
func (s *BusinessServer) ListCards(ctx context.Context, req *tkv1.ListCardsRequest) (*tkv1.JsonDataReply, error) {
	return s.lotteryReadRPC.ListCards(ctx, req)
}

// LotteryDetail 转发到彩票模块 RPC。
func (s *BusinessServer) LotteryDetail(ctx context.Context, req *tkv1.IDRequest) (*tkv1.LotteryDetailReply, error) {
	return s.lotteryReadRPC.LotteryDetail(ctx, req)
}

// LotteryHistory 转发到彩票读取模块 RPC。
func (s *BusinessServer) LotteryHistory(ctx context.Context, req *tkv1.IDRequest) (*tkv1.LotteryHistoryReply, error) {
	return s.lotteryReadRPC.LotteryHistory(ctx, req)
}

// LotteryResults 转发到彩票读取模块 RPC。
func (s *BusinessServer) LotteryResults(ctx context.Context, req *tkv1.IDRequest) (*tkv1.LotteryDetailReply, error) {
	return s.lotteryReadRPC.LotteryResults(ctx, req)
}

// VoteRecord 转发到彩票模块 RPC。
func (s *BusinessServer) VoteRecord(ctx context.Context, req *tkv1.VoteRecordRequest) (*tkv1.JsonDataReply, error) {
	return s.voteRPC.VoteRecord(ctx, req)
}

// Vote 转发到彩票模块 RPC。
func (s *BusinessServer) Vote(ctx context.Context, req *tkv1.VoteRequest) (*tkv1.JsonDataReply, error) {
	return s.voteRPC.Vote(ctx, req)
}

// marshalOK 将业务 payload 序列化为统一 gRPC 响应。
func marshalOK(payload interface{}) (*tkv1.JsonDataReply, error) {
	// 定义并初始化当前变量。
	raw, err := json.Marshal(payload)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return &tkv1.JsonDataReply{Code: 50099, Msg: "marshal response failed"}, nil
	}
	// 返回当前处理结果。
	return &tkv1.JsonDataReply{Code: 0, Msg: "ok", DataJson: string(raw)}, nil
}
