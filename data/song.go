package data

import (
	"errors"
	"io"
	"os"

	"github.com/wtolson/go-taglib"
)

// Constants representing the various song file types which wavepipe can index
const (
	APE = iota
	FLAC
	M4A
	MP3
	MPC
	OGG
	WMA
	WV
)

var (
	// ErrSongTags is returned when required tags could not be extracted from a TagLib file
	ErrSongTags = errors.New("song: required tags could not be extracted from TagLib file")
	// ErrSongProperties is returned when required properties could not be extracted from a TagLib file
	ErrSongProperties = errors.New("song: required properties could not be extracted from TagLib file")
)

// FileTypeMap maps song extension to wavepipe file type IDs
var FileTypeMap = map[string]int{
	".ape":  APE,
	".flac": FLAC,
	".m4a":  M4A,
	".mp3":  MP3,
	".mpc":  MPC,
	".ogg":  OGG,
	".wma":  WMA,
	".wv":   WV,
}

// CodecMap maps wavepipe file type IDs to file types
var CodecMap = map[int]string{
	APE:  "APE",
	FLAC: "FLAC",
	M4A:  "M4A",
	MP3:  "MP3",
	MPC:  "MPC",
	OGG:  "OGG",
	WMA:  "WMA",
	WV:   "WV",
}

// MIMEMap maps a wavepipe file type ID its MIME type
// BUG(mdlayher): MIMEMap: verify correctness of MIME types
var MIMEMap = map[int]string{
	APE:  "audio/ape",
	FLAC: "audio/flac",
	M4A:  "audio/aac",
	MP3:  "audio/mpeg",
	MPC:  "audio/mpc",
	OGG:  "audio/ogg",
	WMA:  "audio/wma",
	WV:   "audio/wv",
}

// Song represents a song known to wavepipe, and contains metadata regarding
// the song, and where it resides in the filsystem
type Song struct {
	ID           int    `json:"id"`
	Album        string `json:"album"`
	AlbumID      int    `db:"album_id" json:"albumId"`
	ArtID        int    `db:"art_id" json:"artId"`
	Artist       string `json:"artist"`
	ArtistID     int    `db:"artist_id" json:"artistId"`
	Bitrate      int    `json:"bitrate"`
	Channels     int    `json:"channels"`
	Comment      string `json:"comment"`
	FileName     string `db:"file_name" json:"fileName"`
	FileSize     int64  `db:"file_size" json:"fileSize"`
	FileTypeID   int    `db:"file_type_id" json:"fileTypeId"`
	FolderID     int    `db:"folder_id" json:"folderId"`
	Genre        string `json:"genre"`
	LastModified int64  `db:"last_modified" json:"lastModified"`
	Length       int    `json:"length"`
	SampleRate   int    `db:"sample_rate" json:"sampleRate"`
	Title        string `json:"title"`
	Track        int    `json:"track"`
	Year         int    `json:"year"`
}

// SongFromFile creates a new Song from a TagLib file, extracting its tags and properties
// into the fields of the struct.
func SongFromFile(file *taglib.File) (*Song, error) {
	// Retrieve some tags needed by wavepipe, check for empty
	// At minimum, we will need an artist and title to do anything useful with this file
	title := file.Title()
	artist := file.Artist()
	if title == "" || artist == "" {
		return nil, ErrSongTags
	}

	// Retrieve all properties, check for empty
	// Note: length will probably be more useful as an integer, and a Duration method can
	// be added later on if needed
	bitrate := file.Bitrate()
	channels := file.Channels()
	length := int(file.Length().Seconds())
	sampleRate := file.Samplerate()

	if bitrate == 0 || channels == 0 || length == 0 || sampleRate == 0 {
		return nil, ErrSongProperties
	}

	// Copy over fields from TagLib tags and properties, as well as OS information
	return &Song{
		Album:      file.Album(),
		Artist:     artist,
		Bitrate:    bitrate,
		Channels:   channels,
		Comment:    file.Comment(),
		Genre:      file.Genre(),
		Length:     length,
		SampleRate: sampleRate,
		Title:      title,
		Track:      file.Track(),
		Year:       file.Year(),
	}, nil
}

// Delete removes an existing Song from the database
func (s *Song) Delete() error {
	return DB.DeleteSong(s)
}

// Load pulls an existing Song from the database
func (s *Song) Load() error {
	return DB.LoadSong(s)
}

// Save creates a new Song in the database
func (s *Song) Save() error {
	return DB.SaveSong(s)
}

// Update updates an existing Song in the database
func (s *Song) Update() error {
	return DB.UpdateSong(s)
}

// Stream generates a binary file stream from this Song's file location
func (s Song) Stream() (io.ReadSeeker, error) {
	// Attempt to open the file associated with this song
	return os.Open(s.FileName)
}

// SongSlice represents a slice of songs, and provides convenience methods to access their
// aggregate properties
type SongSlice []Song

// Length returns the total duration of a slice of songs
func (s SongSlice) Length() int {
	// Iterate and sum duration
	length := 0
	for _, song := range s {
		length += song.Length
	}

	return length
}
