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
	// 处理当前语句逻辑。
	SystemComments []map[string]interface{} `json:"system_comments"`
	// 处理当前语句逻辑。
	UserComments []map[string]interface{} `json:"user_comments"`
	// 处理当前语句逻辑。
	HotComments []map[string]interface{} `json:"hot_comments"`
	// 处理当前语句逻辑。
	LatestComments []map[string]interface{} `json:"latest_comments"`
}

// Client 定义业务服务访问用户域微服务的最小接口。
type Client interface {
	// 调用IsEnabled完成当前处理。
	IsEnabled() bool
	// 调用TopicList完成当前处理。
	TopicList(ctx context.Context, limit int) ([]map[string]interface{}, error)
	// 调用LotteryCommentGroups完成当前处理。
	LotteryCommentGroups(ctx context.Context, lotteryInfoID uint) (LotteryCommentGroups, error)
}

// 定义当前类型结构。
type grpcClient struct {
	// 处理当前语句逻辑。
	enabled bool
	// 处理当前语句逻辑。
	rpc tkv1.UserServiceClient
}

// New 创建对象实例。
func New(clientConf zrpc.RpcClientConf) Client {
	// 判断条件并进入对应分支逻辑。
	if len(clientConf.Endpoints) == 0 && clientConf.Target == "" {
		// 返回当前处理结果。
		return &grpcClient{enabled: false}
	}
	// 连接用户域 RPC；当前由 tk-user 服务承载论坛评论能力。
	cli := zrpc.MustNewClient(clientConf)
	// 返回当前处理结果。
	return &grpcClient{enabled: true, rpc: tkv1.NewUserServiceClient(cli.Conn())}
}

// IsEnabled 处理IsEnabled相关逻辑。
func (c *grpcClient) IsEnabled() bool { return c != nil && c.enabled && c.rpc != nil }

// TopicList 处理TopicList相关逻辑。
func (c *grpcClient) TopicList(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	// 判断条件并进入对应分支逻辑。
	if !c.IsEnabled() {
		// 返回当前处理结果。
		return nil, fmt.Errorf("user grpc client disabled")
	}
	// 定义并初始化当前变量。
	resp, err := c.rpc.TopicList(ctx, &tkv1.TopicListRequest{Limit: int32(limit)})
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 判断条件并进入对应分支逻辑。
	if resp.GetCode() != 0 {
		// 返回当前处理结果。
		return nil, fmt.Errorf("%s", resp.GetMsg())
	}
	// 声明当前变量。
	var payload struct {
		// 处理当前语句逻辑。
		Items []map[string]interface{} `json:"items"`
	}
	// 判断条件并进入对应分支逻辑。
	if err := json.Unmarshal([]byte(resp.GetDataJson()), &payload); err != nil {
		// 返回当前处理结果。
		return nil, err
	}
	// 返回当前处理结果。
	return payload.Items, nil
}

// LotteryCommentGroups 处理LotteryCommentGroups相关逻辑。
func (c *grpcClient) LotteryCommentGroups(ctx context.Context, lotteryInfoID uint) (LotteryCommentGroups, error) {
	// 判断条件并进入对应分支逻辑。
	if !c.IsEnabled() {
		// 返回当前处理结果。
		return LotteryCommentGroups{}, fmt.Errorf("user grpc client disabled")
	}
	// 定义并初始化当前变量。
	resp, err := c.rpc.LotteryCommentGroups(ctx, &tkv1.LotteryCommentGroupsRequest{LotteryInfoId: uint64(lotteryInfoID)})
	// 判断条件并进入对应分支逻辑。
	if err != nil {
		// 返回当前处理结果。
		return LotteryCommentGroups{}, err
	}
	// 判断条件并进入对应分支逻辑。
	if resp.GetCode() != 0 {
		// 返回当前处理结果。
		return LotteryCommentGroups{}, fmt.Errorf("%s", resp.GetMsg())
	}
	// 定义并初始化当前变量。
	payload := LotteryCommentGroups{}
	// 判断条件并进入对应分支逻辑。
	if err := json.Unmarshal([]byte(resp.GetDataJson()), &payload); err != nil {
		// 返回当前处理结果。
		return LotteryCommentGroups{}, err
	}
	// 返回当前处理结果。
	return payload, nil
}
