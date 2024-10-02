package receive_lyrics

import (
	"database/sql"
	"effective-mobile/internal/storage/postgres"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

type SongLyricsResponse struct {
	CurrentPage int      `json:"current_page"`
	TotalPages  int      `json:"total_pages"`
	Lyrics      []string `json:"lyrics"`
}

// @BasePath /song/lyrics

// New creates a handler to get the lyrics of a song broken down by pages
// @Summary Get the lyrics of the song
// @Description Returns the lyrics of the song, divided into pages.
// @Tags lyrics
// @Produce json
// @Param group query string true "group"
// @Param song query string true "song"
// @Param page query int false "Page number (default is 1)"
// @Param limit query int false "Number of verses per page (2 by default))"
// @Success 200 {object} SongLyricsResponse "Lyrics by page"
// @Failure 400 {string} string "Invalid request parameters"
// @Failure 500 {string} string "Server error"
// @Router /song/lyrics [get]
func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.receive-lyrics.New"
		log = log.With(
			slog.String("op", op),
		)

		log.Debug("Received a request", slog.Any("request", r))

		group := r.URL.Query().Get("group")
		song := r.URL.Query().Get("song")
		pageStr := r.URL.Query().Get("page")
		limitStr := r.URL.Query().Get("limit")
		log.Info("Incoming request parameters",
			slog.String("group", group),
			slog.String("song", song),
			slog.String("pageStr", pageStr),
			slog.String("limitStr", limitStr),
		)
		page := 1
		limit := 2

		if pageStr != "" {
			var err error
			page, err = strconv.Atoi(pageStr)
			if err != nil {
				http.Error(w, "Invalid page parameter", http.StatusBadRequest)
				log.Error("Invalid page parameter", slog.Any("error", err))
				return
			}
			log.Debug("Parsed page parameter", slog.Int("page", page))
		}

		if limitStr != "" {
			var err error
			limit, err = strconv.Atoi(limitStr)
			if err != nil {
				http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
				log.Error("Invalid limit parameter", slog.Any("error", err))
				return
			}
			log.Debug("Parsed limit parameter", slog.Int("limit", limit))
		}
		lyrics, err := storage.GetLyrics(song, group)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Failed to get song lyrics", http.StatusBadRequest)
				log.Error("Bad request", slog.Any("error", err))
				return
			} else {
				http.Error(w, "Failed to get song lyrics", http.StatusInternalServerError)
				log.Error("Failed to select lyrics", slog.Any("error", err))
				return
			}

		}
		log.Info("Lyrics retrieved from storage", slog.String("song", song), slog.String("group", group))

		// Dividing the lyrics into verses
		verses := strings.Split(lyrics, "\n\n")

		// Total page count
		totalPages := (len(verses) + limit - 1) / limit
		log.Debug("Total pages calculated", slog.Int("totalPages", totalPages))

		// The restriction for the page is not to go beyond
		if page > totalPages {
			page = totalPages
			log.Debug("Adjusted page to total pages", slog.Int("adjustedPage", page))

		} else if page < 1 {
			page = 1
			log.Debug("Adjusted page to minimum", slog.Int("adjustedPage", page))
		}

		// Calculating the beginning and end of the verses for the current page
		start := (page - 1) * limit
		end := start + limit
		if end > len(verses) {
			end = len(verses)
		}

		response := SongLyricsResponse{
			CurrentPage: page,
			TotalPages:  totalPages,
			Lyrics:      verses[start:end],
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			log.Error("Failed to encode JSON response", slog.Any("error", err))
			return
		}

		log.Info(
			"Lyrics successfully sent to client",
			slog.Int("current_page", page),
			slog.Int("total_pages", totalPages),
		)
	}
}
