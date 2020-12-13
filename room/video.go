package room

import (
	"github.com/google/uuid"
	iso8601 "github.com/senseyeio/duration"
	"github.com/syncedvideo/backend/youtube"
)

// Provider to identify the video provider
type Provider string

// YouTubeProvider for Youtube videos
const YouTubeProvider = Provider("youtube")

// Video represents a video that can be
// played by a VideoPlayer or added to a VideoQueue
type Video struct {
	ID          string              `json:"id"`
	Provider    Provider            `json:"provider"`
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Duration    int                 `json:"duration"`
	Thumbnail   string              `json:"thumbnail"`
	AddedBy     *User               `json:"addedBy"`
	Votes       map[uuid.UUID]*User `json:"votes"`
	Statistics  videoStatistics     `json:"statistics"`
}

type videoStatistics struct {
	ViewCount    uint64 `json:"viewCount"`
	LikeCount    uint64 `json:"likeCount"`
	DislikeCount uint64 `json:"dislikeCount"`
}

// ToggleVote of video
func (v *Video) ToggleVote(user *User) {
	_, voted := v.Votes[user.ID]
	if voted {
		delete(v.Votes, user.ID)
		return
	}
	v.Votes[user.ID] = user
}

// VideoSearch handles the YouTube video search
type VideoSearch struct {
	Query         string   `json:"query"`
	Videos        []*Video `json:"videos"`
	youTubeAPIkey string   `json:"-"`
}

// NewVideoSearch returns a new VideoSearch
func NewVideoSearch(youTubeAPIkey string) *VideoSearch {
	return &VideoSearch{
		youTubeAPIkey: youTubeAPIkey,
	}
}

// Do execute a YouTube video search
func (search *VideoSearch) Do(query string) (*VideoSearch, error) {
	yt := youtube.New(search.youTubeAPIkey)
	ytVideos, err := yt.SearchVideos(query)
	if err != nil {
		return nil, err
	}

	videos := []*Video{}
	for _, ytVideo := range ytVideos {
		duration, _ := iso8601.ParseISO8601(ytVideo.ContentDetails.Duration)
		videos = append(videos, &Video{
			ID:          ytVideo.ID,
			Provider:    YouTubeProvider,
			Title:       ytVideo.Snippet.Title,
			Description: ytVideo.Snippet.Description,
			Thumbnail:   ytVideo.Snippet.Thumbnails.High.Url,
			Duration:    (duration.TM * 60) + duration.TS,
			Statistics: videoStatistics{
				ViewCount:    ytVideo.Statistics.ViewCount,
				LikeCount:    ytVideo.Statistics.LikeCount,
				DislikeCount: ytVideo.Statistics.DislikeCount,
			},
		})
	}

	return &VideoSearch{
		Query:  query,
		Videos: videos,
	}, nil
}