package repository

import (
    "context"
    "log"
    "time"

    "blog-service/model"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type LikeRepository struct {
    coll *mongo.Collection
}

func NewLikeRepository(db *mongo.Database) *LikeRepository {
    r := &LikeRepository{coll: db.Collection("likes")}
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := r.ensureIndexes(ctx); err != nil {
        log.Printf("like repository: ensure indexes: %v", err)
    }
    return r
}

func (r *LikeRepository) Create(ctx context.Context, l *model.Like) error {
    if l.CreatedAt.IsZero() {
        l.CreatedAt = time.Now().UTC()
    }
    _, err := r.coll.InsertOne(ctx, l)
    return err
}

func (r *LikeRepository) DeleteByBlogAndUser(ctx context.Context, blogID, userID primitive.ObjectID) (int64, error) {
    res, err := r.coll.DeleteOne(ctx, bson.M{"blog_id": blogID, "user_id": userID})
    if err != nil {
        return 0, err
    }
    return res.DeletedCount, nil
}

func (r *LikeRepository) CountByBlogID(ctx context.Context, blogID primitive.ObjectID) (int64, error) {
    return r.coll.CountDocuments(ctx, bson.M{"blog_id": blogID})
}

func (r *LikeRepository) ensureIndexes(ctx context.Context) error {
    models := []mongo.IndexModel{
        {Keys: bson.D{{Key: "blog_id", Value: 1}, {Key: "user_id", Value: 1}} , Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "created_at", Value: -1}}},
    }
    _, err := r.coll.Indexes().CreateMany(ctx, models)
    return err
}
