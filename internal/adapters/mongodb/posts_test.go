package mongodb

import (
	"context"
	"reflect"
	"testing"

	"github.com/akrovv/redditclone/internal/domain"
	"github.com/akrovv/redditclone/pkg/generator"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

type TestCases struct {
	user      *domain.User
	post      *domain.Post
	id        []string
	threeData [3]string
}

func TestPostSave(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	ctx := context.Background()
	user := &domain.User{
		Username: "akro",
	}

	post := &domain.Post{
		Type:   "text",
		Author: &domain.Profile{Username: "akro"},
	}

	testNegative := []TestCases{
		{user, post, []string{}, [3]string{}},
	}

	mt.Run("Save", func(mt *mtest.T) {
		repo := NewPostStorage(ctx, mt.Client)

		mt.AddMockResponses(mtest.CreateSuccessResponse(
			bson.D{
				{Key: "ok", Value: 1},
			}...))

		_, err := repo.Save(post)

		if err != nil {
			t.Errorf("unexpected error: %s", err)
			return
		}

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   1,
			Code:    11000,
			Message: "duplicate key error",
		}))

		for _, test := range testNegative {
			_, err = repo.Save(test.post)

			if err == nil {
				t.Errorf("expected error, got nil")
				return
			}
		}
	})
}

func TestGetOne(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	id := generator.GenerateNewID("TextDatatext")

	mt.Run("GetOne", func(mt *mtest.T) {
		expectPost := &domain.Post{
			Title:    "Text Data",
			ID:       id,
			Category: "text",
		}

		find := mtest.CreateCursorResponse(1, "homework.users", mtest.FirstBatch, bson.D{
			{Key: "ok", Value: 1},
			{Key: "Title", Value: "Text Data"},
			{Key: "ID", Value: id},
			{Key: "Category", Value: "text"},
		})
		killCursors := mtest.CreateCursorResponse(0, "homework.users", mtest.NextBatch)

		mt.AddMockResponses(find, killCursors)

		ctx := context.Background()
		repo := NewPostStorage(ctx, mt.Client)

		post, err := repo.GetOne(id)

		if err != nil {
			t.Errorf("unexpected error: %s", err)
			return
		}

		if !reflect.DeepEqual(expectPost, post) {
			t.Errorf("results not match, want %v, have %v", expectPost, post)
			return
		}

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   1,
			Code:    11000,
			Message: "not exists",
		}))

		_, err = repo.GetOne(id)

		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}

		_, err = repo.GetOne("")

		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}
	})
}

func TestGet(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	ctx := context.Background()

	mt.Run("Get", func(mt *mtest.T) {
		id1 := generator.GenerateNewID("test1text")
		id2 := generator.GenerateNewID("test2link")
		id3 := generator.GenerateNewID("test3text")

		expectPosts := []*domain.Post{
			{
				Title: "Test 1",
				ID:    id1,
				Type:  "text",
				Score: 1,
			},
			{
				Title: "Test 3",
				ID:    id3,
				Type:  "text",
				Score: 2,
			},
			{
				Title: "Test 2",
				ID:    id2,
				Type:  "link",
				Score: 3,
			},
		}

		findOne := mtest.CreateCursorResponse(1, "homework.users", mtest.FirstBatch, bson.D{
			{Key: "ok", Value: 1},
			{Key: "title", Value: "Test 1"},
			{Key: "id", Value: id1},
			{Key: "type", Value: "text"},
			{Key: "score", Value: 1},
		})

		findTwo := mtest.CreateCursorResponse(1, "homework.users", mtest.NextBatch, bson.D{
			{Key: "ok", Value: 1},
			{Key: "title", Value: "Test 2"},
			{Key: "id", Value: id2},
			{Key: "type", Value: "link"},
			{Key: "score", Value: 3},
		})

		findThree := mtest.CreateCursorResponse(1, "homework.users", mtest.NextBatch, bson.D{
			{Key: "ok", Value: 1},
			{Key: "title", Value: "Test 3"},
			{Key: "id", Value: id3},
			{Key: "type", Value: "text"},
			{Key: "score", Value: 2},
		})

		killCursors := mtest.CreateCursorResponse(0, "homework.users", mtest.NextBatch)
		mt.AddMockResponses(findOne, findThree, findTwo, killCursors)

		repo := NewPostStorage(ctx, mt.Client)
		posts, err := repo.Get()

		if err != nil {
			t.Errorf("unexpected error: %s", err)
			return
		}

		if !reflect.DeepEqual(expectPosts, posts) {
			t.Errorf("results not match, want %v, have %v", expectPosts, posts)
		}

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   1,
			Code:    11000,
			Message: "not exists",
		}))

		_, err = repo.Get()

		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}

		mt.AddMockResponses(findOne)
		_, err = repo.Get()
		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}
	})
}

