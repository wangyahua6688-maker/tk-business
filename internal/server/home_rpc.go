package server

import (
	"context"

	tkv1 "github.com/wangyahua6688-maker/tk-proto/gen/go/tk/v1"
)

// HomeOverview 首页聚合接口（使用带 Redis 缓存版本）。
// 修复点：原版每次请求都执行 5+ 条 DB 查询（banners/broadcasts/popup/tabs/links/categories），
//
//	改为优先读 Redis 缓存，缓存 miss 时才查 DB 并写入缓存。
func (s *BusinessServer) HomeOverview(ctx context.Context, req *tkv1.HomeOverviewRequest) (*tkv1.JsonDataReply, error) {
	// 使用带缓存的服务方法（CachedHomeCore 由 ServiceContext 初始化）
	payload, err := s.ctx.CachedHomeCore.BuildOverviewWithCache(ctx)
	if err != nil {
		return &tkv1.JsonDataReply{Code: 50001, Msg: "home overview unavailable"}, nil
	}
	return marshalOK(payload)
}

// LotteryCategories 图库分类搜索接口。
func (s *BusinessServer) LotteryCategories(_ context.Context, req *tkv1.CategoryLibraryRequest) (*tkv1.JsonDataReply, error) {
	payload, err := s.ctx.HomeCore.ListCategoryLibrary(req.GetKeyword())
	if err != nil {
		return &tkv1.JsonDataReply{Code: 50002, Msg: "category library unavailable"}, nil
	}
	return marshalOK(payload)
}
