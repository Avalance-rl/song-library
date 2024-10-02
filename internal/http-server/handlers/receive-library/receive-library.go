package receive_library

import (
	"effective-mobile/internal/storage/postgres"
	"effective-mobile/internal/storage/postgres/queries"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	_ "github.com/lib/pq"
)

// @BasePath /song/library

// New godoc
// @Summary Retrieve the user's song library
// @Description Retrieves the user's entire song library, optionally filtered by group, song, or release date.
// @Tags songs
// @Produce  json
// @Param group query string false "Group name"
// @Param song query string false "Song name"
// @Param releaseDate query string false "Release date (YYYY-MM-DD)"
// @Success 200 "List of songs"
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal server error"
// @Router /song/library [get]
func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.receive-library.New"
		log = log.With(
			slog.String("op", op),
		)

		log.Debug("Received a request", slog.Any("request", r))

		group := r.URL.Query().Get("group")
		song := r.URL.Query().Get("song")
		releaseDate := r.URL.Query().Get("releaseDate")
		log.Info("Incoming request parameters",
			slog.String("group", group),
			slog.String("song", song),
			slog.String("releaseDate", releaseDate),
		)
		query := queries.GetLibrary

		var filters []interface{}
		count := 1
		if group != "" {
			query += fmt.Sprintf(" AND group_name = $%v", count)
			filters = append(filters, group)
			count++
		}
		if song != "" {
			query += fmt.Sprintf(" AND song_name = $%v", count)
			filters = append(filters, song)
			count++
		}
		if releaseDate != "" {
			query += fmt.Sprintf(" AND release_date = $%v", count)
			filters = append(filters, releaseDate)
		}

		log.Info("Executing SQL query", slog.String("query", query), slog.Any("filters", filters))

		res, err := storage.SelectSongs(query, filters)
		if err != nil {
			http.Error(w, "Failed to get a successful response", http.StatusInternalServerError)
			log.Error("Failed to select", slog.Any("statusCode", err))
			return
		}

		log.Info("Songs retrieved", slog.Int("count", len(res)))
		log.Debug("Retrieved songs data", slog.Any("songs", res))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(res); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			log.Error("Failed to encode", slog.Any("statusCode", err))
		}
		log.Info("Request successfully processed", slog.Int("status", http.StatusOK))
	}
}
