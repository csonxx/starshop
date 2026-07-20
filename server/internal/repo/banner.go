package repo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"star/server/internal/model"
)

type BannerRepo struct{ coll *mongo.Collection }

func NewBannerRepo(c *mongo.Collection) *BannerRepo { return &BannerRepo{coll: c} }

func (r *BannerRepo) ListEnabled(ctx context.Context) ([]model.Banner, error) {
	cur, err := r.coll.Find(ctx,
		bson.M{"enabled": true},
		options.Find().SetSort(bson.D{{Key: "sort", Value: 1}, {Key: "createdAt", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []model.Banner
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *BannerRepo) ListAll(ctx context.Context) ([]model.Banner, error) {
	cur, err := r.coll.Find(ctx, bson.M{},
		options.Find().SetSort(bson.D{{Key: "sort", Value: 1}, {Key: "createdAt", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []model.Banner
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *BannerRepo) Insert(ctx context.Context, b *model.Banner) error {
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now()
	}
	res, err := r.coll.InsertOne(ctx, b)
	if err != nil {
		return err
	}
	b.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *BannerRepo) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func (r *BannerRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *BannerRepo) Clear(ctx context.Context) error {
	_, err := r.coll.DeleteMany(ctx, bson.M{})
	return err
}