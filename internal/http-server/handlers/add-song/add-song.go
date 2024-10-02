package add_song

import (
	"effective-mobile/internal/storage/postgres"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Song struct {
	Group string `json:"group"`
	Song  string `json:"song"`
}

type SongDetail struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

// @BasePath /song/add

// New creates a handler for adding a new song
// @Summary Add a new song
// @Description Adds a song with information about the band, name, release date, lyrics and a link to YouTube
// @Tags song
// @Accept json
// @Produce json
// @Param song body Song true "Information about the song"
// @Success 201 {string} string "Song added successfully"
// @Failure 400 {string} string "Invalid JSON format"
// @Failure 415 {string} string "Content-Type header is not application/json"
// @Failure 500 {string} string "Internal server error"
// @Router /songs [post]
func New(log *slog.Logger, storage *postgres.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.add-song.New"
		log = log.With(
			slog.String("op", op),
		)

		log.Debug("Processing request", slog.Any("request", r))

		ct := r.Header.Get("Content-Type")
		if ct != "" {
			mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
			if mediaType != "application/json" {
				msg := "Content-Type header is not application/json"
				http.Error(w, msg, http.StatusUnsupportedMediaType)
				log.Info("%s", msg)
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

		log.Info("Received song data", slog.Any("song", song))
		log.Debug("Decoded song data", slog.Any("decodedSong", song))

		url := fmt.Sprintf("/info?group=%s&song=%s", song.Group, song.Song)
		log.Info("Fetching song details from API", slog.String("url", url))
		resp, err := http.Get(url)
		if err != nil {
			http.Error(w, "Error internal server", http.StatusInternalServerError)
			log.Error("Failed to connect towards api", slog.Any("error", err))
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			if resp.StatusCode == http.StatusBadRequest {
				http.Error(w, "Failed to get a successful response", http.StatusBadRequest)
				log.Error("Bad request", slog.Any("statusCode", resp.StatusCode))
				return
			} else {
				http.Error(w, "Failed to get a successful response", http.StatusInternalServerError)
				log.Error("Internal server error", slog.Any("statusCode", resp.StatusCode))
				return
			}
		}
		var songDetail SongDetail
		err = json.NewDecoder(resp.Body).Decode(&songDetail)
		if err != nil {
			http.Error(w, "Failed to decode API response", http.StatusInternalServerError)
			log.Error("Failed to decode API response", slog.Any("error", err))
			return
		}
		log.Info("Received song details", slog.Any("songDetail", songDetail))
		log.Debug("Decoded song details", slog.Any("decodedSongDetail", songDetail))

		t, err := time.Parse("02.01.2006", songDetail.ReleaseDate)
		if err != nil {
			log.Error("Invalid date format from API", slog.Any("error", err))
			http.Error(w, "Invalid date format from API", http.StatusInternalServerError)
			return
		}
		formattedDate := t.Format("2006-01-02")

		log.Debug("Formatted release date", slog.String("formattedDate", formattedDate))

		err = storage.InsertSong(postgres.Song{
			GroupName:   song.Group,
			SongName:    song.Song,
			ReleaseDate: formattedDate,
			Lyrics:      songDetail.Text,
			YoutubeLink: songDetail.Link,
		})
		if err != nil {
			http.Error(w, "Error internal server", http.StatusInternalServerError)
			log.Error("Failed to insert song at storage", slog.Any("error", err))
			return
		}

		log.Info("Song successfully added", slog.String("song", song.Song), slog.String("group", song.Group))
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Song added successfully"))
	}
}
