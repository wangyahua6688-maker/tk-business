package svc

import (
	"context"
	"strings"
	"time"

	redisx "github.com/wangyahua6688-maker/tk-common/utils/redisx/v8"
	"tk-business/internal/config"
	"tk-business/internal/dao"
	homeModule "tk-business/internal/modules/home"
	lotteryModule "tk-business/internal/modules/lottery"
	"tk-business/internal/platform/database"
	"tk-business/internal/userclient"

	"github.com/go-redis/redis/v8"
)

// ServiceContext 业务服务上下文。
type ServiceContext struct {
	// Config 保存启动配置，便于下游模块读取限流/开关参数。
	Config config.Config
	// HomeCore 负责首页聚合与分类库能力（无缓存，供内部穿透调用）。
	HomeCore *homeModule.Service
	// CachedHomeCore 带 Redis 缓存的首页服务（RPC 层优先使用此版本）。
	CachedHomeCore *homeModule.CachedService
	// LotteryCore 负责开奖、图纸详情、投票、现场页等核心业务能力。
	LotteryCore *lotteryModule.Service
}

// NewServiceContext 创建 ServiceContext 实例。
func NewServiceContext(c config.Config) (*ServiceContext, error) {
	// 1) 初始化数据库连接，供 DAO 层复用。
	db, err := database.NewMySQL(c.Database.DSN)
	if err != nil {
		return nil, err
	}

	// 2) 初始化首页 DAO 与首页聚合服务（基础版，无缓存）。
	homeDAO := dao.NewHomeDAO(db)
	homeCore := homeModule.NewService(homeDAO)

	// 3) 初始化开奖 DAO 与用户域（论坛）RPC 客户端。
	lotteryDAO := dao.NewLotteryDAO(db)
	userClient := userclient.New(c.UserRpc)

	// 4) 初始化 Redis 缓存客户端（可选，地址为空时退化为无缓存模式）。
	var redisClient *redis.Client
	if strings.TrimSpace(c.CacheRedis.Addr) != "" {
		redisCfg := redisx.DefaultConfig()
		redisCfg.Addr = strings.TrimSpace(c.CacheRedis.Addr)
		redisCfg.Password = c.CacheRedis.Password
		redisCfg.DB = c.CacheRedis.DB
		cli, redisErr := redisx.NewClient(context.Background(), redisCfg)
		if redisErr == nil {
			redisClient = cli
		}
		// Redis 连接失败时仅记录，不阻断启动（降级为直接查 DB）
	}
	sceneTTL := time.Duration(c.CacheRedis.SceneTTLSeconds) * time.Second

	// 5) 初始化带缓存的首页服务（RPC 层优先使用此版本）。
	// overviewTTL 使用与 scene 相同的配置入口；如需独立配置可扩展 config.go。
	overviewTTL := 5 * time.Minute // 首页聚合数据缓存默认 5 分钟
	cachedHomeCore := homeModule.NewCachedService(homeCore, redisClient, overviewTTL)

	// 6) 初始化开奖服务（含投票防刷 + Redis 现场缓存）。
	lotteryCore := lotteryModule.NewService(lotteryDAO, userClient, redisClient, sceneTTL)

	// 7) 将全部核心模块注入上下文。
	return &ServiceContext{
		Config:         c,
		HomeCore:       homeCore,
		CachedHomeCore: cachedHomeCore,
		LotteryCore:    lotteryCore,
	}, nil
}
