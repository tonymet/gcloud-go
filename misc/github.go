package misc

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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

type responseError struct {
	res *http.Response
}

func (re responseError) Error() string {
	// extract body
	buf := make([]byte, 0, re.res.ContentLength+10)
	bodyBuf, err := io.ReadAll(re.res.Body)
	if err != nil {
		return err.Error()
	}
	buf = append(buf, fmt.Sprintf("Error: %d\n . Body: \n", re.res.StatusCode)...)
	buf = append(buf, bodyBuf...)
	return string(buf)
}

type GithubTransport struct {
	T http.RoundTripper
}

func (adt *GithubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "Bearer "+os.Getenv("GH_TOKEN"))
	req.Header.Set("Accept", "*/*")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	return adt.T.RoundTrip(req)
}

func getClient() *http.Client {
	return &http.Client{Transport: &GithubTransport{T: http.DefaultTransport}}
}

func ReleaseAssetResponse(res *http.Response) (*ReleaseAsset, error) {
	bodyBuffer, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var asset ReleaseAsset
	err = json.Unmarshal(bodyBuffer, &asset)
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

// upload release asset file to github.
// uses io.Reader instead of os.File
func githubUploadReleaseAsset(owner,
	repo string, id int64, filename string, file io.Reader,
	contentLength int64, mediaType string) (*ReleaseAsset, *http.Response, error) {
	if mediaType == "" {
		panic("mediaType is unset")
	}
	u := fmt.Sprintf("https://uploads.github.com/repos/%s/%s/releases/%d/assets", owner, repo, id)
	client := getClient()
	var query = make(url.Values)
	query["name"] = []string{filename}
	u = u + "?" + query.Encode()
	req, err := http.NewRequest("POST", u, file)
	req.ContentLength = contentLength
	req.Header.Add("Content-type", mediaType)
	if err != nil {
		return nil, nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	asset, err := ReleaseAssetResponse(res)
	if err != nil {
		return nil, nil, err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, nil, responseError{res: res}
	}
	return asset, res, nil
}

type GithubArgs struct {
	Owner, Repo, Commit, File, Tag, KeyPath, SignatureFile string
}
type KMSArgs struct {
	Filename, Output, Keypath string
}
