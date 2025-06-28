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

type UserPostInfo struct {
	Disliked bool `json:"disliked"`
	Favorite bool `json:"favorite"`
	Liked    bool `json:"liked"`
}
type LocalShowTechSchema struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type LocalMovieContentDetailSchema struct {
	ID               int                    `json:"id"`
	Description      string                 `json:"description"`
	ImageURL         string                 `json:"imageUrl"`
	PostID           *int64                 `json:"postId,omitempty"`
	NameEn           string                 `json:"nameEn"`
	NameFa           string                 `json:"nameFa"`
	Casts            []PersonDTO            `json:"casts"`
	Ages             int                    `json:"ages"`
	Company          string                 `json:"company"`
	IMDBCode         string                 `json:"imdbCode"`
	IMDBRate         *float64               `json:"imdbRate,omitempty"`
	YearsOFBroadcast int                    `json:"yearsOfBroadcast"`
	Genres           []MovieGenre           `json:"genres"`
	Galleries        []*string              `json:"galleries"`
	Counteries       []*Countery            `json:"counteries"`
	BannerURL        string                 `json:"bannerUrl"`
	MobileBannerURL  string                 `json:"mobileBannerUrl"`
	Duration         int                    `json:"duration"`
	ShowTech         []*LocalShowTechSchema `json:"showTech"`
	UserPostInfo     UserPostInfo           `json:"userPostInfo"`
	NumOfLikes       int                    `json:"numOfLikes"`
	NumOfDisLikes    int                    `json:"numOfDisLikes"`
	LikePercent      int                    `json:"likePercent"`
	Price            int                    `json:"price"`
	PriceWithVat     int                    `json:"priceWithVat"`
	Vat              int                    `json:"vat"`
	MultiQuality     bool                   `json:"multiQuality"`
	Saleable         bool                   `json:"saleable"`
	DiscountPercent  int                    `json:"discountPercent"`
	MultiProvider    bool                   `json:"multiProvider"`
}

type LocalSeriesEpisodeContentDetailSchema struct {
	ID            int64        `json:"id"`
	PostID        *int64       `json:"postId,omitempty"`
	Name          string       `json:"name"`
	NameEn        string       `json:"nameEn"`
	ImageURL      string       `json:"imageUrl"`
	UserPostInfo  UserPostInfo `json:"userPostInfo"`
	NumOfLikes    int          `json:"numOfLikes"`
	NumOfDisLikes int          `json:"numOfDisLikes"`
	LikePercent   int          `json:"likePercent"`
	Price         int          `json:"price"`
	PriceWithVat  int          `json:"priceWithVat"`
	Vat           int          `json:"vat"`
	Index         int          `json:"index"`
	MultiQuality  bool         `json:"multiQuality"`
	MultiProvider bool         `json:"multiProvider"`
}

type LocalSeriesSeasonContentDetailSchema struct {
	ID            int          `json:"id"`
	Description   *string      `json:"description,omitempty"`
	ImageURL      string       `json:"imageUrl"`
	NameEn        string       `json:"nameEn"`
	Name          string       `json:"name"`
	UserPostInfo  UserPostInfo `json:"userPostInfo"`
	NumOfLikes    int          `json:"numOfLikes"`
	NumOfDisLikes int          `json:"numOfDisLikes"`
	LikePercent   int          `json:"likePercent"`
	Index         int          `json:"index"`
}

type LocalSeriesContentDetailSchema struct {
	ID               int                    `json:"id"`
	Description      string                 `json:"description"`
	ImageURL         string                 `json:"imageUrl"`
	PostID           *int64                 `json:"postId,omitempty"`
	NameEn           string                 `json:"nameEn"`
	NameFa           string                 `json:"nameFa"`
	Casts            []PersonDTO            `json:"casts"`
	Ages             int                    `json:"ages"`
	Company          string                 `json:"company"`
	IMDBCode         string                 `json:"imdbCode"`
	IMDBRate         *float64               `json:"imdbRate,omitempty"`
	Summary          *string                `json:"summary,omitempty"`
	YearsOFBroadcast int                    `json:"yearsOfBroadcast"`
	Genres           []MovieGenre           `json:"genres"`
	Galleries        []*string              `json:"galleries"`
	Counteries       []*Countery            `json:"counteries"`
	BannerURL        string                 `json:"bannerUrl"`
	MobileBannerURL  string                 `json:"mobileBannerUrl"`
	ShowTech         []*LocalShowTechSchema `json:"showTech"`
	UserPostInfo     UserPostInfo           `json:"userPostInfo"`
	NumOfLikes       int                    `json:"numOfLikes"`
	NumOfDisLikes    int                    `json:"numOfDisLikes"`
	LikePercent      int                    `json:"likePercent"`
	Price            int                    `json:"price"`
	Saleable         bool                   `json:"saleable"`
	DiscountPercent  int                    `json:"discountPercent"`
}

