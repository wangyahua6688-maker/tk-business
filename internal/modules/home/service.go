package home

import "tk-business/internal/dao"

// Service 首页领域服务。
// 该服务只负责“首页与分类库”的业务聚合，不直接处理传输协议。
type Service struct {
	// 处理当前语句逻辑。
	dao *dao.HomeDAO
}

// NewService 构建首页服务实例。
func NewService(homeDAO *dao.HomeDAO) *Service {
	// 返回当前处理结果。
	return &Service{dao: homeDAO}
}
