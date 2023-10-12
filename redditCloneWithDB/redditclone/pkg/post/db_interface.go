package post

import (
	"context"

	mgo "go.mongodb.org/mongo-driver/mongo"
)

type CollectionHelper interface {
	Find(ctx context.Context, filter interface{}) (*mgo.Cursor, error)
	ReplaceOne(ctx context.Context, filter interface{}, replacement interface{}) (*mgo.UpdateResult, error)
	InsertOne(ctx context.Context, filter interface{}) (interface{}, error)
	FindOne(ctx context.Context, filter interface{}) *mgo.SingleResult
	DeleteOne(ctx context.Context, filter interface{}) (*mgo.DeleteResult, error)
}

type MgoLayer struct {
	c *mgo.Collection
}

func (m MgoLayer) Find(ctx context.Context, filter interface{}) (*mgo.Cursor, error) {
	return m.c.Find(ctx, filter)
}

func (m MgoLayer) ReplaceOne(ctx context.Context, filter interface{}, replacement interface{}) (*mgo.UpdateResult, error) {
	return m.c.ReplaceOne(ctx, filter, replacement)
}

func (m MgoLayer) FindOne(ctx context.Context, filter interface{}) *mgo.SingleResult {
	return m.c.FindOne(ctx, filter)
}

func (m MgoLayer) DeleteOne(ctx context.Context, filter interface{}) (*mgo.DeleteResult, error) {
	return m.c.DeleteOne(ctx, filter)
}

func (m MgoLayer) InsertOne(ctx context.Context, filter interface{}) (interface{}, error) {
	return m.c.InsertOne(ctx, filter)
}
