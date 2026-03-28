package server

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// decodePayloadToProto 将业务层 map payload 转成 typed proto message。
func decodePayloadToProto(payload interface{}, msg proto.Message) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload to json failed: %w", err)
	}
	options := protojson.UnmarshalOptions{DiscardUnknown: false}
	if err := options.Unmarshal(raw, msg); err != nil {
		return fmt.Errorf("unmarshal payload to proto failed: %w", err)
	}
	return nil
}
