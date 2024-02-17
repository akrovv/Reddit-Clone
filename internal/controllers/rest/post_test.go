package rest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"strings"

	"github.com/akrovv/redditclone/internal/domain"
	"github.com/akrovv/redditclone/internal/service"
	"github.com/akrovv/redditclone/internal/service/mocks"
	"github.com/akrovv/redditclone/pkg/logger"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
)

func getTestData() (*domain.User, *domain.Post) {
	user := &domain.User{
		Username: "akro",
		Password: "akroakroakro",
		ID:       "18890d6a9ce0cdbb5a8ce0c1",
	}

	author := &domain.Profile{
		Username: user.Username,
		ID:       user.ID,
	}

	votes := make([]*domain.Vote, 1)
	votes[0] = &domain.Vote{
		User: user.ID,
		Vote: 1,
	}
	postTextRes := &domain.Post{
		Score:            1,
		Views:            1,
		Type:             "text",
		Title:            "akro",
		Author:           author,
		Category:         "music",
		Text:             "my text in post",
		Comments:         make([]*domain.Comment, 0),
		Votes:            votes,
		UpvotePercentage: 100,
	}

	return user, postTextRes
}

func getRequestRecorder(ctx context.Context, isNeedCtx bool, method, path, body string) (*httptest.ResponseRecorder, *http.Request) {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	w := httptest.NewRecorder()

	if isNeedCtx {
		return w, req.WithContext(ctx)
	}

	return w, req
}

