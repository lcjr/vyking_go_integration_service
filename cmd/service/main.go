package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"vyking-service/internal/handler"
	"vyking-service/internal/helpwise"
	"vyking-service/internal/repository"
	"vyking-service/internal/service"

	_ "github.com/go-sql-driver/mysql"
)

func main() {

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:rootpassword@tcp(127.0.0.1:3306)/vyking_db?parseTime=true"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		slog.Error("failed to connect to mysql", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	repo := repository.NewMySQLRepository(db)
	hwClient := helpwise.NewClient("https://api.helpwise.io")

	// Brand-specific API Keys mapping
	brandCredentials := map[uint32]string{
		1: "brand_1_hw_key",
		2: "brand_2_hw_key",
	}

	syncSvc := service.NewSyncService(repo, hwClient, brandCredentials)
	txSvc := service.NewTransactionService(repo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := syncSvc.ProcessQueueBatch(ctx, 10); err != nil {
					slog.Error("queue processing error", "err", err)
				}
			}
		}
	}()

	h := handler.NewHTTPHandler(txSvc, repo)
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: h,
	}

	go func() {
		slog.Info("starting partner transaction API server", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server failed to start", "err", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down server gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", "err", err)
	}
}
