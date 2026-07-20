// Package repo Mongo 数据访问层, 负责把 model 转成 bson 查询.
package repo

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"star/server/internal/model"
)

// CaseRepo 案例 collection 的 CRUD 仓库
type CaseRepo struct{ coll *mongo.Collection }

func NewCaseRepo(c *mongo.Collection) *CaseRepo { return &CaseRepo{coll: c} }

// CaseFilter 前台案例过滤条件
type CaseFilter struct {
	Style  string // 一级风格 (AND 精确匹配)
	Space  string // 二级空间
	Color  string // 二级颜色
	Size   string // 二级尺寸
	Price  string // 二级价格区间
	Search string // 标题模糊搜索
	Page   int    // 分页: 第几页 (从 1 起)
	Size2  int    // 分页: 每页条数 (最大 60)
}

// List 查询案例. 一级风格用 AND, 其余二级条件用 OR 组合, 满足 UI "任一选中都有数据".
// 排序: 置顶优先 + 按时间倒序.
func (r *CaseRepo) List(ctx context.Context, f CaseFilter) ([]model.Case, int64, error) {
	q := bson.M{"enabled": true}

	// 一级: 风格精确匹配
	if f.Style != "" {
		q["style"] = f.Style
	}
	// 标题模糊搜索
	if f.Search != "" {
		q["title"] = bson.M{"$regex": primitive.Regex{
			Pattern: strings.ReplaceAll(f.Search, " ", ".*"),
			Options: "i",
		}}
	}

	// 二级条件 (空间/尺寸/颜色/价格) 用 OR 拼接
	var secondary []bson.M
	if f.Space != "" {
		secondary = append(secondary, bson.M{"space": f.Space})
	}
	if f.Size != "" {
		secondary = append(secondary, bson.M{"size": f.Size})
	}
	if f.Color != "" {
		secondary = append(secondary, bson.M{"colors": f.Color})
	}
	if f.Price != "" {
		secondary = append(secondary, bson.M{"priceLabel": f.Price})
	}
	if len(secondary) > 0 {
		q["$or"] = secondary
	}

	total, err := r.coll.CountDocuments(ctx, q)
	if err != nil {
		return nil, 0, err
	}

	page := f.Page
	if page < 1 {
		page = 1
	}
	size := f.Size2
	if size < 1 || size > 60 {
		size = 24
	}

	cur, err := r.coll.Find(ctx, q,
		options.Find().
			SetSort(bson.D{{Key: "pinned", Value: -1}, {Key: "createdAt", Value: -1}}).
			SetSkip(int64((page-1)*size)).
			SetLimit(int64(size)),
	)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)

	var out []model.Case
	if err := cur.All(ctx, &out); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// ListPinned 首页置顶案例 (后台可置顶, 最多 8 个)
func (r *CaseRepo) ListPinned(ctx context.Context) ([]model.Case, error) {
	cur, err := r.coll.Find(ctx,
		bson.M{"enabled": true, "pinned": true},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}).SetLimit(8),
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var out []model.Case
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Get 单条案例
func (r *CaseRepo) Get(ctx context.Context, id primitive.ObjectID) (*model.Case, error) {
	var c model.Case
	if err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

// ListAll 后台全量列表 (包括禁用), 按置顶+时间排序
func (r *CaseRepo) ListAll(ctx context.Context) ([]model.Case, error) {
	cur, err := r.coll.Find(ctx, bson.M{},
		options.Find().SetSort(bson.D{{Key: "pinned", Value: -1}, {Key: "createdAt", Value: -1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []model.Case
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Insert 入库, 自动填 CreatedAt 和 ID
func (r *CaseRepo) Insert(ctx context.Context, c *model.Case) error {
	if c.CreatedAt.IsZero() {
		c.CreatedAt = time.Now()
	}
	res, err := r.coll.InsertOne(ctx, c)
	if err != nil {
		return err
	}
	c.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

// Update 局部字段更新 (用于 TogglePin 等 $set 操作)
func (r *CaseRepo) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

// Delete 物理删除 (后台软删除可改用 enabled=false)
func (r *CaseRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// Clear 整表清空 (仅 seed 使用)
func (r *CaseRepo) Clear(ctx context.Context) error {
	_, err := r.coll.DeleteMany(ctx, bson.M{})
	return err
}