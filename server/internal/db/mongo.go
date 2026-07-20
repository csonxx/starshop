// Package db 封装 MongoDB 连接池.
package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Store 是 MongoDB 客户端 + 选定 DB 的容器
type Store struct {
	Client *mongo.Client
	DB     *mongo.Database
}

// Connect 连上 MongoDB 并 Ping 验证可用性.
// 超时 5s, 失败返回 err, 由 cmd/server 或 cmd/seed 决定是否 fatal.
func Connect(ctx context.Context, uri, dbName string) (*Store, error) {
	clientOpts := options.Client().
		ApplyURI(uri).
		SetServerSelectionTimeout(5 * time.Second)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}
	return &Store{Client: client, DB: client.Database(dbName)}, nil
}

// Close 释放连接
func (s *Store) Close(ctx context.Context) error {
	if s == nil || s.Client == nil {
		return nil
	}
	return s.Client.Disconnect(ctx)
}

// Coll 简写: 给名字拿 collection (所有 repo 复用此 API)
func (s *Store) Coll(name string) *mongo.Collection {
	return s.DB.Collection(name)
}