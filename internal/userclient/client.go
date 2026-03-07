package userclient

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/zeromicro/go-zero/zrpc"
	tkv1 "tk-proto/tk/v1"
)

// LotteryCommentGroups 对应彩种详情页需要的四组评论数据。
type LotteryCommentGroups struct {
	SystemComments []map[string]interface{} `json:"system_comments"`
	UserComments   []map[string]interface{} `json:"user_comments"`
	HotComments    []map[string]interface{} `json:"hot_comments"`
	LatestComments []map[string]interface{} `json:"latest_comments"`
}

// Client 定义业务服务访问用户域微服务的最小接口。
type Client interface {
	IsEnabled() bool
	TopicList(ctx context.Context, limit int) ([]map[string]interface{}, error)
	LotteryCommentGroups(ctx context.Context, lotteryInfoID uint) (LotteryCommentGroups, error)
}

type grpcClient struct {
	enabled bool
	rpc     tkv1.UserServiceClient
}

func New(clientConf zrpc.RpcClientConf) Client {
	if len(clientConf.Endpoints) == 0 && clientConf.Target == "" {
		return &grpcClient{enabled: false}
	}
	// 连接用户域 RPC；当前由 tk-user 服务承载论坛评论能力。
	cli := zrpc.MustNewClient(clientConf)
	return &grpcClient{enabled: true, rpc: tkv1.NewUserServiceClient(cli.Conn())}
}

func (c *grpcClient) IsEnabled() bool { return c != nil && c.enabled && c.rpc != nil }

func (c *grpcClient) TopicList(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	if !c.IsEnabled() {
		return nil, fmt.Errorf("user grpc client disabled")
	}
	resp, err := c.rpc.TopicList(ctx, &tkv1.TopicListRequest{Limit: int32(limit)})
	if err != nil {
		return nil, err
	}
	if resp.GetCode() != 0 {
		return nil, fmt.Errorf("%s", resp.GetMsg())
	}
	var payload struct {
		Items []map[string]interface{} `json:"items"`
	}
	if err := json.Unmarshal([]byte(resp.GetDataJson()), &payload); err != nil {
		return nil, err
	}
	return payload.Items, nil
}

func (c *grpcClient) LotteryCommentGroups(ctx context.Context, lotteryInfoID uint) (LotteryCommentGroups, error) {
	if !c.IsEnabled() {
		return LotteryCommentGroups{}, fmt.Errorf("user grpc client disabled")
	}
	resp, err := c.rpc.LotteryCommentGroups(ctx, &tkv1.LotteryCommentGroupsRequest{LotteryInfoId: uint64(lotteryInfoID)})
	if err != nil {
		return LotteryCommentGroups{}, err
	}
	if resp.GetCode() != 0 {
		return LotteryCommentGroups{}, fmt.Errorf("%s", resp.GetMsg())
	}
	payload := LotteryCommentGroups{}
	if err := json.Unmarshal([]byte(resp.GetDataJson()), &payload); err != nil {
		return LotteryCommentGroups{}, err
	}
	return payload, nil
}
