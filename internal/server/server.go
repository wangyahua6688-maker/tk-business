package server

import (
	"encoding/json"

	"tk-business/internal/svc"
	tkv1 "tk-proto/tk/v1"
)

// BusinessServer 业务域 gRPC 服务实现。
// 该层负责协议适配，不承载核心业务逻辑。
type BusinessServer struct {
	// 处理当前语句逻辑。
	tkv1.UnimplementedBusinessServiceServer
	// 处理当前语句逻辑。
	ctx *svc.ServiceContext
}

// NewBusinessServer 构建业务服务端实例。
func NewBusinessServer(ctx *svc.ServiceContext) *BusinessServer {
	// 返回当前处理结果。
	return &BusinessServer{ctx: ctx}
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
