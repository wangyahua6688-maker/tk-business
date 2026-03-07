package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

// Config 业务服务配置。
type Config struct {
	// RpcServerConf：go-zero gRPC 服务端监听配置。
	zrpc.RpcServerConf
	// Database：业务库连接配置。
	Database struct {
		DSN string
	}
	// CacheRedis 现场页缓存配置（用于 live-scene 聚合结果缓存）。
	CacheRedis struct {
		// Addr Redis 地址，例：127.0.0.1:6379。
		Addr string
		// Password Redis 鉴权密码。
		Password string
		// DB Redis 分库编号。
		DB int
		// SceneTTLSeconds 现场页聚合缓存 TTL（秒）。
		SceneTTLSeconds int
	}
	// UserRpc 用户域 RPC 客户端（当前承载论坛评论能力）。
	UserRpc zrpc.RpcClientConf
}
