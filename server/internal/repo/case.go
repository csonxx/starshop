// Package repo 提供案例仓储.
package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"star/server/internal/model"
)

// CaseRepo 负责 Case 集合的持久化操作.
type CaseRepo struct {
	coll *mongo.Collection
}

// NewCaseRepo 构造.
func NewCaseRepo(store *mongo.Database) *CaseRepo {
	return &CaseRepo{coll: store.Collection(model.CollCase)}
}

// Coll 暴露 collection.
func (r *CaseRepo) Coll() *mongo.Collection {
	return r.coll
}

// Clear 删除全部 Case.
func (r *CaseRepo) Clear(ctx context.Context) error {
	_, err := r.coll.DeleteMany(ctx, bson.M{})
	return err
}

// CaseFilter 是公开/后台筛选条件.
// 为了保证语义清晰, 不同维度之间使用 AND, 同一维度的多选用 $in.
type CaseFilter struct {
	Style      string
	Space      []string
	Color      []string
	Size       []string
	Price      []string
	Q          string
	OnlyPinned bool
	OnlyActive bool
	Page       int
	Size2      int
}

// normalizeLimit 统一 size/分页边界.
func (f CaseFilter) normalizeLimit() (int, int) {
	page := f.Page
	if page < 1 {
		page = 1
	}
	size := f.Size2
	if size <= 0 || size > 60 {
		size = 24
	}
	return page, size
}

// List 返回案例列表和总数.
// 总数与当前页通过 $facet 在一次聚合内获取, 保证一致性.
func (r *CaseRepo) List(ctx context.Context, f CaseFilter) ([]model.Case, int64, error) {
	page, size := f.normalizeLimit()
	match := buildCaseMatch(f)

	pipeline := []bson.M{
		{"$match": match},
		{
			"$facet": bson.M{
				"items": []bson.M{
					{"$sort": bson.M{"pinned": -1, "createdAt": -1, "_id": -1}},
					{"$skip": int64((page - 1) * size)},
					{"$limit": int64(size)},
				},
				"total": []bson.M{{"$count": "n"}},
			},
		},
	}

	cur, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, fmt.Errorf("list cases aggregate: %w", err)
	}
	defer cur.Close(ctx)

	var raw []bson.M
	if err := cur.All(ctx, &raw); err != nil {
		return nil, 0, fmt.Errorf("list cases cursor: %w", err)
	}

	out := make([]model.Case, 0)
	var total int64
	if len(raw) > 0 {
		if items, ok := raw[0]["items"].(bson.A); ok {
			for _, item := range items {
				if m, ok := item.(bson.M); ok {
					cc, err := decodeCase(m)
					if err != nil {
						return nil, 0, fmt.Errorf("decode case: %w", err)
					}
					out = append(out, cc)
				}
			}
		}
		if totals, ok := raw[0]["total"].(bson.A); ok && len(totals) > 0 {
			if m, ok := totals[0].(bson.M); ok {
				if v, ok := m["n"]; ok {
					switch x := v.(type) {
					case int32:
						total = int64(x)
					case int64:
						total = x
					}
				}
			}
		}
	}
	return out, total, nil
}

// Pinned 返回置顶的启用案例.
func (r *CaseRepo) Pinned(ctx context.Context) ([]model.Case, error) {
	cur, err := r.coll.Find(ctx, bson.M{"pinned": true, "enabled": true},
		options.Find().SetSort(bson.M{"createdAt": -1}).SetLimit(8))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var list []model.Case
	if err := cur.All(ctx, &list); err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.Case{}
	}
	return list, nil
}

// Get 按 ID 查询. 公开用仅返回启用, 后台可调用 GetAny 不带 enabled 过滤.
func (r *CaseRepo) Get(ctx context.Context, id any) (model.Case, error) {
	res := r.coll.FindOne(ctx, bson.M{"_id": id, "enabled": true})
	return decodeCaseResult(res)
}

// GetAny 后台用: 不强制 enabled, 可用于编辑/审计已下架内容.
func (r *CaseRepo) GetAny(ctx context.Context, id any) (model.Case, error) {
	res := r.coll.FindOne(ctx, bson.M{"_id": id})
	return decodeCaseResult(res)
}

