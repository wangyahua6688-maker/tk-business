package server

import (
	"encoding/json"

	"tk-business/internal/svc"
	tkv1 "tk-proto/tk/v1"
)

// BusinessServer 业务域 gRPC 服务实现。
// 该层负责协议适配，不承载核心业务逻辑。
type BusinessServer struct {
	tkv1.UnimplementedBusinessServiceServer
	ctx *svc.ServiceContext
}

// NewBusinessServer 构建业务服务端实例。
func NewBusinessServer(ctx *svc.ServiceContext) *BusinessServer {
	return &BusinessServer{ctx: ctx}
}

// marshalOK 将业务 payload 序列化为统一 gRPC 响应。
func marshalOK(payload interface{}) (*tkv1.JsonDataReply, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return &tkv1.JsonDataReply{Code: 50099, Msg: "marshal response failed"}, nil
	}
	return &tkv1.JsonDataReply{Code: 0, Msg: "ok", DataJson: string(raw)}, nil
}
