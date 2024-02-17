package mongodb

import (
	"context"
	"testing"

	"github.com/akrovv/redditclone/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestAddComment(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	ctx := context.Background()
	author := &domain.Profile{
		Username: "akro",
	}

	body := "good review"
	id := "1"

	mt.Run("Add", func(mt *mtest.T) {
		repo := NewCommentStorage(ctx, mt.Client)

		// OK
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.D{
				{Key: "ok", Value: 1},
				{Key: "nModified", Value: 1},
			}...))

		err := repo.Add(author, body, id)

		if err != nil {
			t.Errorf("unexpected error: %s", err)
			return
		}

		// Zero rows affected
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.D{
				{Key: "ok", Value: 1},
			}...))

		err = repo.Add(author, body, id)

		if err == nil {
			t.Error("expected error, got nil")
			return
		}

		// Add error
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   1,
			Code:    11000,
			Message: "some error",
		}))

		err = repo.Add(author, body, id)

		if err == nil {
			t.Error("expected error, got nil")
			return
		}
	})
}

func TestDeleteComment(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	ctx := context.Background()
	postID := "1"
	commentID := "1"

	mt.Run("Delete", func(mt *mtest.T) {
		repo := NewCommentStorage(ctx, mt.Client)

		// OK
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.D{
				{Key: "ok", Value: 1},
				{Key: "nModified", Value: 1},
			}...))

		err := repo.Delete(postID, commentID)

		if err != nil {
			t.Errorf("unexpected error: %s", err)
			return
		}

		// Zero rows affected
		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.D{
				{Key: "ok", Value: 1},
			}...))

		err = repo.Delete(postID, commentID)

		if err == nil {
			t.Error("expected error, got nil")
			return
		}

		// Delete error
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   1,
			Code:    11000,
			Message: "some error",
		}))

		err = repo.Delete(postID, commentID)

		if err == nil {
			t.Error("expected error, got nil")
			return
		}
	})
}
