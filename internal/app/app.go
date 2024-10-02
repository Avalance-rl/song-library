package app

import (
	"context"
	_ "effective-mobile/docs"
	"effective-mobile/internal/config"
	addSong "effective-mobile/internal/http-server/handlers/add-song"
	receiveLibrary "effective-mobile/internal/http-server/handlers/receive-library"
	receiveLyrics "effective-mobile/internal/http-server/handlers/receive-lyrics"
	removeSong "effective-mobile/internal/http-server/handlers/remove-song"
	updateSongData "effective-mobile/internal/http-server/handlers/update-song-data"
	"effective-mobile/internal/services/middleware/logger"
	"effective-mobile/internal/storage/postgres"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpSwagger "github.com/swaggo/http-swagger"
)

func Run() {
	cfg := config.MustLoad()
	log := setupLogger()
	db, err := postgres.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage")
		os.Exit(1)
	}
	log.Info("starting app", slog.String("version", "1"))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /song/library", receiveLibrary.New(log, db))
	mux.HandleFunc("GET /song/lyrics", receiveLyrics.New(log, db))
	mux.HandleFunc("POST /song/add", addSong.New(log, db))
	mux.HandleFunc("PATCH /song/update", updateSongData.New(log, db))
	mux.HandleFunc("DELETE /song/remove", removeSong.New(log, db))
	mux.HandleFunc("/swagger/", httpSwagger.WrapHandler)

	loggedMux := logger.New(log)(mux)

	srv := &http.Server{
		Addr:    net.JoinHostPort(cfg.Address, cfg.Port),
		Handler: loggedMux,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {

		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}

	}()

	log.Info("server started")

	<-done
	log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", err)

		return
	}
	err = db.Stop()
	if err != nil {
		log.Error("error after db.stop", err)
	} else {
		log.Info("storage closed")
	}
	log.Info("server stopped")
}

func setupLogger() *slog.Logger {
	var log *slog.Logger
	log = slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)

	return log
}
