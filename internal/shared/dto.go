package shared

import (
	"time"
	// You might need "gorm.io/datatypes" for datatypes.JSON if you choose to use it
	// for some JSON fields instead of custom structs, but custom structs are generally better for typed access.
)

// --- Placeholder DTOs (Define these based on your actual DTO structures) ---

type Profession struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Type string `json:"type"` // ENUM
}

type Countery struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// PersonDTO placeholder
type PersonDTO struct {
	Id                int          `json:"id"`
	Description       string       `json:"description"`
	Name              string       `json:"name"`
	Family            string       `json:"family"`
	NickName          string       `json:"nickname"`
	Professions       []Profession `json:"professions,omitempty"`
	ImageURL          *string      `json:"imageUrl,omitempty"`
	BannerUrl         string       `json:"bannerUrl"`
	MediaBannerURL    *string      `json:"mediaBannerUrl,omitempty"`
	SmallBannerURL    *string      `json:"smallBannerUrl,omitempty"`
	LogoURL           *string      `json:"logoUrl,omitempty"`
	BirthPlaceCity    *string      `json:"birthPlaceCity,omitempty"`
	BirthPlaceCountry *Countery    `json:"birthPlaceCountry,omitempty"`
	BirthDate         *time.Time   `json:"birthDate,omitempty"`
	Galleries         []string     `json:"galleries,omitempty"`
}

// PollQuestionAnswerDTO placeholder
type PollQuestionAnswerDTO struct {
	Id    string `json:"id"`
	Title string `json:"title"`
}

// PollQuestionDTO placeholder
type PollQuestionDTO struct {
	Id      string                  `json:"id"`
	Title   string                  `json:"title"`
	Answers []PollQuestionAnswerDTO `json:"answers"`
}

// --- Helper Structs for Embedded JSONB Objects ---

// AdvertisementLink is for the 'link' field in Advertisement
type AdvertisementLink struct {
	PlayLink     string `json:"playLink"`
	FileHash     string `json:"fileHash"`
	LinkType     string `json:"linkType"`
	OriginalLink string `json:"originalLink"`
}

// AlbumImage is for the 'image' field in Album
type AlbumImage struct {
	ImageURL  *string `json:"imageUrl,omitempty"`
	BannerUrl *string `json:"bannerUrl,omitempty"`
}

// AlbumGenre is for the 'genre' field in Album
type AlbumGenre struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// AudioBookLink is for the 'link' field in AudioBook
type AudioBookLink struct {
	PlayLink string `json:"playLink"`
	FileHash string `json:"fileHash"`
}

// AudioBookImage is for the 'image' field in AudioBook
type AudioBookImage struct {
	ImageURL  *string `json:"imageUrl,omitempty"`
	BannerUrl *string `json:"bannerUrl,omitempty"`
}

// AudioBookGenre is for the 'genre' field in AudioBook
type AudioBookGenre struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// MovieImage is for the 'image' field in Movie/Series
type MovieImage struct {
	ImageURL        string  `json:"imageUrl"`
	BannerUrl       *string `json:"bannerUrl,omitempty"`
	MobileBannerUrl *string `json:"mobileBannerUrl,omitempty"`
}

type MagazineImage struct {
	ImageURL  string  `json:"imageUrl"`
	BannerUrl *string `json:"bannerUrl,omitempty"`
}

// MovieLink is for the 'link' field in Movie
type MovieLink struct {
	PlayLink string `json:"playLink"`
	FileHash string `json:"fileHash"`
}

// MovieGenre is for the 'genres' field in Movie/Series (if not linking to main Genre table)
type MovieGenre struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// MusicImage is for the 'image' field in Music
type MusicImage struct {
	ImageURL  *string `json:"imageUrl,omitempty"`
	BannerUrl *string `json:"bannerUrl,omitempty"`
}

// MusicLink is for the 'link' field in Music
type MusicLink struct {
	PlayLink string `json:"playLink"`
	FileHash string `json:"fileHash"`
}

// PodcastImage is for the 'image' field in Podcast
type PodcastImage struct {
	ImageURL  *string `json:"imageUrl,omitempty"`
	BannerUrl *string `json:"bannerUrl,omitempty"`
}

// PodcastLink is for the 'link' field in Podcast
type PodcastLink struct {
	PlayLink string `json:"playLink"`
	FileHash string `json:"fileHash"`
}

// PodcastGenre is for the 'genre' field in Podcast
type PodcastGenre struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// SeriesEpisodeImage is for the 'image' field in SeriesEpisode
type SeriesEpisodeImage struct {
	ImageURL *string `json:"imageUrl,omitempty"`
}

