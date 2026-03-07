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

func main() {
	// 1) 读取配置文件路径参数。
	var configFile = flag.String("f", "etc/business.yaml", "the config file")
	flag.Parse()

	// 2) 加载业务服务配置（RPC/DB/Redis/UserRpc）。
	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 3) 初始化业务上下文：数据库、DAO、模块服务、下游客户端。
	svcCtx, err := svc.NewServiceContext(c)
	if err != nil {
		panic(fmt.Sprintf("init tk-business failed: %v", err))
	}

	// 4) 注册 gRPC 业务服务实现。
	rpcServer := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		tkv1.RegisterBusinessServiceServer(grpcServer, server.NewBusinessServer(svcCtx))
	})
	defer rpcServer.Stop()

	// 5) 打印启动日志并进入阻塞监听。
	logx.Infof("starting tk-business rpc on %s", c.ListenOn)
	rpcServer.Start()
}
