package queries

const InsertSong = "INSERT INTO songs (group_name, song_name, release_date, lyrics, youtube_link) VALUES ($1, $2, $3, $4, $5)"
const GetLibrary = "SELECT * FROM songs WHERE 1=1"
const GetLyrics = "SELECT lyrics FROM songs WHERE song_name = $1 AND group_name = $2"
const DeleteSong = "DELETE FROM songs WHERE group_name = $1 AND song_name = $2"
const UpdateSong = "UPDATE songs SET "
