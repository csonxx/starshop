package repo

import (
	"go.mongodb.org/mongo-driver/mongo/options"
)

func mongoUpsertReturnAfter() *options.FindOneAndUpdateOptions {
	return options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
}

func mongoUpsertOnly() *options.UpdateOptions {
	return options.Update()
}