// Insert 直接写入.
func (r *CaseRepo) Insert(ctx context.Context, c model.Case) (model.Case, error) {
	now := time.Now().UTC()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
	c.UpdatedAt = now
	res, err := r.coll.InsertOne(ctx, c)
	if err != nil {
		return model.Case{}, err
	}
	if oid, ok := res.InsertedID.(interface{ Hex() string }); ok {
		c.ID, _ = decodeObjectID(oid.Hex())
	}
	return c, nil
}

// Update 按 ID 局部更新. 实际命中行数为 0 时返回 NotFound 错误.
func (r *CaseRepo) Update(ctx context.Context, id any, fields bson.M) error {
	fields["updatedAt"] = time.Now().UTC()
	res, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": fields})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete 按 ID 删除. 实际命中行数为 0 时返回 NotFound 错误.
func (r *CaseRepo) Delete(ctx context.Context, id any) error {
	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// CountPinned 用于后台校验 8 个置顶上限 (仅启用案例进入名额).
func (r *CaseRepo) CountPinned(ctx context.Context) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.M{"pinned": true, "enabled": true})
}

// ListAll 后台用: 全量返回案例, 业务上限依赖总规模.
func (r *CaseRepo) ListAll(ctx context.Context) ([]model.Case, error) {
	cur, err := r.coll.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var list []model.Case
	if err := cur.All(ctx, &list); err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.Case{}
	}
	return list, nil
}

// EnsureIndexes 为常用查询/唯一约束建立索引.
func (r *CaseRepo) EnsureIndexes(ctx context.Context) error {
	models := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "style", Value: 1}, {Key: "enabled", Value: 1}, {Key: "pinned", Value: -1}, {Key: "createdAt", Value: -1}},
			Options: options.Index().SetName("ix_case_style_active_pinned"),
		},
		{
			Keys:    bson.D{{Key: "space", Value: 1}, {Key: "enabled", Value: 1}},
			Options: options.Index().SetName("ix_case_space_active"),
		},
		{
			Keys:    bson.D{{Key: "size", Value: 1}, {Key: "enabled", Value: 1}},
			Options: options.Index().SetName("ix_case_size_active"),
		},
		{
			Keys:    bson.D{{Key: "priceLabel", Value: 1}, {Key: "enabled", Value: 1}},
			Options: options.Index().SetName("ix_case_priceLabel_active"),
		},
		{
			Keys:    bson.D{{Key: "colors", Value: 1}, {Key: "enabled", Value: 1}},
			Options: options.Index().SetName("ix_case_colors_active"),
		},
		{
			Keys:    bson.D{{Key: "pinned", Value: 1}, {Key: "enabled", Value: 1}},
			Options: options.Index().SetName("ix_case_pinned_active"),
		},
	}
	_, err := r.coll.Indexes().CreateMany(ctx, models)
	return err
}

// buildCaseMatch 把统一筛选条件翻译为 BSON 过滤器.
// 一级风格固定 AND; 二级维度统一 AND (不同维度同时满足); 同维度多选用 $in.
func buildCaseMatch(f CaseFilter) bson.M {
	q := bson.M{}
	if f.Style != "" {
		q["style"] = f.Style
	}
	if f.OnlyActive {
		q["enabled"] = true
	}
	if f.OnlyPinned {
		q["pinned"] = true
	}

	if len(f.Space) > 0 {
		if len(f.Space) == 1 {
			q["space"] = f.Space[0]
		} else {
			q["space"] = bson.M{"$in": f.Space}
		}
	}
	if len(f.Color) > 0 {
		if len(f.Color) == 1 {
			q["colors"] = f.Color[0]
		} else {
			q["colors"] = bson.M{"$in": f.Color}
		}
	}
	if len(f.Size) > 0 {
		if len(f.Size) == 1 {
			q["size"] = f.Size[0]
		} else {
			q["size"] = bson.M{"$in": f.Size}
		}
	}
	if len(f.Price) > 0 {
		if len(f.Price) == 1 {
			q["priceLabel"] = f.Price[0]
		} else {
			q["priceLabel"] = bson.M{"$in": f.Price}
		}
	}

	if f.Q != "" {
		// 模糊搜索使用转义后的正则匹配标题/小区/户型.
		q["$or"] = []bson.M{
			{"title": bson.M{"$regex": regexMeta(f.Q), "$options": "i"}},
			{"estate": bson.M{"$regex": regexMeta(f.Q), "$options": "i"}},
			{"rooms": bson.M{"$regex": regexMeta(f.Q), "$options": "i"}},
		}
	}
	return q
}

// ErrNotFound 表示未命中.
var ErrNotFound = errors.New("repo: not found")
