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

type TagRepo struct{ coll *mongo.Collection }

func NewTagRepo(c *mongo.Collection) *TagRepo { return &TagRepo{coll: c} }

func (r *TagRepo) ListEnabled(ctx context.Context, typ string) ([]model.Tag, error) {
	q := bson.M{"enabled": true}
	if typ != "" {
		q["type"] = typ
	}
	cur, err := r.coll.Find(ctx, q,
		options.Find().SetSort(bson.D{{Key: "sort", Value: 1}, {Key: "createdAt", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []model.Tag
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *TagRepo) ListAll(ctx context.Context, typ string) ([]model.Tag, error) {
	q := bson.M{}
	if typ != "" {
		q["type"] = typ
	}
	cur, err := r.coll.Find(ctx, q,
		options.Find().SetSort(bson.D{{Key: "sort", Value: 1}, {Key: "createdAt", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []model.Tag
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *TagRepo) Insert(ctx context.Context, t *model.Tag) error {
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	res, err := r.coll.InsertOne(ctx, t)
	if err != nil {
		return err
	}
	t.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *TagRepo) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func (r *TagRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *TagRepo) Clear(ctx context.Context) error {
	_, err := r.coll.DeleteMany(ctx, bson.M{})
	return err
}