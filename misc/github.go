package misc

import (
	"context"
	"fmt"
	"github.com/google/go-github/v67/github"
	"io"
	"net/http"
	"net/url"
)

func githubUploadReleaseAsset(ctx context.Context, c *github.Client, owner, repo string, id int64, query url.Values, file io.Reader, contentLength int, mediaType string) (*github.ReleaseAsset, *github.Response, error) {
	u := fmt.Sprintf("repos/%s/%s/releases/%d/assets", owner, repo, id)
	// add url encoding
	u = u + "?" + query.Encode()
	modRequest := func(req *http.Request) {
		//req.PostForm.Add("Name", query.Get("Name"))
	}
	req, err := c.NewUploadRequest(u, file, int64(contentLength), mediaType, modRequest)
	if err != nil {
		return nil, nil, err
	}

	asset := new(github.ReleaseAsset)
	resp, err := c.Do(ctx, req, asset)
	if err != nil {
		return nil, resp, err
	}
	return asset, resp, nil
}

type GithubArgs struct {
	Owner, Repo, Commit, File, Tag, KeyPath, SignatureFile string
}
type KMSArgs struct {
	Filename, Output, Keypath string
}
