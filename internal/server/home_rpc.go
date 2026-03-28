package server

import (
	"context"

	tkv1 "github.com/wangyahua6688-maker/tk-proto/gen/go/tk/v1"
	"tk-business/internal/svc"
)

// homeOverviewService 定义首页聚合读取接口。
type homeOverviewService interface {
	BuildOverviewWithCache(ctx context.Context) (map[string]interface{}, error)
}

// homeCategoryService 定义分类库读取接口。
type homeCategoryService interface {
	ListCategoryLibrary(keyword string) (map[string]interface{}, error)
}

// HomeRPC 负责首页聚合与分类相关 RPC。
type HomeRPC struct {
	homeOverviewSvc homeOverviewService
	homeCategorySvc homeCategoryService
}

// HomeRPCDeps 定义首页模块依赖。
type HomeRPCDeps struct {
	HomeOverviewService homeOverviewService
	HomeCategoryService homeCategoryService
}

// NewHomeRPC 根据服务上下文创建首页模块 RPC。
func NewHomeRPC(ctx *svc.ServiceContext) *HomeRPC {
	return NewHomeRPCWithDeps(HomeRPCDeps{
		HomeOverviewService: ctx.CachedHomeService,
		HomeCategoryService: ctx.HomeService,
	})
}

// NewHomeRPCWithDeps 使用显式依赖创建首页模块 RPC。
func NewHomeRPCWithDeps(deps HomeRPCDeps) *HomeRPC {
	return &HomeRPC{
		homeOverviewSvc: deps.HomeOverviewService,
		homeCategorySvc: deps.HomeCategoryService,
	}
}

// HomeOverview 首页聚合接口（使用带 Redis 缓存版本）。
// 修复点：原版每次请求都执行 5+ 条 DB 查询（banners/broadcasts/popup/tabs/links/categories），
//
//	改为优先读 Redis 缓存，缓存 miss 时才查 DB 并写入缓存。
func (h *HomeRPC) HomeOverview(ctx context.Context, req *tkv1.HomeOverviewRequest) (*tkv1.JsonDataReply, error) {
	// 使用带缓存的服务方法（CachedHomeService 由 ServiceContext 初始化）
	payload, err := h.homeOverviewSvc.BuildOverviewWithCache(ctx)
	if err != nil {
		return &tkv1.JsonDataReply{Code: 50001, Msg: "home overview unavailable"}, nil
	}
	return marshalOK(payload)
}

// LotteryCategories 图库分类搜索接口。
func (h *HomeRPC) LotteryCategories(_ context.Context, req *tkv1.CategoryLibraryRequest) (*tkv1.JsonDataReply, error) {
	payload, err := h.homeCategorySvc.ListCategoryLibrary(req.GetKeyword())
	if err != nil {
		return &tkv1.JsonDataReply{Code: 50002, Msg: "category library unavailable"}, nil
	}
	return marshalOK(payload)
}
