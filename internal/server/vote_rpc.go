package server

import (
	"context"

	tkv1 "github.com/wangyahua6688-maker/tk-proto/gen/go/tk/v1"
	lotteryModule "tk-business/internal/modules/lottery"
	"tk-business/internal/svc"
)

// voteService 定义投票接口。
type voteService interface {
	GetVoteRecord(infoID uint, meta lotteryModule.VoteMeta) (map[string]interface{}, error)
	Vote(infoID, optionID uint, meta lotteryModule.VoteMeta) (map[string]interface{}, error)
}

// VoteRPC 负责投票相关 RPC。
type VoteRPC struct {
	voteSvc voteService
}

// VoteRPCDeps 定义投票模块依赖。
type VoteRPCDeps struct {
	VoteService voteService
}

// NewVoteRPC 根据服务上下文创建投票模块 RPC。
func NewVoteRPC(ctx *svc.ServiceContext) *VoteRPC {
	return NewVoteRPCWithDeps(VoteRPCDeps{
		VoteService: ctx.LotteryService,
	})
}

// NewVoteRPCWithDeps 使用显式依赖创建投票模块 RPC。
func NewVoteRPCWithDeps(deps VoteRPCDeps) *VoteRPC {
	return &VoteRPC{voteSvc: deps.VoteService}
}

// VoteRecord 查询当前请求端在图纸下的投票状态。
func (l *VoteRPC) VoteRecord(_ context.Context, req *tkv1.VoteRecordRequest) (*tkv1.JsonDataReply, error) {
	// 将 RPC 请求透传到投票领域服务，统一使用设备元信息识别请求端。
	payload, err := l.voteSvc.GetVoteRecord(uint(req.GetLotteryInfoId()), lotteryModule.VoteMeta{
		// 调用req.GetDeviceId完成当前处理。
		DeviceID: req.GetDeviceId(),
		// 调用req.GetClientIp完成当前处理。
		ClientIP: req.GetClientIp(),
		// 调用req.GetUserAgent完成当前处理。
		UserAgent: req.GetUserAgent(),
	})
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 业务错误统一映射为 code/msg，RPC 层不抛 gRPC error。
		return &tkv1.JsonDataReply{Code: 50041, Msg: "failed to load vote record"}, nil
	}
	// 成功时返回标准 JSON 包装结构。
	return marshalOK(payload)
}

// Vote 提交投票并返回更新后的投票结果。
func (l *VoteRPC) Vote(_ context.Context, req *tkv1.VoteRequest) (*tkv1.JsonDataReply, error) {
	// 执行投票：包含防刷校验、重复投票校验、票数落库与统计计算。
	payload, err := l.voteSvc.Vote(uint(req.GetLotteryInfoId()), uint(req.GetOptionId()), lotteryModule.VoteMeta{
		// 调用req.GetDeviceId完成当前处理。
		DeviceID: req.GetDeviceId(),
		// 调用req.GetClientIp完成当前处理。
		ClientIP: req.GetClientIp(),
		// 调用req.GetUserAgent完成当前处理。
		UserAgent: req.GetUserAgent(),
	})
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 将业务错误细分为可识别的前端错误码，便于 UI 提示。
		switch err.Error() {
		case "vote too frequent":
			// 返回当前处理结果。
			return &tkv1.JsonDataReply{Code: 42931, Msg: "vote too frequent"}, nil
		case "already voted":
			// 返回当前处理结果。
			return &tkv1.JsonDataReply{Code: 40033, Msg: "already voted"}, nil
		default:
			// 返回当前处理结果。
			return &tkv1.JsonDataReply{Code: 50031, Msg: "vote failed"}, nil
		}
	}
	// 投票成功后返回最新结果，前端可直接刷新图表。
	return marshalOK(payload)
}
