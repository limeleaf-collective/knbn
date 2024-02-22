package main

import (
	"context"
	"github.com/a-h/templ"
	"github.com/limeleaf-coop/knbn/pkg/store"
	"github.com/limeleaf-coop/knbn/templs"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
	"os"
)

var (
	defaultRedisAddr = "0.0.0.0:6379"
)

func main() {
	dbAddr := os.Getenv("DB_ADDR")
	if dbAddr == "" {
		dbAddr = defaultRedisAddr
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: dbAddr,
	})

	s := store.NewRedis(rdb)
	boards, err := s.GetBoards(context.TODO())
	if err != nil {
		slog.Error("fetching boards", "err", err)
		os.Exit(1)
	}

	//http.Handle("/boards/{id}", templ.Handler(templs.BoardPage(board)))
	http.Handle("/boards", templ.Handler(templs.BoardsPage(boards)))
	http.Handle("/", templ.Handler(templs.IndexPage()))
	http.ListenAndServe(":8080", nil)
}
