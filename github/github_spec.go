package github

import (
	"time"
)

type ReleaseAsset struct {
	URL                string    `json:"url"`
	BrowserDownloadURL string    `json:"browser_download_url"`
	ID                 int       `json:"id"`
	NodeID             string    `json:"node_id"`
	Name               string    `json:"name"`
	Label              string    `json:"label"`
	State              string    `json:"state"`
	ContentType        string    `json:"content_type"`
	Size               int       `json:"size"`
	DownloadCount      int       `json:"download_count"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	Uploader           struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"uploader"`
}

type CreateReleaseResponse struct {
	URL             string    `json:"url,omitempty"`
	HTMLURL         string    `json:"html_url,omitempty"`
	AssetsURL       string    `json:"assets_url,omitempty"`
	UploadURL       string    `json:"upload_url,omitempty"`
	TarballURL      string    `json:"tarball_url,omitempty"`
	ZipballURL      string    `json:"zipball_url,omitempty"`
	DiscussionURL   string    `json:"discussion_url,omitempty"`
	ID              int64     `json:"id,omitempty"`
	NodeID          string    `json:"node_id,omitempty"`
	TagName         string    `json:"tag_name,omitempty"`
	TargetCommitish string    `json:"target_commitish,omitempty"`
	Name            string    `json:"name,omitempty"`
	Body            string    `json:"body,omitempty"`
	Draft           bool      `json:"draft,omitempty"`
	Prerelease      bool      `json:"prerelease,omitempty"`
	CreatedAt       time.Time `json:"created_at,omitempty"`
	PublishedAt     time.Time `json:"published_at,omitempty"`
	Author          struct {
		Login             string `json:"login,omitempty"`
		ID                int    `json:"id,omitempty"`
		NodeID            string `json:"node_id,omitempty"`
		AvatarURL         string `json:"avatar_url,omitempty"`
		GravatarID        string `json:"gravatar_id,omitempty"`
		URL               string `json:"url,omitempty"`
		HTMLURL           string `json:"html_url,omitempty"`
		FollowersURL      string `json:"followers_url,omitempty"`
		FollowingURL      string `json:"following_url,omitempty"`
		GistsURL          string `json:"gists_url,omitempty"`
		StarredURL        string `json:"starred_url,omitempty"`
		SubscriptionsURL  string `json:"subscriptions_url,omitempty"`
		OrganizationsURL  string `json:"organizations_url,omitempty"`
		ReposURL          string `json:"repos_url,omitempty"`
		EventsURL         string `json:"events_url,omitempty"`
		ReceivedEventsURL string `json:"received_events_url,omitempty"`
		Type              string `json:"type,omitempty"`
		SiteAdmin         bool   `json:"site_admin,omitempty"`
	} `json:"author,omitempty"`
	Assets []struct {
		URL                string    `json:"url,omitempty"`
		BrowserDownloadURL string    `json:"browser_download_url,omitempty"`
		ID                 int       `json:"id,omitempty"`
		NodeID             string    `json:"node_id,omitempty"`
		Name               string    `json:"name,omitempty"`
		Label              string    `json:"label,omitempty"`
		State              string    `json:"state,omitempty"`
		ContentType        string    `json:"content_type,omitempty"`
		Size               int       `json:"size,omitempty"`
		DownloadCount      int       `json:"download_count,omitempty"`
		CreatedAt          time.Time `json:"created_at,omitempty"`
		UpdatedAt          time.Time `json:"updated_at,omitempty"`
		Uploader           struct {
			Login             string `json:"login,omitempty"`
			ID                int    `json:"id,omitempty"`
			NodeID            string `json:"node_id,omitempty"`
			AvatarURL         string `json:"avatar_url,omitempty"`
			GravatarID        string `json:"gravatar_id,omitempty"`
			URL               string `json:"url,omitempty"`
			HTMLURL           string `json:"html_url,omitempty"`
			FollowersURL      string `json:"followers_url,omitempty"`
			FollowingURL      string `json:"following_url,omitempty"`
			GistsURL          string `json:"gists_url,omitempty"`
			StarredURL        string `json:"starred_url,omitempty"`
			SubscriptionsURL  string `json:"subscriptions_url,omitempty"`
			OrganizationsURL  string `json:"organizations_url,omitempty"`
			ReposURL          string `json:"repos_url,omitempty"`
			EventsURL         string `json:"events_url,omitempty"`
			ReceivedEventsURL string `json:"received_events_url,omitempty"`
			Type              string `json:"type,omitempty"`
			SiteAdmin         bool   `json:"site_admin,omitempty"`
		} `json:"uploader,omitempty"`
	} `json:"assets,omitempty"`
}