// SeriesEpisodeLink is for the 'link' field in SeriesEpisode
type SeriesEpisodeLink struct {
	PlayLink string `json:"playLink"`
	FileHash string `json:"fileHash"`
}

// SliderImage is for the 'image' field in Slider
type SliderImage struct {
	ImageURL       string  `json:"imageUrl"`
	MediumImageUrl *string `json:"mediumImageUrl,omitempty"`
	SmallImageUrl  *string `json:"smallImageUrl,omitempty"`
	LogoImageUrl   *string `json:"logoImageUrl,omitempty"`
}

type VideoImage struct {
	ImageURL  string  `json:"imageUrl"`
	BannerUrl *string `json:"bannerUrl,omitempty"`
}

type VideoLink struct {
	PlayLink string `json:"playLink"`
	FileHash string `json:"fileHash"`
}

// --- GORM Models ---

type Advertisement struct {
	ContentId    int64             `gorm:"primaryKey;type:bigint;column:contentId"`
	SkipDuration int32             `gorm:"not null"`
	Link         AdvertisementLink `gorm:"not null;type:jsonb;default:'{}'"`
	ViewCount    int32             `gorm:"not null;default:0"`
	Synced       bool              `gorm:"not null"`
}

type Album struct {
	ContentId   int64       `gorm:"primaryKey;type:bigint;column:contentId"`
	EntityId    int64       `gorm:"not null;type:bigint;column:entityId"`
	Description string      `gorm:"not null;default:'';column:description"`         // Assuming 'description' is not nullable
	Image       AlbumImage  `gorm:"not null;type:jsonb;default:'{}';column:image"`  // Stored as JSON
	Agents      []PersonDTO `gorm:"not null;type:jsonb;default:'[]';column:agents"` // Stored as JSON array
	Genre       AlbumGenre  `gorm:"not null;type:jsonb;default:'{}';column:genre"`  // Stored as JSON
	Name        string      `gorm:"not null;column:name"`                           // Assuming 'name' is not nullable
}

type AudioBook struct {
	ContentId   int64          `gorm:"primaryKey;type:bigint;column:contentId"`
	EntityId    int64          `gorm:"type:bigint;column:entityId"`
	Description string         `gorm:"default:'';type:varchar;column:description"`
	Ages        *int32         `gorm:"type:integer;default:0;column:ages"`
	Link        AudioBookLink  `gorm:"not null;type:jsonb;default:'{}';column:link"`
	Image       AudioBookImage `gorm:"not null;type:jsonb;default:'{}';column:image"`
	// AudiobookAlbumContentId *int64         // Foreign Key for AudiobookAlbum
	AudiobookAlbumContentId *AudiobookAlbum `gorm:"foreignKey:AudiobookAlbumContentId;column:audiobookAlbumContentId"`
	Genre                   AudioBookGenre  `gorm:"not null;type:jsonb;default:'{}';column:genre"`
	Agents                  []PersonDTO     `gorm:"not null;type:jsonb;default:'[]';column:agents"`
	Name                    string          `gorm:"not null;type:varchar;column:name"`
	PublishDate             *time.Time      `gorm:"type:timestamptz;column:publishDate"`
	Duration                *int            `gorm:"default:0;column:duration"`
}

type AudiobookAlbum struct {
	ContentId   int64          `gorm:"type:bigint;primaryKey;column:contentId"`
	EntityId    int64          `gorm:"type:bigint;column:entityId"`
	Description string         `gorm:"default:'';type:varchar;column:description"`
	Ages        *int32         `gorm:"type:integer;default:0;column:ages"`
	Image       AudioBookImage `gorm:"not null;type:jsonb;default:'{}';column:image"`
	Genre       AudioBookGenre `gorm:"not null;type:jsonb;default:'{}';column:genre"`
	Agents      []PersonDTO    `gorm:"not null;type:jsonb;default:'[]';column:agents"`
	PublishDate *time.Time     `gorm:"type:timestamptz;column:publishDate"`
	// Audiobooks  []AudioBook    `gorm:"foreignKey:AudiobookAlbumId;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Duration int    `gorm:"not null;column:duration"` // Assuming not null
	Name     string `gorm:"not null;type:varchar;column:name"`
}

type EntityInfo struct {
	ContentId int64  `gorm:"primaryKey;type:bigint;column:contentId"`
	Type      string `gorm:"not null;type:varchar;column:type"`
}