type PodcastLanguage struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type LocalPodcastParentContentDetailSchema struct {
	ID int64 `json:"id"`
	// Description      string                 `json:"description"`
	ImageURL     string          `json:"imageUrl"`
	BannerURL    string          `json:"bannerUrl"`
	PostID       *int64          `json:"postId,omitempty"`
	NameFa       string          `json:"nameFa"`
	Genre        *PodcastGenre   `json:"genre,omitempty"`
	Duration     int             `json:"duration"`
	Size         int             `json:"size"`
	SectionCount int             `json:"sectionCount"`
	Language     PodcastLanguage `json:"language"`
	PublishDate  int             `json:"publishDate"`
	Ages         *int            `json:"ages"`
	Galleries    []*string       `json:"galleries"`
	Agents       []*PersonDTO    `json:"agents"`
	Note         *string         `json:"note"`
	Price        int             `json:"price"`
	UserPostInfo UserPostInfo    `json:"userPostInfo"`
	// NameEn           string                 `json:"nameEn"`
	// Casts            []PersonDTO            `json:"casts"`
	// Company          string                 `json:"company"`
	// IMDBCode         string                 `json:"imdbCode"`
	// IMDBRate         *float64               `json:"imdbRate,omitempty"`
	// YearsOFBroadcast int                    `json:"yearsOfBroadcast"`
	// Counteries       []*Countery            `json:"counteries"`
	// MobileBannerURL  string                 `json:"mobileBannerUrl"`
	// ShowTech         []*LocalShowTechSchema `json:"showTech"`
	// NumOfLikes       int                    `json:"numOfLikes"`
	// NumOfDisLikes    int                    `json:"numOfDisLikes"`
	// LikePercent      int                    `json:"likePercent"`
	// PriceWithVat     int                    `json:"priceWithVat"`
	// Vat              int                    `json:"vat"`
	// MultiQuality     bool                   `json:"multiQuality"`
	// Saleable         bool                   `json:"saleable"`
	// DiscountPercent  int                    `json:"discountPercent"`
	// MultiProvider    bool                   `json:"multiProvider"`
}

type LocalMovieContentSchema struct {
	Type      string                        `json:"type"`
	Content   LocalMovieContentDetailSchema `json:"content"`
	IsBuyed   bool                          `json:"isBuyed"`
	BuyedTime int                           `json:"buyedTime"`
}

type LocalSeriesContentSchema struct {
	Type      string                         `json:"type"`
	Content   LocalSeriesContentDetailSchema `json:"content"`
	IsBuyed   bool                           `json:"isBuyed"`
	BuyedTime int                            `json:"buyedTime"`
}

type LocalSeriesSeasonContentSchema struct {
	Type      string                               `json:"type"`
	Content   LocalSeriesSeasonContentDetailSchema `json:"content"`
	IsBuyed   bool                                 `json:"isBuyed"`
	BuyedTime int                                  `json:"buyedTime"`
}
type LocalSeriesEpisodeContentSchema struct {
	Type      string                                `json:"type"`
	Content   LocalSeriesEpisodeContentDetailSchema `json:"content"`
	IsBuyed   bool                                  `json:"isBuyed"`
	BuyedTime int                                   `json:"buyedTime"`
}
type LocalPodcastParentContentSchema struct {
	Type      string                                `json:"type"`
	Content   LocalPodcastParentContentDetailSchema `json:"content"`
	IsBuyed   bool                                  `json:"isBuyed"`
	BuyedTime int                                   `json:"buyedTime"`
}
type LocalMovieSchema struct {
	FileLink string `json:"fileLink"`
	MovieID  int64  `json:"movieId"`
}
type LocalSeriesSchema struct {
	SeriesID int `json:"seriesId"`
}
type LocalSeriesSeasonSchema struct {
	LocalSeriesID int64 `json:"localSeriesId"`
	SeasonID      int   `json:"seasonId"`
}
type LocalSeriesEpisodeSchema struct {
	FileLink      string `json:"fileLink"`
	LocalSeasonID int64  `json:"localSeasonId"`
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
	LocalSectionID    int64  `json:"localSectionId"`
	EntityContentID   int64  `json:"entitiContentId"`
	EntityContentType string `json:"entitiContentType"`
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