func TestAdd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pSrv := mocks.NewMockPostService(ctrl)
	cSrv := mocks.NewMockCommentService(ctrl)
	sSrv := mocks.NewMockSessionService(ctrl)

	postHandler := NewPostHandler(logger.NewLogger(), pSrv, cSrv, sSrv)
	var sess domain.SessionContextKey = "session"
	user, post := getTestData()

	ctx := context.TODO()
	ctx = context.WithValue(ctx,
		sess,
		&domain.Session{
			User: user,
		})

	body := `
	{
		"category": "music",
		"text": "my text in post",
		"title": "akro",
		"type": "text"
	}`

	// OK
	w, req := getRequestRecorder(ctx, true, "POST", "/api/posts", body)
	req.Header.Add("Content-type", "application/json")
	pSrv.EXPECT().Save(post).Return(post, nil)

	postHandler.AddPost(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
		return
	}

	// Pass Content-type
	w, req = getRequestRecorder(ctx, false, "POST", "/api/post", "")
	postHandler.AddPost(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got: %d", w.Code)
		return
	}

	// ReadAll returns error
	req = httptest.NewRequest("POST", "/api/post", &BadReader{})
	w = httptest.NewRecorder()
	req.Header.Add("Content-type", "application/json")

	postHandler.AddPost(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Pass context with session
	w, req = getRequestRecorder(ctx, false, "POST", "/api/posts", `
	{
		"category": "music",
		"text": "my text in post",
		"title": "akro",
		"type": "text"
	}`)
	req.Header.Add("Content-type", "application/json")

	postHandler.AddPost(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got: %d", w.Code)
		return
	}

	// Context with another value
	w, req = getRequestRecorder(context.WithValue(context.TODO(), sess, 1), true, "POST", "/api/posts", `
	{
		"category": "music",
		"text": "my text in post",
		"title": "akro",
		"type": "text"
	}`)
	req.Header.Add("Content-type", "application/json")

	postHandler.AddPost(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Got nil instead of user from getUserFromSession
	w, req = getRequestRecorder(context.WithValue(context.TODO(), sess, &domain.Session{}), true, "POST", "/api/posts", `
	{
		"category": "music",
		"text": "my text in post",
		"title": "akro",
		"type": "text"
	}`)
	req.Header.Add("Content-type", "application/json")

	postHandler.AddPost(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Method's Save returns error
	w, req = getRequestRecorder(ctx, true, "POST", "/api/posts", `
	{
		"category": "music",
		"text": "my text in post",
		"title": "akro",
		"type": "text"
	}`)
	req.Header.Add("Content-type", "application/json")

	pSrv.EXPECT().Save(post).Return(nil, errors.New("some error"))

	postHandler.AddPost(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Writter returns error
	_, req = getRequestRecorder(ctx, true, "POST", "/api/posts", `
	{
		"category": "music",
		"text": "my text in post",
		"title": "akro",
		"type": "text"
	}`)

	req.Header.Add("Content-type", "application/json")
	w = httptest.NewRecorder()

	pSrv.EXPECT().Save(post).Return(post, nil)

	postHandler.AddPost(&BadResponseWriter{ResponseWriter: w}, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}
}

func TestDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pSrv := mocks.NewMockPostService(ctrl)
	cSrv := mocks.NewMockCommentService(ctrl)
	sSrv := mocks.NewMockSessionService(ctrl)

	postHandler := NewPostHandler(logger.NewLogger(), pSrv, cSrv, sSrv)

	var sess domain.SessionContextKey = "session"
	user, _ := getTestData()

	ctx := context.TODO()
	ctx = context.WithValue(ctx,
		sess,
		&domain.Session{
			User: user,
		})

	deletePostDto := &service.DeletePost{PostID: "1"}
	vars := map[string]string{"POST_ID": "1"}

	// OK
	w, req := getRequestRecorder(ctx, true, "DELETE", "/api/post/{POST_ID}", "")
	req = mux.SetURLVars(req, vars)

	pSrv.EXPECT().Delete(deletePostDto).Return(nil)

	postHandler.DeletePost(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
		return
	}

	// Pass POST_ID
	w, req = getRequestRecorder(ctx, true, "DELETE", "/api/post/{POST_ID}", "")
	postHandler.DeletePost(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got: %d", w.Code)
		return
	}

	// Method's Delete returns error
	w, req = getRequestRecorder(ctx, true, "DELETE", "/api/post/{POST_ID}", "")
	req = mux.SetURLVars(req, vars)

	pSrv.EXPECT().Delete(deletePostDto).Return(errors.New("some error"))

	postHandler.DeletePost(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Writter returns error
	w, req = getRequestRecorder(ctx, true, "DELETE", "/api/post/{POST_ID}", "")
	req = mux.SetURLVars(req, vars)
	req.Header.Add("Content-type", "application/json")

	pSrv.EXPECT().Delete(deletePostDto).Return(nil)

	postHandler.DeletePost(&BadResponseWriter{ResponseWriter: w}, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}
}

func TestGetUserPosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pSrv := mocks.NewMockPostService(ctrl)
	cSrv := mocks.NewMockCommentService(ctrl)
	sSrv := mocks.NewMockSessionService(ctrl)

	ctx := context.TODO()
	postHandler := NewPostHandler(logger.NewLogger(), pSrv, cSrv, sSrv)

	getByDto := &service.GetByPost{Category: "author.username", Data: "1", SortField: "created"}
	vars := map[string]string{"USER_LOGIN": "1"}

	// OK
	w, req := getRequestRecorder(ctx, false, "GET", "/u/{USER_LOGIN}", "")
	req = mux.SetURLVars(req, vars)

	pSrv.EXPECT().GetBy(getByDto).Return([]*domain.Post{}, nil)

	postHandler.GetUserPosts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
		return
	}

	// Pass USER_LOGIN
	w, req = getRequestRecorder(ctx, false, "GET", "/u/{USER_LOGIN}", "")
	postHandler.GetUserPosts(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got: %d", w.Code)
		return
	}

	// Method's GetBy returns error
	w, req = getRequestRecorder(ctx, false, "GET", "/u/{USER_LOGIN}", "")
	req = mux.SetURLVars(req, vars)

	pSrv.EXPECT().GetBy(getByDto).Return(nil, errors.New("some error"))

	postHandler.GetUserPosts(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Writter returns error
	w, req = getRequestRecorder(ctx, false, "GET", "/u/{USER_LOGIN}", "")
	req = mux.SetURLVars(req, vars)
	req.Header.Add("Content-type", "application/json")

	pSrv.EXPECT().GetBy(getByDto).Return([]*domain.Post{}, nil)

	postHandler.GetUserPosts(&BadResponseWriter{ResponseWriter: w}, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}
}

func TestShowPosts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pSrv := mocks.NewMockPostService(ctrl)
	cSrv := mocks.NewMockCommentService(ctrl)
	sSrv := mocks.NewMockSessionService(ctrl)

	postHandler := NewPostHandler(logger.NewLogger(), pSrv, cSrv, sSrv)
	ctx := context.TODO()

	// OK
	w, req := getRequestRecorder(ctx, false, "GET", "/api/posts/", "")

	pSrv.EXPECT().Get().Return([]*domain.Post{}, nil)

	postHandler.ShowPosts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
		return
	}

	// Method's Get returns error
	w, req = getRequestRecorder(ctx, false, "GET", "/api/posts/", "")

	pSrv.EXPECT().Get().Return(nil, errors.New("some error"))

	postHandler.ShowPosts(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Writter returns error
	w, req = getRequestRecorder(ctx, false, "GET", "/api/posts/", "")
	req.Header.Add("Content-type", "application/json")

	pSrv.EXPECT().Get().Return([]*domain.Post{}, nil)

	postHandler.ShowPosts(&BadResponseWriter{ResponseWriter: w}, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}
}

func TestShowPostsByFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pSrv := mocks.NewMockPostService(ctrl)
	cSrv := mocks.NewMockCommentService(ctrl)
	sSrv := mocks.NewMockSessionService(ctrl)

	postHandler := NewPostHandler(logger.NewLogger(), pSrv, cSrv, sSrv)
	ctx := context.TODO()

	getByDto := &service.GetByPost{Category: "category", Data: "music", SortField: "score"}
	vars := map[string]string{"CATEGORY_NAME": "music"}

	// OK
	w, req := getRequestRecorder(ctx, false, "GET", "/api/posts/{CATEGORY_NAME}", "")
	req = mux.SetURLVars(req, vars)

	pSrv.EXPECT().GetBy(getByDto).Return([]*domain.Post{}, nil)

	postHandler.ShowPostsByFilter(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
		return
	}

	// Pass CATEGORY_NAME
	w, req = getRequestRecorder(ctx, false, "GET", "/api/posts/{CATEGORY_NAME}", "")
	postHandler.ShowPostsByFilter(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got: %d", w.Code)
		return
	}

	// Method's GetBy returns error
	w, req = getRequestRecorder(ctx, false, "GET", "/api/posts/{CATEGORY_NAME}", "")
	req = mux.SetURLVars(req, vars)

	pSrv.EXPECT().GetBy(getByDto).Return(nil, errors.New("some error"))

	postHandler.ShowPostsByFilter(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Writter returns error
	w, req = getRequestRecorder(ctx, false, "GET", "/api/posts/{CATEGORY_NAME}", "")
	req = mux.SetURLVars(req, vars)
	req.Header.Add("Content-type", "application/json")

	pSrv.EXPECT().GetBy(getByDto).Return([]*domain.Post{}, nil)

	postHandler.ShowPostsByFilter(&BadResponseWriter{ResponseWriter: w}, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}
}

func TestPostDetail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pSrv := mocks.NewMockPostService(ctrl)
	cSrv := mocks.NewMockCommentService(ctrl)
	sSrv := mocks.NewMockSessionService(ctrl)

	postHandler := NewPostHandler(logger.NewLogger(), pSrv, cSrv, sSrv)

	ctx := context.TODO()
	getOnePostDto := &service.GetOnePost{PostID: "1"}
	incViewsDto := &service.IncrViewsPost{PostID: "1"}
	vars := map[string]string{"POST_ID": "1"}

	// OK
	w, req := getRequestRecorder(ctx, false, "GET", "/api/post/{POST_ID}", "")
	req = mux.SetURLVars(req, vars)

	pSrv.EXPECT().GetOne(getOnePostDto).Return(&domain.Post{}, nil)
	pSrv.EXPECT().IncrViews(incViewsDto).Return(nil)

	postHandler.PostDetail(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
		return
	}

	// Pass POST_ID
	w, req = getRequestRecorder(ctx, false, "GET", "/api/post/{POST_ID}", "")
	postHandler.PostDetail(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got: %d", w.Code)
		return
	}

	// Method's GetOne returns error
	w, req = getRequestRecorder(ctx, false, "GET", "/api/post/{POST_ID}", "")
	req = mux.SetURLVars(req, vars)

	pSrv.EXPECT().GetOne(getOnePostDto).Return(nil, errors.New("some error"))

	postHandler.PostDetail(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Method's IncrViews returns error
	w, req = getRequestRecorder(ctx, false, "GET", "/api/post/{POST_ID}", "")
	req = mux.SetURLVars(req, vars)

	pSrv.EXPECT().GetOne(getOnePostDto).Return(&domain.Post{}, nil)
	pSrv.EXPECT().IncrViews(incViewsDto).Return(errors.New("some error"))

	postHandler.PostDetail(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Writter returns error
	w, req = getRequestRecorder(ctx, false, "GET", "/api/post/{POST_ID}", "")
	req = mux.SetURLVars(req, vars)
	req.Header.Add("Content-type", "application/json")

	pSrv.EXPECT().GetOne(getOnePostDto).Return(&domain.Post{}, nil)
	pSrv.EXPECT().IncrViews(incViewsDto).Return(nil)

	postHandler.PostDetail(&BadResponseWriter{ResponseWriter: w}, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}
}

func TestPostVote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pSrv := mocks.NewMockPostService(ctrl)
	cSrv := mocks.NewMockCommentService(ctrl)
	sSrv := mocks.NewMockSessionService(ctrl)

	postHandler := NewPostHandler(logger.NewLogger(), pSrv, cSrv, sSrv)
	var sess domain.SessionContextKey = "session"
	user, _ := getTestData()

	ctx := context.TODO()
	ctx = context.WithValue(ctx,
		sess,
		&domain.Session{
			User: user,
		})

	updatePostDto := &service.UpdateMetricsPost{
		PostID:   "1",
		Inc:      1,
		AuthorID: user.ID,
	}

	getOnePostDto := &service.GetOnePost{PostID: "1"}
	vars := map[string]string{"POST_ID": "1"}

	// OK upVote
	w, req := getRequestRecorder(ctx, true, "GET", "/api/post/{POST_ID}/upvote", "")
	req = mux.SetURLVars(req, vars)

	pSrv.EXPECT().UpdateMetrics(updatePostDto).Return(nil)
	pSrv.EXPECT().GetOne(getOnePostDto).Return(&domain.Post{}, nil)

	postHandler.PostVote(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
		return
	}

	// OK downVote
	updatePostDto.Inc = -1
	w, req = getRequestRecorder(ctx, true, "GET", "/api/post/{POST_ID}/downvote", "")
	req = mux.SetURLVars(req, vars)

	pSrv.EXPECT().UpdateMetrics(updatePostDto).Return(nil)
	pSrv.EXPECT().GetOne(getOnePostDto).Return(&domain.Post{}, nil)

	postHandler.PostVote(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
		return
	}

	// Pass POST_ID
	w, req = getRequestRecorder(ctx, false, "GET", "/api/post/{POST_ID}/downvote", "")
	postHandler.PostVote(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got: %d", w.Code)
		return
	}

	// Pass context
	w, req = getRequestRecorder(ctx, false, "GET", "/api/post/{POST_ID}/downvote", "")
	req = mux.SetURLVars(req, vars)
	postHandler.PostVote(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got: %d", w.Code)
		return
	}

	// Context with another value
	w, req = getRequestRecorder(context.WithValue(context.TODO(), sess, 1), true, "GET", "/api/post/{POST_ID}/downvote", "")
	req = mux.SetURLVars(req, vars)

	postHandler.PostVote(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Got nil instead of user from getUserFromSession
	w, req = getRequestRecorder(context.WithValue(context.TODO(), sess, &domain.Session{}), true, "GET", "/api/post/{POST_ID}/downvote", "")
	req.Header.Add("Content-type", "application/json")

	postHandler.AddPost(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Method's UpdateMetrics returns error
	w, req = getRequestRecorder(ctx, true, "GET", "/api/post/{POST_ID}/downvote", "")
	req = mux.SetURLVars(req, vars)

	pSrv.EXPECT().UpdateMetrics(updatePostDto).Return(errors.New("some error"))

	postHandler.PostVote(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Method's GetOne returns error
	w, req = getRequestRecorder(ctx, true, "GET", "/api/post/{POST_ID}/downvote", "")
	req = mux.SetURLVars(req, vars)

	pSrv.EXPECT().UpdateMetrics(updatePostDto).Return(nil)
	pSrv.EXPECT().GetOne(getOnePostDto).Return(nil, errors.New("some error"))

	postHandler.PostVote(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Writter returns error
	w, req = getRequestRecorder(ctx, true, "GET", "/api/post/{POST_ID}/downvote", "")
	req = mux.SetURLVars(req, vars)
	req.Header.Add("Content-type", "application/json")

	pSrv.EXPECT().UpdateMetrics(updatePostDto).Return(nil)
	pSrv.EXPECT().GetOne(getOnePostDto).Return(&domain.Post{}, nil)

	postHandler.PostVote(&BadResponseWriter{ResponseWriter: w}, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}
}

func TestAddComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pSrv := mocks.NewMockPostService(ctrl)
	cSrv := mocks.NewMockCommentService(ctrl)
	sSrv := mocks.NewMockSessionService(ctrl)

	postHandler := NewPostHandler(logger.NewLogger(), pSrv, cSrv, sSrv)

	body := `{"comment": "my comment!"}`

	var sess domain.SessionContextKey = "session"
	user, _ := getTestData()
	getOnePostDto := &service.GetOnePost{PostID: "1"}
	addCommentDto := &service.AddComment{
		User:   user,
		Body:   "my comment!",
		PostID: "1",
	}

	ctx := context.TODO()
	ctx = context.WithValue(ctx,
		sess,
		&domain.Session{
			User: user,
		})

	vars := map[string]string{"POST_ID": "1"}

	// OK
	w, req := getRequestRecorder(ctx, true, "POST", "/api/post/{POST_ID}", body)
	req.Header.Add("Content-type", "application/json")
	req = mux.SetURLVars(req, vars)

	cSrv.EXPECT().Add(addCommentDto).Return(nil)
	pSrv.EXPECT().GetOne(getOnePostDto).Return(&domain.Post{}, nil)

	postHandler.AddComment(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
		return
	}

	// Pass Content-type
	w, req = getRequestRecorder(ctx, false, "POST", "/api/post/{POST_ID}", "")
	postHandler.AddComment(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got: %d", w.Code)
		return
	}

	// ReadAll returns error
	req = httptest.NewRequest("POST", "/api/post/{POST_ID}", &BadReader{})
	w = httptest.NewRecorder()
	req.Header.Add("Content-type", "application/json")

	postHandler.AddComment(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Pass POST_ID
	w, req = getRequestRecorder(ctx, false, "POST", "/api/post/{POST_ID}", `{"comment": "my comment!"}`)
	req.Header.Add("Content-type", "application/json")
	postHandler.AddComment(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got: %d", w.Code)
		return
	}

	// Pass context
	w, req = getRequestRecorder(ctx, false, "POST", "/api/post/{POST_ID}", `{"comment": "my comment!"}`)
	req = mux.SetURLVars(req, vars)
	req.Header.Add("Content-type", "application/json")
	postHandler.AddComment(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Context with another value
	w, req = getRequestRecorder(context.WithValue(context.TODO(), sess, 1), true, "POST", "/api/post/{POST_ID}", `{"comment": "my comment!"}`)
	req = mux.SetURLVars(req, vars)
	req.Header.Add("Content-type", "application/json")

	postHandler.AddComment(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Got nil instead of user from getUserFromSession
	w, req = getRequestRecorder(context.WithValue(context.TODO(), sess, &domain.Session{}), true, "POST", "/api/post/{POST_ID}", `{"comment": "my comment!"}`)
	req = mux.SetURLVars(req, vars)
	req.Header.Add("Content-type", "application/json")

	postHandler.AddComment(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Method's Add returns error
	w, req = getRequestRecorder(ctx, true, "POST", "/api/post/{POST_ID}", `{"comment": "my comment!"}`)
	req = mux.SetURLVars(req, vars)
	req.Header.Add("Content-type", "application/json")

	cSrv.EXPECT().Add(addCommentDto).Return(errors.New("some error"))

	postHandler.AddComment(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Method's GetOne returns error
	w, req = getRequestRecorder(ctx, true, "POST", "/api/post/{POST_ID}", body)
	req = mux.SetURLVars(req, vars)
	req.Header.Add("Content-type", "application/json")

	cSrv.EXPECT().Add(addCommentDto).Return(nil)
	pSrv.EXPECT().GetOne(getOnePostDto).Return(nil, errors.New("some error"))

	postHandler.AddComment(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Writter returns error
	w, req = getRequestRecorder(ctx, true, "POST", "/api/post/{POST_ID}", body)
	req = mux.SetURLVars(req, vars)
	req.Header.Add("Content-type", "application/json")

	cSrv.EXPECT().Add(addCommentDto).Return(nil)
	pSrv.EXPECT().GetOne(getOnePostDto).Return(&domain.Post{}, nil)

	postHandler.AddComment(&BadResponseWriter{ResponseWriter: w}, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}
}

func TestPostDelComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pSrv := mocks.NewMockPostService(ctrl)
	cSrv := mocks.NewMockCommentService(ctrl)
	sSrv := mocks.NewMockSessionService(ctrl)

	postHandler := NewPostHandler(logger.NewLogger(), pSrv, cSrv, sSrv)

	var sess domain.SessionContextKey = "session"
	user, _ := getTestData()
	ctx := context.Background()
	ctx = context.WithValue(ctx,
		sess,
		&domain.Session{
			User: user,
		})

	deleteCommentDto := &service.DeleteComment{PostID: "1", CommentID: "1"}
	getOnePostDto := &service.GetOnePost{PostID: "1"}
	vars := map[string]string{"POST_ID": "1", "COMMENT_ID": "1"}

	// OK
	w, req := getRequestRecorder(ctx, true, "DELETE", "/api/post/{POST_ID}/{COMMENT_ID}", "")
	req = mux.SetURLVars(req, vars)

	cSrv.EXPECT().Delete(deleteCommentDto).Return(nil)
	pSrv.EXPECT().GetOne(getOnePostDto).Return(&domain.Post{}, nil)

	postHandler.DeleteComment(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got: %d", w.Code)
		return
	}

	// Pass POST_ID
	w, req = getRequestRecorder(ctx, true, "DELETE", "/api/post/{POST_ID}/{COMMENT_ID}", "")
	postHandler.DeleteComment(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got: %d", w.Code)
		return
	}

	// Pass COMMENT_ID
	onlyPost := map[string]string{"POST_ID": "1"}
	w, req = getRequestRecorder(ctx, true, "DELETE", "/api/post/{POST_ID}/{COMMENT_ID}", "")
	req = mux.SetURLVars(req, onlyPost)
	postHandler.DeleteComment(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got: %d", w.Code)
		return
	}

	// Method's Delete returns error
	w, req = getRequestRecorder(ctx, true, "DELETE", "/api/post/{POST_ID}/{COMMENT_ID}", "")
	req = mux.SetURLVars(req, vars)

	cSrv.EXPECT().Delete(deleteCommentDto).Return(errors.New("some error"))

	postHandler.DeleteComment(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Method's GetOne returns error
	w, req = getRequestRecorder(ctx, true, "DELETE", "/api/post/{POST_ID}/{COMMENT_ID}", "")
	req = mux.SetURLVars(req, vars)

	cSrv.EXPECT().Delete(deleteCommentDto).Return(nil)
	pSrv.EXPECT().GetOne(getOnePostDto).Return(nil, errors.New("some error"))

	postHandler.DeleteComment(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}

	// Writter returns error
	w, req = getRequestRecorder(ctx, true, "DELETE", "/api/post/{POST_ID}/{COMMENT_ID}", "")
	req = mux.SetURLVars(req, vars)
	req.Header.Add("Content-type", "application/json")

	cSrv.EXPECT().Delete(deleteCommentDto).Return(nil)
	pSrv.EXPECT().GetOne(getOnePostDto).Return(&domain.Post{}, nil)

	postHandler.DeleteComment(&BadResponseWriter{ResponseWriter: w}, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got: %d", w.Code)
		return
	}
}