type Genre struct {
	ContentId int64   `gorm:"primaryKey;type:bigint;column:contentId"`
	Code      string  `gorm:"not null;uniqueIndex;type:varchar;column:code"`
	Name      string  `gorm:"not null;type:varchar;column:name"`
	ImageURL  *string `gorm:"type:varchar;column:imageUrl"`
	Enable    bool    `gorm:"default:true;not null;column:enable"`
}

type Magazine struct {
	ContentId int64         `gorm:"primaryKey;type:bigint;column:contentId"`
	Title     string        `gorm:"not null;type:varchar;column:title"`
	LongText  string        `gorm:"not null;type:text;column:longText"`
	Text      *string       `gorm:"type:text;column:text"`
	Image     MagazineImage `gorm:"not null;type:jsonb;default:'{}';column:image"`
}

type Music struct {
	ContentId      int64        `gorm:"primaryKey;type:bigint"`
	EntityId       int64        `gorm:"type:bigint"`
	Description    string       `gorm:"default:''"`
	Image          MusicImage   `gorm:"not null;type:jsonb;default:'{}'"`
	Link           MusicLink    `gorm:"not null;type:jsonb;default:'{}'"`
	AlbumContentId *int64       // Foreign key for Album
	Album          *Album       `gorm:"foreignKey:AlbumContentId"`        // Belongs to Album
	Genres         []MovieGenre `gorm:"not null;type:jsonb;default:'[]'"` // Embedded JSON
	Agents         []PersonDTO  `gorm:"not null;type:jsonb;default:'[]'"`
	Name           string       `gorm:"not null"`
	PublishDate    *time.Time   `gorm:"type:timestamptz"`
	CreatedAt      time.Time    `gorm:"type:timestamptz;autoCreateTime"`
}

type Movie struct {
	ContentId        int64        `gorm:"primaryKey;type:bigint;column:contentId"`
	NameFa           string       `gorm:"not null;type:varchar;column:nameFa"`
	Enable           bool         `gorm:"default:true;not null;column:enable"`
	EntityId         *int64       `gorm:"type:bigint;column:entityId"`
	PostId           *int64       `gorm:"type:bigint;column:postId"`
	NameEn           *string      `gorm:"type:varchar;column:nameEn"`
	Description      string       `gorm:"not null;type:varchar;column:description"`
	Image            MovieImage   `gorm:"not null;type:jsonb;default:'{}';column:image"`
	Link             MovieLink    `gorm:"not null;type:jsonb;default:'{}';column:link"`
	Ages             *int         `gorm:"column:ages"`
	Company          *string      `gorm:"type:varchar;column:company"`
	ImdbCode         *string      `gorm:"type:varchar;column:imdbCode"`
	ImdbRate         *float64     `gorm:"type:real;column:imdbRate"`
	YearsOfBroadcast *int         `gorm:"type:integer;column:yearsOfBroadcast"`
	Duration         *int         `gorm:"column:duration"`
	PlayLink         *string      `gorm:"type:varchar;column:playLink"`
	FileHash         *string      `gorm:"type:varchar;column:fileHash"`
	Genres           []MovieGenre `gorm:"not null;type:jsonb;default:'[]'"`
	Casts            []PersonDTO  `gorm:"not null;type:jsonb;default:'[]'"`
}

type Page struct {
	ContentId int64   `gorm:"primaryKey;type:bigint;column:contentId"`
	Name      *string `gorm:"type:varchar;column:name"`
	Type      string  `gorm:"not null;type:varchar;column:type"`
	Tabs      []*Tab  `gorm:"many2many:page_tabs_tab"`
}

// type PageTabsTab struct {
// 	PageContentId int `gorm:"primaryKey;type:bigint;"`
// 	TabContentId  int `gorm:"primaryKey;type:bigint;"`
// }

type Podcast struct {
	ContentId   int64        `gorm:"primaryKey;type:bigint"`
	EntityId    int64        `gorm:"not null;type:bigint"`
	Description string       `gorm:"default:'';not null; type:varchar"`
	Ages        *int32       `gorm:"type:integer;default:0"`
	Image       PodcastImage `gorm:"not null;type:jsonb;default:'{}'"`
	Link        PodcastLink  `gorm:"not null;type:jsonb;default:'{}'"`
	Agents      []PersonDTO  `gorm:"not null;type:jsonb;default:'[]'"`
	Duration    *int         `gorm:"default:0"`
	// PodcastAlbumId *int64        // Foreign Key
	PodcastAlbumContentId *PodcastAlbum `gorm:"foreignKey:PodcastAlbumContentId;column:podcastAlbumContentId"`
	Genre                 PodcastGenre  `gorm:"not null;type:jsonb;default:'{}'"`
	Name                  string        `gorm:"not null;type:varchar"`
	PublishDate           *time.Time    `gorm:"type:timestamptz"`
}

