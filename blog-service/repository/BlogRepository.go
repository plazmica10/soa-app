package repository

import (
    "context"
    "log"
    "time"

    "blog-service/model"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type BlogRepository struct {
    coll *mongo.Collection
}

func NewBlogRepository(db *mongo.Database) *BlogRepository {
    r := &BlogRepository{coll: db.Collection("blogs")}
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := r.ensureIndexes(ctx); err != nil {
        log.Printf("blog repository: ensure indexes: %v", err)
    }
    return r
}

func (r *BlogRepository) Create(ctx context.Context, b *model.Blog) error {
    if b.CreatedAt.IsZero() {
        b.CreatedAt = time.Now().UTC()
    }
    _, err := r.coll.InsertOne(ctx, b)
    return err
}

func (r *BlogRepository) GetAll(ctx context.Context) ([]model.Blog, error) {
    cur, err := r.coll.Find(ctx, bson.D{})
    if err != nil {
        return nil, err
    }
    defer cur.Close(ctx)
    var res []model.Blog
    for cur.Next(ctx) {
        var b model.Blog
        if err := cur.Decode(&b); err != nil {
            return nil, err
        }
        res = append(res, b)
    }
    return res, cur.Err()
}

func (r *BlogRepository) GetByID(ctx context.Context, id string) (*model.Blog, error) {
    var b model.Blog
    err := r.coll.FindOne(ctx, bson.M{"id": id}).Decode(&b)
    return &b, err
}

func (r *BlogRepository) ensureIndexes(ctx context.Context) error {
    models := []mongo.IndexModel{
        {Keys: bson.D{{Key: "id", Value: 1}}, Options: options.Index().SetUnique(true)},
        {Keys: bson.D{{Key: "created_at", Value: -1}}},
    }
    _, err := r.coll.Indexes().CreateMany(ctx, models)
    return err
}
