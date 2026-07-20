package repo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"star/server/internal/model"
)

type UserRepo struct{ coll *mongo.Collection }

func NewUserRepo(c *mongo.Collection) *UserRepo { return &UserRepo{coll: c} }

func (r *UserRepo) Coll() *mongo.Collection { return r.coll }

func (r *UserRepo) FindByPhone(ctx context.Context, phone string) (*model.User, error) {
	var u model.User
	err := r.coll.FindOne(ctx, bson.M{"phone": phone}).Decode(&u)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) UpsertByPhone(ctx context.Context, phone string) (*model.User, error) {
	now := time.Now()
	update := bson.M{
		"$setOnInsert": bson.M{
			"phone":     phone,
			"role":      model.RoleUser,
			"createdAt": now,
		},
		"$set": bson.M{
			"updatedAt": now,
		},
	}
	opts := mongoUpsertReturnAfter()
	var u model.User
	err := r.coll.FindOneAndUpdate(ctx, bson.M{"phone": phone}, update, opts).Decode(&u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) EnsureAdmin(ctx context.Context, phone string) error {
	now := time.Now()
	_, err := r.coll.UpdateOne(ctx,
		bson.M{"phone": phone},
		bson.M{
			"$setOnInsert": bson.M{
				"phone":     phone,
				"role":      model.RoleAdmin,
				"nickname":  "星仔管理员",
				"createdAt": now,
			},
		},
		mongoUpsertOnly().SetUpsert(true),
	)
	return err
}

func (r *UserRepo) Get(ctx context.Context, id primitive.ObjectID) (*model.User, error) {
	var u model.User
	if err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}