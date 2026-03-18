package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"tk-business/internal/config"
	"tk-business/internal/server"
	"tk-business/internal/svc"
	tkv1 "tk-proto/tk/v1"
)

// main 启动程序入口。
func main() {
	// 1) 读取配置文件路径参数。
	var configFile = flag.String("f", "etc/business.yaml", "the config file")
	// 调用flag.Parse完成当前处理。
	flag.Parse()

	// 2) 加载业务服务配置（RPC/DB/Redis/UserRpc）。
	var c config.Config
	// 调用conf.MustLoad完成当前处理。
	conf.MustLoad(*configFile, &c)

	// 3) 初始化业务上下文：数据库、DAO、模块服务、下游客户端。
	svcCtx, err := svc.NewServiceContext(c)
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 调用panic完成当前处理。
		panic(fmt.Sprintf("init tk-business failed: %v", err))
	}

	// 4) 注册 gRPC 业务服务实现。
	rpcServer := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		// 调用tkv1.RegisterBusinessServiceServer完成当前处理。
		tkv1.RegisterBusinessServiceServer(grpcServer, server.NewBusinessServer(svcCtx))
	})
	// 注册延迟执行逻辑。
	defer rpcServer.Stop()

	// 5) 打印启动日志并进入阻塞监听。
	logx.Infof("starting tk-business rpc on %s", c.ListenOn)
	// 调用rpcServer.Start完成当前处理。
	rpcServer.Start()
}