type PodcastAlbum struct {
	ContentId   int64        `gorm:"primaryKey;type:bigint"`
	EntityId    int64        `gorm:"type:bigint"`
	Description string       `gorm:"default:'';not null;type:varchar"`
	Ages        *int32       `gorm:"type:integer;default:0"`
	Image       PodcastImage `gorm:"not null;type:jsonb;default:'{}'"`
	Genre       PodcastGenre `gorm:"not null;type:jsonb;default:'{}'"`
	Agents      []PersonDTO  `gorm:"not null;type:jsonb;default:'[]'"`
	PublishDate *time.Time   `gorm:"type:timestamptz"`
	// Podcasts    []Podcast    `gorm:"foreignKey:PodcastAlbumId;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Duration int    `gorm:"not null"`
	Name     string `gorm:"not null;not null;type:varchar"`
}

type Poll struct {
	ContentId int64             `gorm:"primaryKey;type:bigint"`
	Questions []PollQuestionDTO `gorm:"not null;type:jsonb"` // Needs PollQuestionDTO defined
	Title     string            `gorm:"not null; type:varchar"`
	Enable    bool              `gorm:"not null"`
}

// type User struct {
// 	Id          string  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
// 	Name        *string `gorm:"type:varchar"`
// 	UserId      *int64  `gorm:"type:bigint;uniqueIndex"`
// 	SsoId       *string `gorm:"uniqueIndex;type:varchar"`
// 	Username    *string `gorm:"uniqueIndex;type:varchar"`
// 	Phonenumber *string `gorm:"uniqueIndex;type:varchar"`
// 	// Tokens      []Token   `gorm:"foreignKey:UserId"` // UserId in Token model
// 	// Usage       []Usage   `gorm:"foreignKey:UserId"` // UserId in Usage model
// 	CreatedAt time.Time `gorm:"type:timestamptz;autoCreateTime"`
// 	UpdateAt  time.Time `gorm:"type:timestamptz;autoUpdateTime"`
// }

// type PollAnswer struct {
// 	Id                   string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
// 	PollId               int32     `gorm:"not null;index:idx_poll_user_answer,unique"`              // Part of composite unique index
// 	PollQuestionId       string    `gorm:"not null;type:varchar;index:idx_poll_user_answer,unique"` // Part of composite unique index
// 	PollQuestionAnswerId string    `gorm:"not null;type:varchar;index:idx_poll_user_answer,unique"` // Part of composite unique index
// 	Synced               bool      `gorm:"not null;default:false"`
// 	UserId               string    `gorm:"not null;type:uuid;index:idx_poll_user_answer,unique"` // Part of composite unique index
// 	User                 User      `gorm:"foreignKey:UserId"`                                    // Links to User struct
// 	CreatedAt            time.Time `gorm:"type:timestamptz;autoCreateTime"`
// }

type Section struct {
	ContentId  int64   `gorm:"primaryKey;type:bigint"`
	Name       *string `gorm:"type:varchar;"`
	EntityType string  `gorm:"not null"`
	Action     *string `gorm:"type:varchar;"`
	CardType   *string `gorm:"type:varchar;"`
	Priority   *int32
	// Contents   []SectionContent `gorm:"foreignKey:SectionContentId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // Assuming SectionContent has SectionContentId
}

type SectionContent struct {
	ContentId         int64  `gorm:"primaryKey;type:bigint"`
	EntityContentType string `gorm:"not null;type:varchar"`
	EntityContentId   int64  `gorm:"not null;type:bigint"`
	Priority          int    `gorm:"not null;default:0"`
	// SectionContentId  *int64   // Foreign key to Section
	SectionContentId *Section `gorm:"foreignKey:SectionContentId;column:sectionContentId"`
}

type Series struct {
	ContentId        int64        `gorm:"primaryKey;type:bigint"`
	EntityId         *int64       `gorm:"type:bigint"`
	PostId           *int64       `gorm:"type:bigint"`
	NameFa           string       `gorm:"not null;type:varchar"`
	NameEn           *string      `gorm:"type:varchar"`
	Description      string       `gorm:"not null;type:varchar"`
	Image            MovieImage   `gorm:"not null;type:jsonb;default:'{}'"`
	Ages             *int32       `gorm:"type:integer"`
	Company          *string      `gorm:"type:varchar"`
	ImdbCode         *string      `gorm:"type:varchar"`
	ImdbRate         *float64     `gorm:"type:real"`
	YearsOfBroadcast *int32       `gorm:"type:integer"`
	Genres           []MovieGenre `gorm:"not null;type:jsonb;default:'[]'"`
	Casts            []PersonDTO  `gorm:"not null;type:jsonb;default:'[]'"`
	// Seasons          []SeriesSeason `gorm:"foreignKey:SeriesContentId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // SeriesContentId in SeriesSeason
}

