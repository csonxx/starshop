// Package repo 提供 Tag 仓储.
package repo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"star/server/internal/model"
)

// TagRepo 负责 Tag 集合.
type TagRepo struct {
	coll *mongo.Collection
}

// NewTagRepo 构造.
func NewTagRepo(store *mongo.Database) *TagRepo {
	return &TagRepo{coll: store.Collection(model.CollTag)}
}

// Coll 暴露 collection.
func (r *TagRepo) Coll() *mongo.Collection {
	return r.coll
}

// Clear 删除全部 Tag.
func (r *TagRepo) Clear(ctx context.Context) error {
	_, err := r.coll.DeleteMany(ctx, bson.M{})
	return err
}

// ListEnabled 返回某 type 下所有启用标签.
func (r *TagRepo) ListEnabled(ctx context.Context, typ string) ([]model.Tag, error) {
	cur, err := r.coll.Find(ctx, bson.M{"type": typ, "enabled": true},
		options.Find().SetSort(bson.D{{Key: "sort", Value: 1}, {Key: "createdAt", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var list []model.Tag
	if err := cur.All(ctx, &list); err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.Tag{}
	}
	return list, nil
}

// ListAny 后台列出全部, 可指定 type (空表示全部).
func (r *TagRepo) ListAny(ctx context.Context, typ string) ([]model.Tag, error) {
	q := bson.M{}
	if typ != "" {
		q["type"] = typ
	}
	cur, err := r.coll.Find(ctx, q, options.Find().SetSort(bson.D{{Key: "type", Value: 1}, {Key: "sort", Value: 1}, {Key: "createdAt", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var list []model.Tag
	if err := cur.All(ctx, &list); err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.Tag{}
	}
	return list, nil
}

// Get 按 ID 获取.
func (r *TagRepo) Get(ctx context.Context, id primitive.ObjectID) (model.Tag, error) {
	res := r.coll.FindOne(ctx, bson.M{"_id": id})
	if err := res.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Tag{}, ErrNotFound
		}
		return model.Tag{}, err
	}
	var t model.Tag
	if err := res.Decode(&t); err != nil {
		return model.Tag{}, err
	}
	return t, nil
}

// HasEnabledValue 检查同 type 下 value 是否存在并启用.
func (r *TagRepo) HasEnabledValue(ctx context.Context, typ, value string) (bool, error) {
	n, err := r.coll.CountDocuments(ctx, bson.M{"type": typ, "value": value, "enabled": true})
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// CountReferenced 统计有多少启用案例引用了该 tag 的 value.
func (r *TagRepo) CountReferenced(ctx context.Context, typ, value string) (int64, error) {
	field, ok := tagField(typ)
	if !ok {
		return 0, nil
	}
	store := r.coll.Database()
	caseColl := store.Collection(model.CollCase)
	if field == "colors" {
		return caseColl.CountDocuments(ctx, bson.M{field: value, "enabled": true})
	}
	if field == "style" || field == "space" || field == "size" || field == "priceLabel" {
		return caseColl.CountDocuments(ctx, bson.M{field: value, "enabled": true})
	}
	return 0, nil
}

// Insert 新增.
func (r *TagRepo) Insert(ctx context.Context, t model.Tag) (model.Tag, error) {
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now().UTC()
	}
	t.UpdatedAt = t.CreatedAt
	res, err := r.coll.InsertOne(ctx, t)
	if err != nil {
		return model.Tag{}, err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		t.ID = oid
	}
	return t, nil
}

// Update 部分更新.
func (r *TagRepo) Update(ctx context.Context, id primitive.ObjectID, fields bson.M) error {
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

// Delete 删除.
func (r *TagRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// CountAny 统计全部标签.
func (r *TagRepo) CountAny(ctx context.Context, typ string) (int64, error) {
	q := bson.M{}
	if typ != "" {
		q["type"] = typ
	}
	return r.coll.CountDocuments(ctx, q)
}

// EnsureIndexes 建 (type, value) 唯一索引等.
func (r *TagRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "type", Value: 1}, {Key: "value", Value: 1}},
			Options: options.Index().SetName("uq_tag_type_value").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "type", Value: 1}, {Key: "enabled", Value: 1}, {Key: "sort", Value: 1}},
			Options: options.Index().SetName("ix_tag_type_active_sort"),
		},
	})
	return err
}

// tagField 把逻辑字段映射到 Case 文档字段.
func tagField(typ string) (string, bool) {
	switch typ {
	case model.TagStyle:
		return "style", true
	case model.TagSpace:
		return "space", true
	case model.TagColor:
		return "colors", true
	case model.TagSize:
		return "size", true
	case model.TagPrice:
		return "priceLabel", true
	}
	return "", false
}