func TestGetBy(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	ctx := context.Background()
	id2 := generator.GenerateNewID("test2link")
	id3 := generator.GenerateNewID("test3text")

	testNegative := []TestCases{
		{nil, nil, []string{}, [3]string{"category", "news", "score"}},
		{nil, nil, []string{}, [3]string{"", "news", "score"}},
		{nil, nil, []string{}, [3]string{"category", "", "score"}},
		{nil, nil, []string{}, [3]string{"category", "news", ""}},
	}

	mt.Run("GetBy", func(mt *mtest.T) {
		repo := NewPostStorage(ctx, mt.Client)
		expectPosts := []*domain.Post{
			{
				Title:    "Test 3",
				ID:       id3,
				Type:     "text",
				Category: "news",
				Score:    2,
			},
			{
				Title:    "Test 2",
				ID:       id2,
				Type:     "link",
				Category: "news",
				Score:    3,
			},
		}

		findThree := mtest.CreateCursorResponse(1, "homework.users", mtest.FirstBatch, bson.D{
			{Key: "ok", Value: 1},
			{Key: "title", Value: "Test 3"},
			{Key: "id", Value: id3},
			{Key: "type", Value: "text"},
			{Key: "category", Value: "news"},
			{Key: "score", Value: 2},
		})

		findTwo := mtest.CreateCursorResponse(1, "homework.users", mtest.NextBatch, bson.D{
			{Key: "ok", Value: 1},
			{Key: "title", Value: "Test 2"},
			{Key: "id", Value: id2},
			{Key: "type", Value: "link"},
			{Key: "category", Value: "news"},
			{Key: "score", Value: 3},
		})

		killCursors := mtest.CreateCursorResponse(0, "homework.users", mtest.NextBatch)
		mt.AddMockResponses(findThree, findTwo, killCursors)

		posts, err := repo.GetBy("category", "news", "score")

		if err != nil {
			t.Errorf("unexpected error: %s", err)
			return
		}

		if !reflect.DeepEqual(expectPosts, posts) {
			t.Errorf("results not match, want %v, have %v", expectPosts, posts)
		}

		// negative cases
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   1,
			Code:    11000,
			Message: "not exists",
		}))

		for _, test := range testNegative {
			_, err = repo.GetBy(test.threeData[0], test.threeData[1], test.threeData[2])

			if err == nil {
				t.Errorf("expected error, got nil")
				return
			}
		}

		mt.AddMockResponses(findThree)
		_, err = repo.GetBy("test", "test", "test")
		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}
	})
}

