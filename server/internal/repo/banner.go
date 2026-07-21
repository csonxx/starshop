// Package repo 提供 Banner 仓储.
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

// BannerRepo 负责 Banner 集合.
type BannerRepo struct {
	coll *mongo.Collection
}

// NewBannerRepo 构造.
func NewBannerRepo(store *mongo.Database) *BannerRepo {
	return &BannerRepo{coll: store.Collection(model.CollBanner)}
}

// Coll 暴露 collection.
func (r *BannerRepo) Coll() *mongo.Collection {
	return r.coll
}

// Clear 删除全部 Banner.
func (r *BannerRepo) Clear(ctx context.Context) error {
	_, err := r.coll.DeleteMany(ctx, bson.M{})
	return err
}

// FindEnabled 返回已启用 Banner, limit<=0 时不截断.
func (r *BannerRepo) FindEnabled(ctx context.Context, limit int64) ([]model.Banner, error) {
	opts := options.Find().SetSort(bson.D{{Key: "sort", Value: 1}, {Key: "createdAt", Value: -1}})
	if limit > 0 {
		opts = opts.SetLimit(limit)
	}
	cur, err := r.coll.Find(ctx, bson.M{"enabled": true}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var list []model.Banner
	if err := cur.All(ctx, &list); err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.Banner{}
	}
	return list, nil
}

// FindAny 返回全部 Banner, 后台用.
func (r *BannerRepo) FindAny(ctx context.Context) ([]model.Banner, error) {
	cur, err := r.coll.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "sort", Value: 1}, {Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var list []model.Banner
	if err := cur.All(ctx, &list); err != nil {
		return nil, err
	}
	if list == nil {
		list = []model.Banner{}
	}
	return list, nil
}

// Get 按 ID 查询.
func (r *BannerRepo) Get(ctx context.Context, id primitive.ObjectID) (model.Banner, error) {
	res := r.coll.FindOne(ctx, bson.M{"_id": id})
	if err := res.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Banner{}, ErrNotFound
		}
		return model.Banner{}, err
	}
	var b model.Banner
	if err := res.Decode(&b); err != nil {
		return model.Banner{}, err
	}
	return b, nil
}

// Insert 新增.
func (r *BannerRepo) Insert(ctx context.Context, b model.Banner) (model.Banner, error) {
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now().UTC()
	}
	b.UpdatedAt = b.CreatedAt
	res, err := r.coll.InsertOne(ctx, b)
	if err != nil {
		return model.Banner{}, err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		b.ID = oid
	}
	return b, nil
}

// Update 按 ID 部分更新.
func (r *BannerRepo) Update(ctx context.Context, id primitive.ObjectID, fields bson.M) error {
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

// Delete 按 ID 删除.
func (r *BannerRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	res, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// CountEnabled 统计启用数量.
func (r *BannerRepo) CountEnabled(ctx context.Context) (int64, error) {
	return r.coll.CountDocuments(ctx, bson.M{"enabled": true})
}

// EnsureIndexes 为 Banner 集合建索引.
func (r *BannerRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "enabled", Value: 1}, {Key: "sort", Value: 1}, {Key: "createdAt", Value: -1}},
			Options: options.Index().SetName("ix_banner_active_sort"),
		},
	})
	return err
}
