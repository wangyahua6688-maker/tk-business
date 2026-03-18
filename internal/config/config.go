package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 业务服务配置。
type Config struct {
	zrpc.RpcServerConf // go-zero gRPC 服务监听配置。

	Database struct { // 数据库分组。
		DSN string // 业务库 DSN 连接串。
	} // 数据库配置。

	CacheRedis struct { // 缓存分组。
		Addr            string // Redis 地址，例：127.0.0.1:6379。
		Password        string // Redis 鉴权密码。
		DB              int    // Redis 分库编号。
		SceneTTLSeconds int    // 现场页聚合缓存 TTL（秒）。
	} // 缓存配置。

	UserRpc zrpc.RpcClientConf // 用户域 RPC 客户端（评论、论坛相关能力）。
}
