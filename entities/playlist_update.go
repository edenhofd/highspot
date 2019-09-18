package entities

type PlaylistUpdate struct {
	PlaylistID string `json:"playlist_id"`
	NewSongIDs []string `json:"new_song_ids"`
}