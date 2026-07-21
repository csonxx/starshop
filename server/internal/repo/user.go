// Package repo 提供 User 仓储.
package repo

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"star/server/internal/model"
)

// UserRepo 用户仓库.
type UserRepo struct {
	coll *mongo.Collection
}

// NewUserRepo 构造.
func NewUserRepo(store interface{}) *UserRepo {
	switch value := store.(type) {
	case *mongo.Database:
		return &UserRepo{coll: value.Collection(model.CollUser)}
	case *mongo.Collection:
		return &UserRepo{coll: value}
	default:
		panic("unsupported user repository store")
	}
}

// Coll 暴露内部 collection, 用于后台账号列表等场景.
func (r *UserRepo) Coll() *mongo.Collection {
	return r.coll
}

// FindByPhone 通过手机号查找.
func (r *UserRepo) FindByPhone(ctx context.Context, phone string) (model.User, error) {
	res := r.coll.FindOne(ctx, bson.M{"phone": phone})
	if err := res.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, err
	}
	var u model.User
	if err := res.Decode(&u); err != nil {
		return model.User{}, err
	}
	return u, nil
}

// Get 按 ID 取.
func (r *UserRepo) Get(ctx context.Context, id primitive.ObjectID) (model.User, error) {
	res := r.coll.FindOne(ctx, bson.M{"_id": id})
	if err := res.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.User{}, ErrNotFound
		}
		return model.User{}, err
	}
	var u model.User
	if err := res.Decode(&u); err != nil {
		return model.User{}, err
	}
	return u, nil
}

// UpsertOnLogin 根据手机号自动 upsert; 已知角色且要更高权限时不下发.
func (r *UserRepo) UpsertOnLogin(ctx context.Context, phone, nickname, role string) (model.User, error) {
	now := time.Now().UTC()
	phone = strings.TrimSpace(phone)
	filter := bson.M{"phone": phone}
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
	if nickname != "" {
		update["$set"].(bson.M)["nickname"] = nickname
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	res := r.coll.FindOneAndUpdate(ctx, filter, update, opts)
	if err := res.Err(); err != nil {
		return model.User{}, err
	}
	var u model.User
	if err := res.Decode(&u); err != nil {
		return model.User{}, err
	}
	return u, nil
}

// SetRole 后台可调用, 用于授权销售/供应商/管理员.
func (r *UserRepo) SetRole(ctx context.Context, id primitive.ObjectID, role string) error {
	res, err := r.coll.UpdateOne(ctx, bson.M{"_id": id},
		bson.M{"$set": bson.M{"role": role, "updatedAt": time.Now().UTC()}})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// Clear 删除全部用户 (开发态 seed 使用).
func (r *UserRepo) Clear(ctx context.Context) error {
	_, err := r.coll.DeleteMany(ctx, bson.M{})
	return err
}

// UpsertByPhone 等同 UpsertOnLogin, 兼容旧名.
func (r *UserRepo) UpsertByPhone(ctx context.Context, phone, nickname string, role string) (model.User, error) {
	return r.UpsertOnLogin(ctx, phone, nickname, role)
}

// EnsureAdmin 根据白名单手机号强制 admin. 已存在用户保留更高权限.
func (r *UserRepo) EnsureAdmin(ctx context.Context, phone string) error {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return errors.New("admin phone is required")
	}
	now := time.Now().UTC()
	_, err := r.coll.UpdateOne(ctx, bson.M{"phone": phone},
		bson.M{
			"$set":         bson.M{"role": model.RoleAdmin, "updatedAt": now},
			"$setOnInsert": bson.M{"phone": phone, "createdAt": now},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

// EnsureIndex 为 phone 建立唯一索引.
func (r *UserRepo) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "phone", Value: 1}},
		Options: options.Index().SetName("uq_user_phone").SetUnique(true),
	})
	return err
}
