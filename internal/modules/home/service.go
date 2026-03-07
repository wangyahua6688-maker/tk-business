package home

import "tk-business/internal/dao"

// Service 首页领域服务。
// 该服务只负责“首页与分类库”的业务聚合，不直接处理传输协议。
type Service struct {
	dao *dao.HomeDAO
}

// NewService 构建首页服务实例。
func NewService(homeDAO *dao.HomeDAO) *Service {
	return &Service{dao: homeDAO}
}
