package mongodb

import (
	"context"
	"errors"
	"strings"

	"time"

	"github.com/akrovv/redditclone/internal/domain"
	"github.com/akrovv/redditclone/pkg/generator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type commentStorage struct {
	ctx context.Context
	DB  *mongo.Collection
}

func NewCommentStorage(ctx context.Context, db *mongo.Client) *commentStorage {
	collection := db.Database("post").Collection("posts")
	return &commentStorage{ctx: ctx, DB: collection}
}

func (c commentStorage) Add(author *domain.Profile, body, postID string) error {
	t := time.Now()
	dataForID := strings.Trim(body+author.Username+author.ID+t.String(), " ")
	id := generator.GenerateNewID(dataForID)

	newComment := bson.M{
		"id":      id,
		"author":  author,
		"body":    body,
		"created": t,
	}

	postFilter := bson.M{"id": postID}
	result, err := c.DB.UpdateOne(c.ctx, postFilter, bson.M{"$push": bson.M{"comments": newComment}})

	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return errors.New("affected 0 rows")
	}

	return nil
}

func (c commentStorage) Delete(postID, commentID string) error {
	filter := bson.M{"id": postID}
	result, err := c.DB.UpdateOne(c.ctx, filter, bson.M{"$pull": bson.M{"comments": bson.M{"id": commentID}}})

	if err != nil {
		return err
	}

	if result.ModifiedCount == 0 {
		return errors.New("affected 0 rows")
	}

	return nil
}
