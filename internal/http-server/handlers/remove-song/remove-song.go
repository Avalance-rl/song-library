package remove_song

import (
	"effective-mobile/internal/storage/postgres"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
)

type Song struct {
	Group string `json:"group"`
	Song  string `json:"song"`
}

// New creates a handler for deleting a song
// @Summary Delete a song
// @Description Deletes a song from the repository by the name of the band and the name of the song.
// @Tags song
// @Accept json
// @Produce json
// @Param song body Song true "Data for deleting a song"
// @Success 200 {string} string "The song was successfully deleted"
// @Failure 400 {string} string "Invalid request parameters"
// @Failure 404 {string} string "The song was not found"
// @Failure 500 {string} string "Server error"
// @Router /song/remove [delete]
func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.remove-song.New"
		log = log.With(
			slog.String("op", op),
		)

		log.Debug("Received a request", slog.Any("request", r))

		ct := r.Header.Get("Content-Type")
		if ct != "" {
			mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
			if mediaType != "application/json" {
				msg := "Content-Type header is not application/json"
				http.Error(w, msg, http.StatusUnsupportedMediaType)
				log.Warn("Invalid Content-Type", slog.String("content-type", ct))
				return
			}
		}
		var song Song
		err := json.NewDecoder(r.Body).Decode(&song)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			log.Error("Failed to decode JSON", slog.Any("error", err))
			return
		}

		if song.Group == "" || song.Song == "" {
			http.Error(w, "Group and Song fields are required", http.StatusBadRequest)
			log.Error("Missing required fields", slog.Any("song", song))
			return
		}

		log.Info("Received song data", slog.Any("song", song))

		log.Debug("Attempting to delete song", slog.String("group", song.Group), slog.String("song", song.Song))

		err = storage.DeleteSong(song.Song, song.Group)
		if err != nil {
			if errors.Is(err, postgres.ErrSongNotFound) {
				http.Error(w, "Song not found", http.StatusNotFound)
				log.Warn("Song not found", slog.String("group", song.Group), slog.String("song", song.Song))
			} else {
				http.Error(w, "Failed to delete song", http.StatusInternalServerError)
				log.Error("Failed to delete song", slog.Any("error", err))
			}
			return
		}

		log.Info("Song deleted successfully", slog.String("group", song.Group), slog.String("song", song.Song))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Song deleted successfully"))
	}
}
