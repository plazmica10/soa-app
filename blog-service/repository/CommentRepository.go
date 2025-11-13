package repository

import (
    "context"
    "log"
    "time"

    "blog-service/model"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type CommentRepository struct {
    coll *mongo.Collection
}

func NewCommentRepository(db *mongo.Database) *CommentRepository {
    r := &CommentRepository{coll: db.Collection("comments")}
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := r.ensureIndexes(ctx); err != nil {
        log.Printf("comment repository: ensure indexes: %v", err)
    }
    return r
}

func (r *CommentRepository) Create(ctx context.Context, c *model.Comment) error {
    if c.CreatedAt.IsZero() {
        c.CreatedAt = time.Now().UTC()
    }
    _, err := r.coll.InsertOne(ctx, c)
    return err
}

func (r *CommentRepository) GetByBlogID(ctx context.Context, blogID interface{}) ([]model.Comment, error) {
    filter := bson.M{"blog_id": blogID}
    opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})
    cur, err := r.coll.Find(ctx, filter, opts)
    if err != nil {
        return nil, err
    }
    defer cur.Close(ctx)
    var res []model.Comment
    for cur.Next(ctx) {
        var c model.Comment
        if err := cur.Decode(&c); err != nil {
            return nil, err
        }
        res = append(res, c)
    }
    return res, cur.Err()
}

func (r *CommentRepository) ensureIndexes(ctx context.Context) error {
    models := []mongo.IndexModel{
        {Keys: bson.D{{Key: "blog_id", Value: 1}}},
        {Keys: bson.D{{Key: "created_at", Value: -1}}},
    }
    _, err := r.coll.Indexes().CreateMany(ctx, models)
    return err
}

// UpdateText updates the text and last_edited_at of a comment and returns the updated comment.
func (r *CommentRepository) UpdateText(ctx context.Context, commentID primitive.ObjectID, text string) (*model.Comment, error) {
    now := time.Now().UTC()
    _, err := r.coll.UpdateByID(ctx, commentID, bson.M{"$set": bson.M{"text": text, "last_edited_at": now}})
    if err != nil {
        return nil, err
    }
    var c model.Comment
    if err := r.coll.FindOne(ctx, bson.M{"_id": commentID}).Decode(&c); err != nil {
        return nil, err
    }
    return &c, nil
}
