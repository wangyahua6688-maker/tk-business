package server

import (
	"context"

	tkv1 "tk-proto/tk/v1"
)

// HomeOverview 返回客户端首页整页聚合数据。
// 该接口统一由 tk-business 提供，避免拆分过细导致跨服务调用链过长。
func (s *BusinessServer) HomeOverview(_ context.Context, _ *tkv1.HomeOverviewRequest) (*tkv1.JsonDataReply, error) {
	// 1) 调用首页聚合模块查询 banner/广播/分类等基础数据。
	payload, err := s.ctx.HomeCore.BuildOverview()
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 2) 将内部错误映射为业务码，避免把底层错误暴露到外部。
		return &tkv1.JsonDataReply{Code: 50001, Msg: "failed to build home overview"}, nil
	}
	// 3) 使用统一序列化函数输出标准响应。
	return marshalOK(payload)
}

// LotteryCategories 返回图库分类搜索结果。
func (s *BusinessServer) LotteryCategories(_ context.Context, req *tkv1.CategoryLibraryRequest) (*tkv1.JsonDataReply, error) {
	// 1) 传入关键字进行模糊筛选。
	payload, err := s.ctx.HomeCore.ListCategoryLibrary(req.GetKeyword())
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 2) 失败返回业务码，BFF 层会原样透传。
		return &tkv1.JsonDataReply{Code: 50002, Msg: "failed to load lottery categories"}, nil
	}
	// 3) 成功统一输出 data_json。
	return marshalOK(payload)
}
