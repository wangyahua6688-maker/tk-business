package server

import (
	"context"

	lotteryModule "tk-business/internal/modules/lottery"
	tkv1 "tk-proto/tk/v1"
)

// VoteRecord 查询当前请求端在图纸下的投票状态。
func (s *BusinessServer) VoteRecord(_ context.Context, req *tkv1.VoteRecordRequest) (*tkv1.JsonDataReply, error) {
	// 将 RPC 请求透传到投票领域服务，统一使用设备元信息识别请求端。
	payload, err := s.ctx.LotteryCore.GetVoteRecord(uint(req.GetLotteryInfoId()), lotteryModule.VoteMeta{
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
func (s *BusinessServer) Vote(_ context.Context, req *tkv1.VoteRequest) (*tkv1.JsonDataReply, error) {
	// 执行投票：包含防刷校验、重复投票校验、票数落库与统计计算。
	payload, err := s.ctx.LotteryCore.Vote(uint(req.GetLotteryInfoId()), uint(req.GetOptionId()), lotteryModule.VoteMeta{
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
