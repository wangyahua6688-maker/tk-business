package svc

import (
	"strings"
	"time"

	"tk-business/internal/config"
	"tk-business/internal/dao"
	homeModule "tk-business/internal/modules/home"
	lotteryModule "tk-business/internal/modules/lottery"
	"tk-business/internal/platform/database"
	"tk-business/internal/userclient"

	"github.com/redis/go-redis/v9"
)

// ServiceContext 业务服务上下文。
type ServiceContext struct {
	// Config 保存启动配置，便于下游模块读取限流/开关参数。
	Config config.Config
	// HomeCore 负责首页聚合与分类库能力。
	HomeCore *homeModule.Service
	// LotteryCore 负责开奖、图纸详情、投票、现场页等核心业务能力。
	LotteryCore *lotteryModule.Service
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	// 1) 初始化数据库连接，供 DAO 层复用。
	db, err := database.NewMySQL(c.Database.DSN)
	if err != nil {
		return nil, err
	}

	// 2) 初始化首页 DAO 与首页聚合服务。
	homeDAO := dao.NewHomeDAO(db)
	homeCore := homeModule.NewService(homeDAO)

	// 3) 初始化开奖 DAO 与用户域（论坛）RPC 客户端。
	lotteryDAO := dao.NewLotteryDAO(db)
	userClient := userclient.New(c.UserRpc)
	// 4) 初始化 Redis 缓存客户端（可选）。
	var redisClient *redis.Client
	if strings.TrimSpace(c.CacheRedis.Addr) != "" {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     strings.TrimSpace(c.CacheRedis.Addr),
			Password: c.CacheRedis.Password,
			DB:       c.CacheRedis.DB,
		})
	}
	sceneTTL := time.Duration(c.CacheRedis.SceneTTLSeconds) * time.Second

	// 5) 初始化开奖服务；评论聚合优先走用户域 RPC，失败自动本地降级。
	lotteryCore := lotteryModule.NewService(lotteryDAO, userClient, redisClient, sceneTTL)

	// 6) 将全部核心模块注入上下文。
	return &ServiceContext{
		Config:      c,
		HomeCore:    homeCore,
		LotteryCore: lotteryCore,
	}, nil
}
