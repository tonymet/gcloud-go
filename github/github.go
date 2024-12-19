package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

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
	T     http.RoundTripper
	token string
}

type GithubClient struct {
	Client *http.Client
}

func (gt *GithubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", "Bearer "+gt.token)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	return gt.T.RoundTrip(req)
}

func AuthorizeClient(token string) GithubClient {
	return GithubClient{Client: &http.Client{Transport: &GithubTransport{token: token, T: http.DefaultTransport}}}
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
func GenerateCreateReleaseResponse(res *http.Response) (*CreateReleaseResponse, error) {
	bodyBuffer, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var asset CreateReleaseResponse
	err = json.Unmarshal(bodyBuffer, &asset)
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

// upload release asset file to github.
// uses io.Reader instead of os.File
func (gc GithubClient) UploadReleaseAsset(owner,
	repo string, id int64, filename string, file io.Reader,
	contentLength int64, mediaType string) (*ReleaseAsset, *http.Response, error) {
	if mediaType == "" {
		return nil, nil, fmt.Errorf("mediaType is unset")
	}
	var q = make(url.Values)
	q["name"] = []string{filename}
	u := url.URL{
		Scheme:   "https",
		Host:     "uploads.github.com",
		Path:     fmt.Sprintf("/repos/%s/%s/releases/%d/assets", owner, repo, id),
		RawQuery: q.Encode(),
	}
	req, err := http.NewRequest("POST", u.String(), file)
	if err != nil {
		return nil, nil, err
	}
	req.ContentLength = contentLength
	req.Header.Add("Content-type", mediaType)
	res, err := gc.Client.Do(req)
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

func (gc GithubClient) GetReleaseByTag(owner, repo, tag string) (*CreateReleaseResponse, *http.Response, error) {
	u := url.URL{
		Scheme: "https",
		Host:   "api.github.com",
		Path:   fmt.Sprintf("/repos/%s/%s/releases/tags/%s", owner, repo, tag),
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, nil, err
	}
	res, err := gc.Client.Do(req)
	if err != nil {
		return nil, res, err
	}
	asset, err := GenerateCreateReleaseResponse(res)
	if err != nil {
		return nil, res, err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, res, responseError{res: res}
	}
	return asset, res, nil
}

func createReleaseResponseReader(release *CreateReleaseResponse) (io.Reader, error) {
	buf, err := json.Marshal(*release)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf), nil
}
func (gc GithubClient) GithubCreateRelease(owner, repo string, releaseJson CreateReleaseResponse) (*CreateReleaseResponse, *http.Response, error) {
	u := url.URL{
		Scheme: "https",
		Host:   "api.github.com",
		Path:   fmt.Sprintf("/repos/%s/%s/releases", owner, repo),
	}
	encoded, err := createReleaseResponseReader(&releaseJson)
	if err != nil {
		return nil, nil, err
	}
	req, err := http.NewRequest("POST", u.String(), encoded)
	if err != nil {
		return nil, nil, err
	}
	res, err := gc.Client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	asset, err := GenerateCreateReleaseResponse(res)
	if err != nil {
		return nil, nil, err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, nil, responseError{res: res}
	}
	return asset, res, nil
}

type GithubArgs struct {
	Owner, Repo, Commit, File, Tag, KeyPath, SignatureFile, Token string
}
