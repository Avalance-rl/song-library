package update_song_data

import (
	"effective-mobile/internal/storage/postgres"
	"encoding/json"
	"log/slog"
	"net/http"

	_ "github.com/lib/pq"
)

type UpdateSongRequest struct {
	FirstSong   string `json:"firstSong"`
	FirstGroup  string `json:"firstGroup"`
	Group       string `json:"group,omitempty"`
	Song        string `json:"song,omitempty"`
	ReleaseDate string `json:"release_date,omitempty"`
}

// New creates a handler for updating song data
// @Summary Update the song data
// @Description Updates the song data in the repository based on the original song data and new data.
// @Tags songs
// @Accept json
// @Produce json
// @Param updateRequest body UpdateSongRequest true "Data for updating the song"
// @Success 204 {string} string "The song data has been successfully updated"
// @Failure 400 {string} string "Invalid request parameters"
// @Failure 500 {string} string "Server error"
// @Router /song/update [patch]
func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.update-song.New"
		log = log.With(
			slog.String("op", op),
		)

		log.Debug("Received a request", slog.Any("request", r))

		ct := r.Header.Get("Content-Type")
		if ct != "application/json" {
			msg := "Content-Type header is not application/json"
			http.Error(w, msg, http.StatusUnsupportedMediaType)
			log.Info("Unsupported media type", slog.String("content-type", ct))
			return
		}

		var updateRequest UpdateSongRequest
		err := json.NewDecoder(r.Body).Decode(&updateRequest)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			log.Error("Failed to decode JSON", slog.Any("error", err))
			return
		}
		if updateRequest.FirstSong == "" || updateRequest.FirstGroup == "" {
			http.Error(w, "FirstSong and FirstGroup fields are required", http.StatusBadRequest)
			log.Error("Missing required fields", slog.Any("song", updateRequest))
			return
		}
		log.Debug("Attempting to update song data", slog.Any("updateRequest", updateRequest))

		err = storage.UpdateSong(
			updateRequest.FirstSong,
			updateRequest.FirstGroup,
			updateRequest.Song,
			updateRequest.Group,
			updateRequest.ReleaseDate,
		)
		if err != nil {
			http.Error(w, "Failed to update song", http.StatusInternalServerError)
			log.Error("Failed to update song", slog.Any("statusCode", err))
			return
		}

		w.WriteHeader(http.StatusNoContent)
		log.Info("Song data updated successfully", slog.Any("updateRequest", updateRequest))
	}
}
