package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/akrovv/redditclone/internal/adapters/mongodb"
	"github.com/akrovv/redditclone/internal/adapters/pgsqldb"
	"github.com/akrovv/redditclone/internal/adapters/redisdb"
	"github.com/akrovv/redditclone/internal/config"
	"github.com/akrovv/redditclone/internal/controllers/rest"
	"github.com/akrovv/redditclone/internal/controllers/rest/middleware"
	"github.com/akrovv/redditclone/internal/service"
	"github.com/akrovv/redditclone/pkg/logger"
	"github.com/casbin/casbin"
	"github.com/redis/go-redis/v9"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	path     = "."
	filename = ".env"
)

func main() {
	cfg, err := config.New(filename, path)

	if err != nil {
		log.Fatal(err)
		return
	}

	ctxRedis := context.Background()
	dsnRedis := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)

	client := redis.NewClient(&redis.Options{
		Addr: dsnRedis,
		DB:   0,
	})

	err = client.Ping(ctxRedis).Err()

	if err != nil {
		log.Fatal("redis", err)
		return
	}
	defer client.Close()

	ctxMongo := context.Background()
	dsnMongo := fmt.Sprintf("mongodb://%s", cfg.MongoHost)
	mongoClient, err := mongo.Connect(ctxMongo, options.Client().ApplyURI(dsnMongo))

	if err != nil {
		log.Fatal("mongo", err)
		return
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.SSLMode)
	db, err := sql.Open("postgres", dsn)

	if err != nil {
		log.Fatal(err)
		return
	}

	db.SetMaxOpenConns(10)

	err = db.Ping()

	if err != nil {
		log.Fatal(err)
		return
	}

	e, err := casbin.NewEnforcerSafe("basic_model.conf", "basic_policy.csv")
	if err != nil {
		log.Fatal(err)
		return
	}

	l := logger.NewLogger()

	var (
		commentDB = mongodb.NewCommentStorage(ctxMongo, mongoClient)
		postDB    = mongodb.NewPostStorage(ctxMongo, mongoClient)
		userDB    = pgsqldb.NewUserStorage(db)
		sessionDB = redisdb.NewSessionStorage(ctxRedis, client)
	)

	var (
		commentService = service.NewCommentService(commentDB)
		postService    = service.NewPostService(postDB)
		userService    = service.NewUserService(userDB)
		sessionService = service.NewSessionService(sessionDB)
	)

	rootHandler := rest.NewRootHandler(l)
	userHandler := rest.NewUserHandler(l, userService, sessionService)
	postHandler := rest.NewPostHandler(l, postService, commentService, sessionService)

	router := mux.NewRouter()

	// Static
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("front")))
	router.PathPrefix("/static").Handler(staticHandler)
	router.Handle("/static/", staticHandler)

	// Main
	router.HandleFunc("/", rootHandler.Main)

	// API

	// User
	router.HandleFunc("/signup", rootHandler.Main)
	router.HandleFunc("/login", rootHandler.Main)
	router.HandleFunc("/u/{USER_LOGIN}", rootHandler.Main)

	router.HandleFunc("/api/register", userHandler.Register).Methods("POST")
	router.HandleFunc("/api/login", userHandler.Login).Methods("POST")

	// Posts
	router.HandleFunc("/createpost", rootHandler.Main)
	router.HandleFunc("/a/{CATEGORY_NAME}", rootHandler.Main)
	router.HandleFunc("/a/{CATEGORY_NAME}/{POST_ID}", rootHandler.Main)

	router.HandleFunc("/api/posts/", postHandler.ShowPosts).Methods("GET")
	router.HandleFunc("/api/posts", postHandler.AddPost).Methods("POST")
	router.HandleFunc("/api/posts/{CATEGORY_NAME}", postHandler.ShowPostsByFilter).Methods("GET")
	router.HandleFunc("/api/post/{POST_ID}", postHandler.PostDetail).Methods("GET")
	router.HandleFunc("/api/post/{POST_ID}", postHandler.AddComment).Methods("POST")
	router.HandleFunc("/api/post/{POST_ID}/{COMMENT_ID}", postHandler.DeleteComment).Methods("DELETE")
	router.HandleFunc("/api/post/{POST_ID}/upvote", postHandler.PostVote).Methods("GET")
	router.HandleFunc("/api/post/{POST_ID}/downvote", postHandler.PostVote).Methods("GET")
	router.HandleFunc("/api/post/{POST_ID}/unvote", postHandler.PostVote).Methods("GET")
	router.HandleFunc("/api/post/{POST_ID}", postHandler.DeletePost).Methods("DELETE")
	router.HandleFunc("/api/user/{USER_LOGIN}", postHandler.GetUserPosts).Methods("GET")

	siteMux := middleware.Logger(router, l)
	siteMux = middleware.Permissions(siteMux, l, e)
	siteMux = middleware.Auth(siteMux, sessionService)

	log.Println("Server is starting on :8080")
	errListen := http.ListenAndServe(":8080", siteMux)

	if errListen != nil {
		return
	}
}