func TestUpdateMetrics(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	ctx := context.Background()

	postID := generator.GenerateNewID("TextDatatext")
	userID := generator.GenerateNewIDByMD("akro")

	dbPost := mtest.CreateCursorResponse(0, "homework.users", mtest.FirstBatch, bson.D{
		{Key: "totalVotes", Value: 10},
		{Key: "voteCount", Value: 5},
		{Key: "positiveVoteCount", Value: 3},
	})

	user := &domain.Profile{
		Username: "akro",
		ID:       userID,
	}

	post := &domain.Post{
		Author: user,
	}

	mt.Run("Update unVote", func(mt *mtest.T) {
		repo := NewPostStorage(ctx, mt.Client)

		// OK
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 1}, {Key: "nModified", Value: 1}}, dbPost, bson.D{{Key: "ok", Value: 1}, {Key: "nModified", Value: 1}})

		err := repo.UpdateMetrics(postID, 0, post.Author.ID)

		if err != nil {
			t.Errorf("unexpected error: %s", err)
			return
		}

		// updateOne error
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 0}})
		err = repo.UpdateMetrics(postID, 0, post.Author.ID)

		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}

		// updateScorePercent: updateOne: affected 0 rows
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 1}, {Key: "nModified", Value: 1}}, dbPost, bson.D{{Key: "ok", Value: 1}})
		err = repo.UpdateMetrics(postID, 0, post.Author.ID)

		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}

		// updateScorePercent: another error in UpdateOne
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 1}, {Key: "nModified", Value: 1}}, dbPost, bson.D{{Key: "ok", Value: 0}})
		err = repo.UpdateMetrics(postID, 0, post.Author.ID)

		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}

		// updateScorePercent: no matching documents
		nonMatchingData := mtest.CreateCursorResponse(0, "homework.users", mtest.NextBatch)
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 1}, {Key: "nModified", Value: 1}}, nonMatchingData)
		err = repo.UpdateMetrics(postID, 0, post.Author.ID)

		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}

		// updateScorePercent: error decode
		dec := mtest.CreateCursorResponse(0, "homework.users", mtest.FirstBatch, bson.D{{Key: "totalVotes", Value: "abc"}})
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 1}, {Key: "nModified", Value: 1}}, dec)
		err = repo.UpdateMetrics(postID, 0, post.Author.ID)

		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}

		// updateScorePercent: empty Aggregate
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 1}, {Key: "nModified", Value: 1}}, bson.D{}, bson.D{{Key: "ok", Value: 1}, {Key: "nModified", Value: 1}})
		err = repo.UpdateMetrics(postID, 0, post.Author.ID)

		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}
	})

	mt.Run("Update upVote&downVote", func(mt *mtest.T) {
		repo := NewPostStorage(ctx, mt.Client)

		// OK
		mt.AddMockResponses(dbPost, bson.D{{Key: "ok", Value: 1}, {Key: "nModified", Value: 1}}, dbPost, bson.D{{Key: "ok", Value: 1}, {Key: "nModified", Value: 1}})

		err := repo.UpdateMetrics(postID, 1, post.Author.ID)

		if err != nil {
			t.Errorf("unexpected error: %s", err)
			return
		}

		// Error setVote
		mt.AddMockResponses(dbPost, bson.D{{Key: "ok", Value: 0}})

		err = repo.UpdateMetrics(postID, 1, post.Author.ID)

		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}

		// setVote: in if: no responses remaining
		nonMatchingData := mtest.CreateCursorResponse(0, "homework.users", mtest.FirstBatch)
		mt.AddMockResponses(nonMatchingData)

		err = repo.UpdateMetrics(postID, 1, post.Author.ID)

		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}

		// setVote: in if: affected 0 rows
		mt.AddMockResponses(nonMatchingData, bson.D{{Key: "ok", Value: 1}, {Key: "nModified", Value: 0}})

		err = repo.UpdateMetrics(postID, 1, post.Author.ID)

		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}

		// setVote: in else: findOne error
		mt.AddMockResponses()

		err = repo.UpdateMetrics(postID, 1, post.Author.ID)

		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}

		// setVote: in else: affected 0 rows
		mt.AddMockResponses(dbPost, bson.D{{Key: "ok", Value: 1}})

		err = repo.UpdateMetrics(postID, 1, post.Author.ID)

		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}

		// setVote: in else: affected 0 rows
		mt.AddMockResponses(dbPost, bson.D{{Key: "ok", Value: 1}})

		err = repo.UpdateMetrics(postID, 1, post.Author.ID)

		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}
	})

	mt.Run("Common error", func(mt *mtest.T) {
		repo := NewPostStorage(ctx, mt.Client)

		// incorrect increment
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 1}, {Key: "nModified", Value: 1}}, bson.D{}, bson.D{{Key: "ok", Value: 1}, {Key: "nModified", Value: 1}})
		err := repo.UpdateMetrics(postID, -100, post.Author.ID)
		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}
	})
}

func TestIncrViews(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	ctx := context.Background()

	mt.Run("IncrViews", func(mt *mtest.T) {
		repo := NewPostStorage(ctx, mt.Client)

		// OK
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 1}, {Key: "nModified", Value: 1}})

		err := repo.IncrViews("1")

		if err != nil {
			t.Errorf("unexpected error: %s", err)
			return
		}

		// updateOne error
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 0}})

		err = repo.IncrViews("1")

		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}

		// affected 0 rows
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 1}, {Key: "nModified", Value: 0}})

		err = repo.IncrViews("1")

		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}
	})
}

func TestDel(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	ctx := context.Background()

	mt.Run("IncrViews", func(mt *mtest.T) {
		repo := NewPostStorage(ctx, mt.Client)

		// OK
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 1}, {Key: "n", Value: 1}})

		err := repo.Delete("1")

		if err != nil {
			t.Errorf("unexpected error: %s", err)
			return
		}

		// updateOne error
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 0}})

		err = repo.Delete("1")

		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}

		// affected 0 rows
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 1}, {Key: "n", Value: 0}})

		err = repo.Delete("1")

		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}

		// empty postID
		err = repo.Delete("")

		if err == nil {
			t.Errorf("expected error, got nil")
			return
		}
	})
}
