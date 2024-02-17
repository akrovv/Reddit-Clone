package mongodb

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/akrovv/redditclone/internal/domain"
	"github.com/akrovv/redditclone/pkg/generator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type postStorage struct {
	DB  *mongo.Collection
	ctx context.Context
}

func NewPostStorage(ctx context.Context, db *mongo.Client) *postStorage {
	collection := db.Database("post").Collection("posts")
	return &postStorage{DB: collection, ctx: ctx}
}

func (p postStorage) Save(post *domain.Post) (*domain.Post, error) {
	dataForID := strings.Trim(post.Title+post.Author.Username+post.Category, " ")
	post.ID = generator.GenerateNewID(dataForID)
	post.Created = time.Now()

	newPost := bson.M{
		"_id":              primitive.NewObjectID(),
		"id":               post.ID,
		"score":            post.Score,
		"views":            post.Views,
		"type":             post.Type,
		"title":            post.Title,
		"url":              post.URL,
		"author":           post.Author,
		"category":         post.Category,
		"text":             post.Text,
		"votes":            post.Votes,
		"comments":         post.Comments,
		"created":          post.Created,
		"upvotePercentage": post.UpvotePercentage,
	}

	_, err := p.DB.InsertOne(p.ctx, newPost)

	if err != nil {
		return nil, errors.New("can't insert into db")
	}

	return post, nil
}

func (p postStorage) GetOne(id string) (*domain.Post, error) {
	post := &domain.Post{}
	err := p.DB.FindOne(p.ctx, bson.M{"id": id}).Decode(post)

	if err != nil {
		return nil, errors.New("can't read data in *Post{}")
	}

	return post, nil
}

func (p postStorage) Get() ([]*domain.Post, error) {
	posts := []*domain.Post{}

	options := options.Find().SetSort(bson.D{{Key: "score", Value: -1}})
	c, err := p.DB.Find(p.ctx, bson.M{}, options)

	if err != nil {
		return nil, errors.New("can't get posts from db")
	}

	err = c.All(p.ctx, &posts)

	if err != nil {
		return nil, errors.New("can't read posts in []*Post{}")
	}

	return posts, nil
}

func (p postStorage) GetBy(category, data, sortField string) ([]*domain.Post, error) {
	posts := []*domain.Post{}
	options := options.Find().SetSort(bson.D{{Key: sortField, Value: -1}})
	c, err := p.DB.Find(p.ctx, bson.M{category: data}, options)

	if err != nil {
		return nil, errors.New("can't get posts from db")
	}

	err = c.All(p.ctx, &posts)

	if err != nil {
		return nil, errors.New("can't read posts in []*Post{}")
	}

	return posts, nil
}

func (p postStorage) UpdateMetrics(postID string, inc int8, authorID string) error {
	postFilter := bson.M{"id": postID}

	switch {
	case inc == 0:
		_, err := p.DB.UpdateOne(p.ctx, postFilter, bson.M{"$pull": bson.M{"votes": bson.M{"user": authorID}}})

		if err != nil {
			return err
		}

	case inc == 1 || inc == -1:
		err := p.setVote(authorID, postID, inc)

		if err != nil {
			return err
		}

	default:
		return errors.New("unknown inc")
	}

	err := p.updateScorePercent(postID)

	if err != nil {
		return err
	}

	return nil
}

func (p postStorage) updateScorePercent(postID string) error {
	pipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.D{{Key: "id", Value: postID}}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "totalVotes", Value: bson.D{{Key: "$sum", Value: "$votes.vote"}}},
			{Key: "voteCount", Value: bson.D{{Key: "$size", Value: "$votes"}}},
			{Key: "positiveVoteCount", Value: bson.D{{Key: "$size", Value: bson.D{{Key: "$filter", Value: bson.D{
				{Key: "input", Value: "$votes"},
				{Key: "as", Value: "vote"},
				{Key: "cond", Value: bson.D{{Key: "$gt", Value: bson.A{"$$vote.vote", 0}}}},
			}}}}}},
		}},
		}}

	cursor, err := p.DB.Aggregate(p.ctx, pipeline)

	if err != nil {
		return errors.New("can't aggregate func")
	}

	var result struct {
		TotalVotes        int `bson:"totalVotes"`
		VoteCount         int `bson:"voteCount"`
		PositiveVoteCount int `bson:"positiveVoteCount"`
	}

	if cursor.Next(p.ctx) {
		if err := cursor.Decode(&result); err != nil {
			return errors.New("can't decode into struct")
		}

		var percent = 0

		if result.VoteCount != 0 {
			percent = int((float32(result.PositiveVoteCount) / float32(result.VoteCount)) * 100.0)
		}

		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "score", Value: result.TotalVotes},
				{Key: "upvotePercentage", Value: percent},
			}},
		}

		filter := bson.D{{Key: "id", Value: postID}}

		updateResult, err := p.DB.UpdateOne(p.ctx, filter, update)

		if err != nil {
			return err
		} else if updateResult.ModifiedCount == 0 {
			return errors.New("affected 0 rows")
		}
	} else {
		return errors.New("no matching documents")
	}

	return nil
}

func (p postStorage) setVote(userID, postID string, inc int8) error {
	err := p.DB.FindOne(p.ctx, bson.M{"id": postID, "votes.user": userID}).Err()

	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}

	if err == mongo.ErrNoDocuments {
		res, err := p.DB.UpdateOne(p.ctx, bson.M{"id": postID}, bson.M{"$push": bson.M{"votes": bson.M{"user": userID, "vote": inc}}})

		if err != nil {
			return err
		} else if res.ModifiedCount == 0 {
			return errors.New("affected 0 rows")
		}
	} else {
		res, err := p.DB.UpdateOne(p.ctx, bson.M{"id": postID, "votes.user": userID}, bson.M{"$set": bson.M{"votes.$.vote": inc}})

		if err != nil {
			return err
		} else if res.ModifiedCount == 0 {
			return errors.New("affected 0 rows")
		}
	}

	return nil
}

func (p postStorage) IncrViews(postID string) error {
	res, err := p.DB.UpdateOne(p.ctx, bson.M{"id": postID}, bson.M{"$inc": bson.M{"views": 1}})

	if err != nil {
		return err
	} else if res.ModifiedCount == 0 {
		return errors.New("affected 0 rows")
	}

	return nil
}

func (p postStorage) Delete(postID string) error {
	res, err := p.DB.DeleteOne(p.ctx, bson.M{"id": postID})

	if err != nil {
		return err
	} else if res.DeletedCount == 0 {
		return errors.New("affected 0 rows")
	}

	return nil
}