type SeriesSeason struct {
	ContentId int64   `gorm:"primaryKey;type:bigint"`
	Index     int64   `gorm:"not null;type:bigint"`
	EntityId  *int64  `gorm:"type:bigint"`
	Name      string  `gorm:"not null;type:varchar"`
	NameEn    *string `gorm:"type:varchar"`
	// SeriesContentId *int64          // Foreign Key to Series
	SeriesContentId *Series `gorm:"foreignKey:SeriesContentId;column:seriesContentId"`
	// Episodes        []SeriesEpisode `gorm:"foreignKey:SeasonContentId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type SeriesEpisode struct {
	ContentId int64              `gorm:"primaryKey;type:bigint"`
	EntityId  *int64             `gorm:"type:bigint"`
	Index     int64              `gorm:"not null;type:bigint"`
	Name      string             `gorm:"not null;type:varchar"`
	NameEn    *string            `gorm:"type:varchar"`
	Image     SeriesEpisodeImage `gorm:"not null;type:jsonb;default:'{}'"`
	Link      SeriesEpisodeLink  `gorm:"not null;type:jsonb;default:'{}'"`
	// SeasonContentId *int64             // Foreign Key to SeriesSeason
	SeasonContentId *SeriesSeason `gorm:"foreignKey:SeasonContentId;column:seasonContentId"`
}

type Slider struct {
	ContentId   int64       `gorm:"primaryKey;type:bigint"`
	Image       SliderImage `gorm:"not null;type:jsonb;default:'{}'"`
	Type        *string     `gorm:"type:varchar"`
	EntityType  *string     `gorm:"type:varchar"`
	EntityId    *int64      `gorm:"type:bigint"`
	ButtonTitle *string     `gorm:"type:varchar"`
	Link        *string     `gorm:"type:varchar"`
	Tabs        []*Tab      `gorm:"many2many:slider_tabs_tab;"`
}

type TermsConditions struct {
	ContentId int64  `gorm:"primaryKey;type:bingint"`
	Name      string `gorm:"not null;type:varchar"`
	Content   string `gorm:"type:text;not null"`
}

type Tab struct {
	ContentId int64      `gorm:"primaryKey;type:bigint;column:contentId"`
	Name      string     `gorm:"not null;type:varchar;column:name"`
	Type      string     `gorm:"not null;type:varchar;column:type"`
	Pages     []*Page    `gorm:"many2many:page_tabs_tab;"`
	Sections  []*Section `gorm:"many2many:tab_sections_section;"`
}

type Video struct {
	ContentId   int64      `gorm:"primaryKey;type:bigint;column:contentId"`
	Name        string     `gorm:"not null;type:varchar;column:name"`
	Description *string    `gorm:"type:varchar;column:description"`
	Image       VideoImage `gorm:"not null;type:jsonb;default:'{}';column:image"`
	Link        VideoLink  `gorm:"not null;type:jsonb;default:'{}';column:link"`
	Duration    int32      `gorm:"not null;type:integer;column:duration"`
	ReleaseTime time.Time  `gorm:"not null;type:timestamptz;column:releaseTime"`
}

type Updater struct {
	ContentId         int64 `gorm:"primaryKey;type:bigint;column:contentId"`
	LastFromTimeStamp int64 `gorm:"not null;default:0;type:bigint;column:lastFromTimeStamp"`
	UniqueFlag        bool  `gorm:"not null;default:false;column:uniqueFlag;index:,unique"`
}

var AutoMigrateList = []any{
	&Advertisement{},
	&Album{},
	&AudioBook{},
	&AudiobookAlbum{},
	&EntityInfo{},
	&Genre{},
	&Magazine{},
	&Music{},
	&Movie{},
	&Page{},
	&Podcast{},
	&PodcastAlbum{},
	&Poll{},
	&Section{},
	&SectionContent{},
	&Series{},
	&SeriesSeason{},
	&SeriesEpisode{},
	&Slider{},
	&TermsConditions{},
	&Tab{},
	&Video{},
}
