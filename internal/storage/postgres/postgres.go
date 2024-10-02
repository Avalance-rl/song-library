package postgres

import (
	"context"
	"database/sql"
	"effective-mobile/internal/storage/postgres/queries"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sqlx.DB
}

var ErrSongNotFound = errors.New("song not found")

type Song struct {
	ID          uint   `db:"id"`
	GroupName   string `db:"group_name"`
	SongName    string `db:"song_name"`
	ReleaseDate string `db:"release_date"`
	Lyrics      string `db:"lyrics"`
	YoutubeLink string `db:"youtube_link"`
}
type Lyrics struct {
	Lyrics string `db:"lyrics"`
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.postgres.New"
	slog.Log(context.TODO(), slog.LevelInfo, op)
	db := sqlx.MustConnect("postgres", storagePath)
	return &Storage{db: db}, nil
}

func (s *Storage) Stop() error {
	const op = "storage.postgres.Stop"
	slog.Log(context.TODO(), slog.LevelInfo, op)
	return s.db.Close()
}

func (s *Storage) InsertSong(song Song) error {
	const op = "storage.postgres.InsertSong"
	slog.Log(context.TODO(), slog.LevelInfo, op)
	tx := s.db.MustBegin()
	var args []interface{}
	args = append(args, song.GroupName, song.SongName, song.ReleaseDate, song.Lyrics, song.YoutubeLink)
	tx.MustExec(
		queries.InsertSong, args...,
	)
	err := tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) SelectSongs(query string, args []interface{}) ([]Song, error) {
	const op = "storage.postgres.SelectSongs"
	slog.Log(context.TODO(), slog.LevelInfo, op)
	var songs []Song
	query += " LIMIT 5"
	err := s.db.Select(&songs, query, args...)
	if err != nil {
		return nil, err
	}
	return songs, nil
}

func (s *Storage) DeleteSong(song string, group string) error {
	const op = "storage.postgres.DeleteSong"
	slog.Log(context.TODO(), slog.LevelInfo, op)
	res, err := s.db.Exec(queries.DeleteSong, group, song)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to retrieve rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrSongNotFound
	}

	return nil
}

func (s *Storage) GetLyrics(song string, group string) (string, error) {
	const op = "storage.postgres.GetLyrics"
	slog.Log(context.TODO(), slog.LevelInfo, op)
	var lyrics Lyrics
	err := s.db.Get(&lyrics, queries.GetLyrics, song, group)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
		return "", err
	}
	return lyrics.Lyrics, nil
}

func (s *Storage) UpdateSong(firstSong, firstGroup, song string, group string, releaseDate string) error {
	const op = "storage.postgres.UpdateSong"
	slog.Log(context.TODO(), slog.LevelInfo, op)
	query := queries.UpdateSong
	var setClauses []string
	var params []interface{}
	count := 1

	if group != "" {
		setClauses = append(setClauses, fmt.Sprintf("group_name = $%d", count))
		params = append(params, group)
		count++
	}
	if song != "" {
		setClauses = append(setClauses, fmt.Sprintf("song_name = $%d", count))
		params = append(params, song)
		count++
	}
	if releaseDate != "" {
		t, err := time.Parse("02.01.2006", releaseDate)
		if err != nil {
			fmt.Errorf("%s: %v", op, err)
		}
		formattedDate := t.Format("2006-01-02")
		setClauses = append(setClauses, fmt.Sprintf("release_date = $%d", count))
		params = append(params, formattedDate)
		count++
	}

	if len(setClauses) == 0 {
		return fmt.Errorf("%s: no fields to update", op)
	}

	query += strings.Join(setClauses, ", ")
	query += " WHERE song_name = $" + strconv.Itoa(count) + " AND group_name = $" + strconv.Itoa(count+1)
	params = append(params, firstSong, firstGroup)
	_, err := s.db.Exec(query, params...)
	if err != nil {
		return err
	}
	return nil
}
