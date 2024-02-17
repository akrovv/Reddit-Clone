package rest

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"strings"

	"github.com/akrovv/redditclone/internal/domain"
	"github.com/akrovv/redditclone/internal/service"
	jsontransfer "github.com/akrovv/redditclone/pkg/jsonTransfer"
	"github.com/akrovv/redditclone/pkg/logger"
	"github.com/gorilla/mux"
)

type postHandler struct {
	logger         logger.Logger
	postService    PostService
	commentService CommentService
	sessionService SessionService
}

func NewPostHandler(logger logger.Logger, postService PostService, commentService CommentService, sessionService SessionService) *postHandler {
	return &postHandler{logger: logger, postService: postService, commentService: commentService, sessionService: sessionService}
}

const contentType = "application/json"

func (h postHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	pi, inPost := vars["POST_ID"]

	if !inPost {
		h.logger.Info("post not found")
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}

	ci, inComment := vars["COMMENT_ID"]

	if !inComment {
		h.logger.Info("post not found")
		http.Error(w, "comment not found", http.StatusNotFound)
		return
	}

	deleteCommentDto := &service.DeleteComment{PostID: pi, CommentID: ci}
	err := h.commentService.Delete(deleteCommentDto)

	if err != nil {
		h.logger.Infof("can't delete post: %w", err)
		http.Error(w, "can't delete post", http.StatusInternalServerError)
		return
	}

	getOnePostDto := &service.GetOnePost{PostID: pi}
	postWithNoComment, err := h.postService.GetOne(getOnePostDto)

	if err != nil {
		h.logger.Infof("can't get post: %w", err)
		http.Error(w, "can't get post", http.StatusInternalServerError)
		return
	}

	postJSON, err := jsontransfer.GetJSON(postWithNoComment)

	if err != nil {
		h.logger.Infof("can't marshall post to json: %w", err)
		http.Error(w, "can't marshall post to json", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-type", contentType)
	_, err = w.Write(postJSON)

	if err != nil {
		h.logger.Infof("server can't write: %w", err)
		http.Error(w, "server can't write", http.StatusInternalServerError)
		return
	}
}

func (h postHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	jsonHeader := r.Header.Get("Content-type")
	if jsonHeader != contentType {
		h.logger.Infof("not found application/json header")
		http.Error(w, "not found application/json header", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		h.logger.Infof("can't read form: %w", err)
		http.Error(w, "can't read form", http.StatusInternalServerError)
		return
	}

	vars := mux.Vars(r)
	pi, in := vars["POST_ID"]

	if !in {
		h.logger.Info("post not found")
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}

	ctx := r.Context()
	sessCtx := ctx.Value(domain.SessionContextKey("session"))
	if sessCtx == nil {
		h.logger.Info("can't find session")
		http.Error(w, "can't find session", http.StatusInternalServerError)
		return
	}

	user, err := getUserFromSession(sessCtx)

	if err != nil {
		h.logger.Infof("can't convert to session: %w", err)
		http.Error(w, "can't convert to session", http.StatusInternalServerError)
		return
	}

	formComment := &struct {
		Comment string `json:"comment"`
	}{}

	err = json.Unmarshal(data, formComment)

	if err != nil {
		h.logger.Infof("can't unmarshall json data to comment: %w", err)
		http.Error(w, "can't unmarshall json", http.StatusInternalServerError)
	}

	addCommentDto := &service.AddComment{
		User:   user,
		Body:   formComment.Comment,
		PostID: pi,
	}

	err = h.commentService.Add(addCommentDto)

	if err != nil {
		h.logger.Infof("can't add comment: %w", err)
		http.Error(w, "can't add comment", http.StatusInternalServerError)
		return
	}

	getOnePostDto := &service.GetOnePost{PostID: pi}
	postWithComments, err := h.postService.GetOne(getOnePostDto)

	if err != nil {
		h.logger.Infof("can't get post: %w", err)
		http.Error(w, "can't get post", http.StatusInternalServerError)
		return
	}

	postJSON, err := jsontransfer.GetJSON(postWithComments)

	if err != nil {
		h.logger.Infof("can't marshall post to json: %w", err)
		http.Error(w, "can't marshall post to json", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-type", contentType)
	_, err = w.Write(postJSON)

	if err != nil {
		h.logger.Infof("server can't write: %w", err)
		http.Error(w, "server can't write", http.StatusInternalServerError)
		return
	}
}

func (h postHandler) PostVote(w http.ResponseWriter, r *http.Request) {
	var inc int8
	url := r.URL.String()
	vars := mux.Vars(r)

	pi, in := vars["POST_ID"]

	if !in {
		h.logger.Info("post not found")
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}

	ctx := r.Context()
	sessCtx := ctx.Value(domain.SessionContextKey("session"))
	if sessCtx == nil {
		h.logger.Info("can't find session")
		http.Error(w, "can't find session", http.StatusNotFound)
		return
	}

	user, err := getUserFromSession(sessCtx)

	if err != nil {
		h.logger.Infof("can't convert to session: %w", err)
		http.Error(w, "can't convert to session", http.StatusInternalServerError)
		return
	}

	switch {
	case strings.Contains(url, "upvote"):
		inc = 1
	case strings.Contains(url, "downvote"):
		inc = -1
	}

	updatePostDto := &service.UpdateMetricsPost{
		PostID:   pi,
		Inc:      inc,
		AuthorID: user.ID,
	}
	err = h.postService.UpdateMetrics(updatePostDto)

	if err != nil {
		h.logger.Infof("can't inc vote: %w", err)
		http.Error(w, "can't inc vote", http.StatusInternalServerError)
		return
	}

	getOnePostDto := &service.GetOnePost{PostID: pi}
	post, err := h.postService.GetOne(getOnePostDto)

	if err != nil {
		h.logger.Infof("can't get post: %w", err)
		http.Error(w, "can't get post", http.StatusInternalServerError)
		return
	}

	postJSON, err := jsontransfer.GetJSON(post)

	if err != nil {
		h.logger.Infof("can't marshall post to json: %w", err)
		http.Error(w, "can't marshall post to json", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-type", contentType)
	_, err = w.Write(postJSON)

	if err != nil {
		h.logger.Infof("server can't write: %w", err)
		http.Error(w, "server can't write", http.StatusInternalServerError)
		return
	}
}

func (h postHandler) PostDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	pi, in := vars["POST_ID"]

	if !in {
		h.logger.Info("post not found")
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}

	getOnePostDto := &service.GetOnePost{PostID: pi}
	post, err := h.postService.GetOne(getOnePostDto)

	if err != nil {
		h.logger.Infof("can't get post: %w", err)
		http.Error(w, "can't get post", http.StatusInternalServerError)
		return
	}

	incViewsDto := &service.IncrViewsPost{PostID: pi}
	err = h.postService.IncrViews(incViewsDto)

	if err != nil {
		h.logger.Infof("can't inc views: %w", err)
		http.Error(w, "can't inc views", http.StatusInternalServerError)
		return
	}

	postJSON, err := jsontransfer.GetJSON(post)

	if err != nil {
		h.logger.Infof("can't marshall post to json: %w", err)
		http.Error(w, "can't marshall post to json", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-type", contentType)
	_, err = w.Write(postJSON)

	if err != nil {
		h.logger.Infof("server can't write: %w", err)
		http.Error(w, "server can't write", http.StatusInternalServerError)
		return
	}
}

func (h postHandler) ShowPostsByFilter(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	cn, in := vars["CATEGORY_NAME"]

	if !in {
		h.logger.Info("category not found")
		http.Error(w, "category not found", http.StatusNotFound)
		return
	}

	getByDto := &service.GetByPost{Category: "category", Data: cn, SortField: "score"}
	posts, err := h.postService.GetBy(getByDto)

	if err != nil {
		h.logger.Infof("can't get posts: %w", err)
		http.Error(w, "can't get posts", http.StatusInternalServerError)
		return
	}

	postsJSON, err := jsontransfer.GetJSON(posts)

	if err != nil {
		h.logger.Infof("can't marshall post to json: %w", err)
		http.Error(w, "can't marshall post to json", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-type", contentType)
	_, err = w.Write(postsJSON)

	if err != nil {
		h.logger.Infof("server can't write: %w", err)
		http.Error(w, "server can't write", http.StatusInternalServerError)
		return
	}
}

func (h postHandler) ShowPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.postService.Get()

	if err != nil {
		h.logger.Infof("can't get posts: %w", err)
		http.Error(w, "can't get posts", http.StatusInternalServerError)
		return
	}

	postsJSON, err := jsontransfer.GetJSON(posts)

	if err != nil {
		h.logger.Infof("can't marshall post to json: %w", err)
		http.Error(w, "can't marshall post to json", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-type", contentType)
	_, err = w.Write(postsJSON)

	if err != nil {
		h.logger.Infof("server can't write: %w", err)
		http.Error(w, "server can't write", http.StatusInternalServerError)
		return
	}
}

func (h postHandler) GetUserPosts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	ul, in := vars["USER_LOGIN"]

	if !in {
		h.logger.Info("user_login not found")
		http.Error(w, "user_login not found", http.StatusNotFound)
		return
	}

	getByDto := &service.GetByPost{Category: "author.username", Data: ul, SortField: "created"}
	posts, err := h.postService.GetBy(getByDto)

	if err != nil {
		h.logger.Infof("can't get user's posts: %w", err)
		http.Error(w, "can't get user's posts", http.StatusInternalServerError)
		return
	}

	postsJSON, err := jsontransfer.GetJSON(posts)

	if err != nil {
		h.logger.Infof("can't marshall post to json: %w", err)
		http.Error(w, "can't marshall posts to json", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-type", "application/json")
	_, err = w.Write(postsJSON)

	if err != nil {
		h.logger.Infof("server can't write: %w", err)
		http.Error(w, "server can't write", http.StatusInternalServerError)
		return
	}
}

func (h postHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	pi, in := vars["POST_ID"]

	if !in {
		h.logger.Info("post not found")
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}

	deletePostDto := &service.DeletePost{PostID: pi}
	err := h.postService.Delete(deletePostDto)

	if err != nil {
		h.logger.Infof("can't delete post: %w", err)
		http.Error(w, "can't delete post", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-type", contentType)
	_, err = w.Write([]byte("{\"message\":\"success\"}"))

	if err != nil {
		h.logger.Infof("server can't write: %w", err)
		http.Error(w, "server can't write", http.StatusInternalServerError)
		return
	}
}

func (h postHandler) AddPost(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-type") != contentType {
		h.logger.Infof("not found application/json header")
		http.Error(w, "not found application/json header", http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		h.logger.Infof("can't read form: %w", err)
		http.Error(w, "can't read form", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	sessCtx := ctx.Value(domain.SessionContextKey("session"))
	if sessCtx == nil {
		h.logger.Info("can't find session")
		http.Error(w, "can't find session", http.StatusNotFound)
		return
	}

	user, err := getUserFromSession(sessCtx)

	if err != nil {
		h.logger.Infof("can't convert to session: %w", err)
		http.Error(w, "can't convert to session", http.StatusInternalServerError)
		return
	}

	post := &domain.Post{}
	err = json.Unmarshal(data, post)

	if err != nil {
		h.logger.Infof("can't unmarshall json data to post: %w", err)
		http.Error(w, "can't unmarshall json", http.StatusInternalServerError)
		return
	}

	post.Author = &domain.Profile{Username: user.Username, ID: user.ID}
	post.Score = 1
	post.Views = 1
	post.UpvotePercentage = 100
	post.Comments = make([]*domain.Comment, 0)
	post.Votes = make([]*domain.Vote, 1)
	post.Votes[0] = &domain.Vote{
		User: user.ID,
		Vote: 1,
	}

	post, err = h.postService.Save(post)

	if err != nil {
		h.logger.Infof("can't save post in repo: %w", err)
		http.Error(w, "can't save post in repo", http.StatusInternalServerError)
		return
	}

	postJSON, err := jsontransfer.GetJSON(post)

	if err != nil {
		h.logger.Infof("can't marshall post to json: %w", err)
		http.Error(w, "can't marshall post to json", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-type", contentType)
	_, err = w.Write(postJSON)

	if err != nil {
		h.logger.Infof("server can't write: %w", err)
		http.Error(w, "server can't write", http.StatusInternalServerError)
		return
	}
}

func getUserFromSession(v any) (*domain.User, error) {
	sess, ok := v.(*domain.Session)

	if !ok {
		return nil, errors.New("can't convert to *Session{}")
	}

	if sess.User == nil {
		return nil, errors.New("empty user")
	}

	return sess.User, nil
}
