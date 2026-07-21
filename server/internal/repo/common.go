// Package repo 公共助手.
package repo

import (
	"errors"
	"fmt"
	"regexp"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"star/server/internal/model"
)

// decodeObjectID 接受 hex 字符串返回 ObjectID.
func decodeObjectID(hex string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(hex)
}

// MustObjectID 解析失败时回退为零 ID. 注意: 调用方应根据上下文判断是否致命.
func MustObjectID(hex string) primitive.ObjectID {
	oid, err := primitive.ObjectIDFromHex(hex)
	if err != nil {
		return primitive.NilObjectID
	}
	return oid
}

// regexMeta 转义正则元字符, 避免用户输入被解读为正则.
func regexMeta(s string) string {
	return regexp.QuoteMeta(s)
}

// decodeCase 把 BSON 文档解析为 Case, 同时把 _id 标准化为 model.ObjectID.
func decodeCase(raw bson.M) (model.Case, error) {
	bb, err := bson.Marshal(raw)
	if err != nil {
		return model.Case{}, err
	}
	var c model.Case
	if err := bson.Unmarshal(bb, &c); err != nil {
		return model.Case{}, err
	}
	return c, nil
}

// decodeCaseResult 把 *mongo.SingleResult 安全地解析为 Case.
func decodeCaseResult(res *mongo.SingleResult) (model.Case, error) {
	if res == nil {
		return model.Case{}, ErrNotFound
	}
	if err := res.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Case{}, ErrNotFound
		}
		return model.Case{}, fmt.Errorf("case find one: %w", err)
	}
	var c model.Case
	if err := res.Decode(&c); err != nil {
		return model.Case{}, fmt.Errorf("case decode: %w", err)
	}
	return c, nil
}

// SplitList 接受逗号或重复参数拼接的字符串, 输出去重/去空切片.
func SplitList(s string) []string {
	return splitList(s)
}
