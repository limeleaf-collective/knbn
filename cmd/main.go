package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/limeleaf-coop/knbn/pkg"
	docdb "github.com/limeleaf-coop/knbn/pkg/db"
)

func main() {
	ctx := context.Background()

	address := flag.String("address", ":8080", "addr to bind the HTTP server to")
	database := flag.String("database", "./knbn.sqlite", "database file location")
	seedDataDir := flag.String("seed-data-dir", "", "directory containing .json file of seed data")
	flag.Parse()

	db, err := docdb.Open(*database)
	if err != nil {
		slog.Error("error opening database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	slog.Info("opened database", "database", *database)

	if *seedDataDir != "" {
		if err := db.SeedFromDir(ctx, *seedDataDir); err != nil {
			slog.Error("error seeding database", "error", err)
		}
		slog.Info("seeded database", "dir", *seedDataDir)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /boards/{boardId}/lists/{listIdx}/cards/{cardIdx}/title", pkg.TitleHandler(db))
	mux.HandleFunc("GET /boards/{boardId}/lists/{listIdx}/cards/{cardIdx}/title/edit", pkg.EditTitleHandler(db))
	mux.HandleFunc("GET /boards/{boardId}/lists/{listIdx}/title", pkg.TitleHandler(db))
	mux.HandleFunc("GET /boards/{boardId}/lists/{listIdx}/title/edit", pkg.EditTitleHandler(db))
	mux.HandleFunc("GET /boards/{id}", pkg.BoardHandler(db))
	mux.HandleFunc("GET /boards", pkg.BoardsHandler(db))
	mux.HandleFunc("POST /sign-in", pkg.SignInHandler(db))
	mux.HandleFunc("GET /", pkg.IndexHandler)

	srv := &http.Server{
		Addr:    *address,
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("error starting HTTP server", "error", err)
			os.Exit(1)
		}
	}()
	slog.Info("starting HTTP server", "address", *address)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	sig := <-done
	slog.Info("shutdown signal received", "signal", sig)

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("error gracefully shutting down", "error", err)
	}
	slog.Info("shutdown completed")
}
