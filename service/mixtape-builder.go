package service

import (
	"mixtape/entities"
	"os"
	"io/ioutil"
	"encoding/json"
	"fmt"
)

type MixtapeBuilder struct {
	// service objects obtained from inputs
	Users map[string]entities.User `json:"users"`
	Playlists map[string]entities.Playlist `json:"playlists"`
	Songs map[string]entities.Song `json:"songs"`
}

type InputOutputFile struct {
	Users []entities.User `json:"users"`
	Playlists []entities.Playlist `json:"playlists"`
	Songs []entities.Song `json:"songs"`
}

type ChangesFile struct {
	NewPlaylists []entities.Playlist `json:"new_playlists"`
	RemovePlaylists []string `json:"remove_playlists"`
	UpdatePlaylists []entities.PlaylistUpdate `json:"update_playlists"`
}

func NewMixtapeBuilder(inputPath string) (*MixtapeBuilder, error) {
	// attempt to open up the input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		// error opening file. pass it up
		return nil, err
	}
	defer inputFile.Close()

	// read file bytes
	inputBytes, err := ioutil.ReadAll(inputFile)
	if err != nil {
		// error reading file. pass it up
		return nil, err
	}

	// convert bytes to json
	var inputJson InputOutputFile
	err = json.Unmarshal(inputBytes, &inputJson)
	if err != nil {
		// error unmarshaling file. pass it up
		return nil, err
	}

	// build our builder struct
	mixtapeBuilder := MixtapeBuilder{
		Users: make(map[string]entities.User, 0),
		Songs: make(map[string]entities.Song, 0),
		Playlists: make(map[string]entities.Playlist, 0),
	}

	for _, user := range inputJson.Users {
		mixtapeBuilder.Users[user.ID] = user
	}
	for _, song := range inputJson.Songs {
		mixtapeBuilder.Songs[song.ID] = song
	}
	for _, playlist := range inputJson.Playlists {
		mixtapeBuilder.Playlists[playlist.ID] = playlist
	}

	return &mixtapeBuilder, nil
}

func (m *MixtapeBuilder) ApplyUpdates(changesPath string) error {
	// attempt to open up the changes file
	changesFile, err := os.Open(changesPath)
	if err != nil {
		// error opening file. pass it up
		return err
	}
	defer changesFile.Close()

	// read file bytes
	changesBytes, err := ioutil.ReadAll(changesFile)
	if err != nil {
		// error reading file. pass it up
		return err
	}

	// convert bytes to json
	var changesJson ChangesFile
	err = json.Unmarshal(changesBytes, &changesJson)
	if err != nil {
		// error unmarshaling file. pass it up
		return err
	}

	// run changes
	// remove playlists
	for _, rmID := range changesJson.RemovePlaylists {
		// chack for existance of playlist to remove
		if _, ok := m.Playlists[rmID]; !ok {
			// if not found, print out but just continue
			fmt.Printf("Could not find Playlist with ID: %s for removal. Skipping\n", rmID)
			continue
		}
		delete(m.Playlists, rmID)
	}

	// add playlists
	for _, playlist := range changesJson.NewPlaylists {
		// check to see if playlist already exists
		if _, ok := m.Playlists[playlist.ID]; ok {
			// if so, print but just skip onward
			fmt.Printf("Playlist with ID already exists! Cannot add Playlist: %v\n", playlist)
			continue
		}
		// check to ensure user exists
		if _, ok := m.Users[playlist.UserID]; !ok {
			// if not, print but continue onward
			fmt.Printf("No User found with ID: %s. Cannot add Playlist: %v\n", playlist.UserID, playlist)
			continue
		}
		// check to ensure SOME songs exist
		foundSongs := make([]string, 0)
		for _, songID := range playlist.SongIDs {
			if _, ok := m.Songs[songID]; ok {
				foundSongs = append(foundSongs, songID)
			}
		}
		if len(foundSongs) == 0 {
			fmt.Printf("No existing songs found. Cannot add Playlist: %v", playlist)
			continue
		}
		playlist.SongIDs = foundSongs
		m.Playlists[playlist.ID] = playlist
	}

	// playlist updates
	for _, plUpdate := range changesJson.UpdatePlaylists {
		// ensure playlist exists
		playlist, ok := m.Playlists[plUpdate.PlaylistID]
		if !ok {
			fmt.Printf("No Playlist found with ID: %s. Cannot update: %v", plUpdate.PlaylistID, plUpdate)
			continue
		}
		// ensure each song to add exists
		for _, songID := range plUpdate.NewSongIDs {
			if _, ok := m.Songs[songID]; !ok {
				fmt.Printf("No Song found with ID: %s. Cannot add to Playlist. Skipping", songID)
				continue
			}
			// optimally we would convert the songs member of the Playlist struct to a map or something for 
			// constant time checking to ensure we dont add a song that already exists. don't really want to
			// double loop for each found song and it feels a bit much to go back and restructure but be aware
			// in a perfect or even better world we would also be checking against adding an already existing song 
			// entry into the same playlist. I'm just going to blind add at this time however.
			playlist.SongIDs = append(playlist.SongIDs, songID)
		}
		m.Playlists[playlist.ID] = playlist
	}
	return nil
}

func (m *MixtapeBuilder) ExportEntities(outputPath string) error {
	// convert internal storage maps to JSON output arrays
	var output InputOutputFile
	// convert users
	for _, user := range m.Users {
		output.Users = append(output.Users, user)
	}
	// convert playlists
	for _, playlist := range m.Playlists {
		output.Playlists = append(output.Playlists, playlist)
	} 
	// convert songs
	for _, song := range m.Songs {
		output.Songs = append(output.Songs, song)
	}

	// grab json bytes
	outputBytes, err := json.Marshal(output)
	if err != nil {
		// issue marshaling json file. passing up
		return err
	}

	// create and open the output file for writing
	outputFile, err := os.Create(outputPath)
	if err != nil {
		// issue creating output file. passing up
		return err
	}
	defer outputFile.Close()

	_, err = outputFile.Write(outputBytes)
	if err != nil {
		// error writing to file. passing up
		return err
	}
	// all set. sync write to disk then return out
	outputFile.Sync()
	return nil
}