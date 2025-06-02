package shared

import "encoding/json"

// UpdateInfo matches the JSON structure for update information.
type UpdateInfo struct {
	VersionCode int    `json:"versionCode"`
	FileURL     string `json:"fileUrl"`
}

// UpdateErr matches the JSON structure for API error messages.
type UpdateErr struct {
	Message string `json:"message"`
}

// StatusReportPayload matches the JSON structure for reporting status.
type StatusReportPayload struct {
	VersionCode   int    `json:"versionCode"`
	StatusMessage string `json:"statusMessage"`
}

// ContentUpdateRequestParams defines parameters for fetching content updates.
type ContentUpdateRequestParams struct {
	From   int64 `url:"from"`   // Timestamp
	Size   int   `url:"size"`   // Page size
	Offset int   `url:"offset"` // Page offset
}

// ContentUpdateResponse is the structure for the /contents/update API response.
type ContentUpdateResponse struct {
	Contents []GenericContentItem `json:"contents"`
	Count    int                  `json:"count"` // Remaining contents count
}

// GenericContentItem is the base structure for items in the "contents" array.
// It includes common fields and a RawMessage field for type-specific content.
type GenericContentItem struct {
	ID        int64           `json:"id"`
	Type      string          `json:"type"`
	UpdatedAt int64           `json:"updatedAt"`
	Enable    bool            `json:"enable"`
	Content   json.RawMessage `json:"content"` // Holds the type-specific content data
}

// --- Specific Content Type Structs ---

// LocalAdvertisementContent defines the structure for "local-advertisement" type.
type LocalAdvertisementSchema struct {
	FileLink     string `json:"fileLink"`
	SkipDuration int    `json:"skipDuration"`
}

type LocalPageSchema struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type LocalTabSchema struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	LocalPageIDs []int  `json:"localPageIds"`
}

type LocalSectionSchema struct {
	Name        string  `json:"name"`
	Action      *string `json:"action,omitempty"`
	CardType    *string `json:"cardType,omitempty"`
	Priority    int     `json:"priority"`
	EntityType  string  `json:"entityType"`
	LocalTabIDs []int   `json:"localTabIds"`
}
type LocalMovieSchema struct {
	FileLink string `json:"fileLink"`
	MovieID  int64  `json:"movieId"`
}
type LocalSeriesSchema struct {
	SeriesID int `json:"seriesId"`
}
type LocalSeriesSeasonSchema struct {
	LocalSeriesID int `json:"localSeriesId"`
	SeasonID      int `json:"seasonId"`
}
type LocalSeriesEpisodeSchema struct {
	FileLink      string `json:"fileLink"`
	LocalSeasonID int    `json:"localSeasonId"`
	EpisodeID     int    `json:"episodeId"`
}
type LocalMovieGenreSchema struct {
	ImageURL string `json:"imageUrl"`
	Code     string `json:"code"`
}
type LocalSliderSchema struct {
	ImageURL       string  `json:"imageUrl"`
	SmallImageURL  string  `json:"smallImageUrl"`
	MediumImageURL string  `json:"mediumImageUrl"`
	LogoImageURL   *string `json:"logoImageUrl,omitempty"`
	MovieURL       string  `json:"movieUrl"`
	ButtonTitle    *string `json:"buttonTitle,omitempty"`
	Link           *string `json:"link,omitempty"`
	LocalTabIDs    []int   `json:"localTabIds"`
	LocalContentID int     `json:"localContentId"`
}
type LocalPollAnswer struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}
type LocalPollQuestion struct {
	ID      string            `json:"id"`
	Title   string            `json:"title"`
	Answers []LocalPollAnswer `json:"answers"`
}
type LocalPollSchema struct {
	Questions []LocalPollQuestion `json:"questions"`
}
type LocalSectionContentSchema struct {
	Priority          int    `json:"priority"`
	LocalSectionID    int    `json:"localSectionId"`
	EntityContentID   int    `json:"entityContentId"`
	EntityContentType string `json:"entityContentType"`
}
type LocalPodcastSchema struct {
	PodcastID            int    `json:"podcastId"`
	LocalPodcastParentID int    `json:"localPodcastParentId"`
	FileLink             string `json:"fileLink"`
}
type LocalPodcastParentSchema struct {
	PodcastParentID int `json:"podcastParentId"`
}
type LocalAudiobookSchema struct {
	AudiobookID            int    `json:"audiobookId"`
	LocalAudiobookParentID int    `json:"localAudiobookParentId"`
	FileLink               string `json:"fileLink"`
}
type LocalAudiobookParentSchema struct {
	AudiobookParentID int `json:"audiobookParentId"`
}
type LocalMusicSchema struct {
	MusicID      int    `json:"musicId"`
	LocalAlbumID int    `json:"localAlbumId"`
	FileLink     string `json:"fileLink"`
}
type LocalAlbumSchema struct {
	AlbumID int `json:"albumId"`
}
type LocalDeviceUpdateSchema struct {
	VersionCode int    `json:"versionCode"`
	FileURL     string `json:"fileUrl"`
}
type LocalTermsConditionsSchema struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// ProcessedContent is a wrapper that can hold any of the processed content types.
// This allows a single channel or function to handle different types.
type ProcessedContentSchema struct {
	ID        int64
	Type      string
	UpdatedAt int64
	Enable    bool
	Details   interface{} // This will hold the specific content struct (e.g., LocalAdvertisementContent)
}
