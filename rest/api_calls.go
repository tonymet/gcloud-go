package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	ppath "path"
	"strings"

	fs "github.com/tonymet/gcloud-go/fs"
	"golang.org/x/sync/errgroup"

	"cloud.google.com/go/auth"
	"cloud.google.com/go/auth/credentials"
	"cloud.google.com/go/auth/httptransport"
)

const REST_PAGE_SIZE = 1000

var FlagConn *int

type JWTConfig struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
}

// an http client with
// oauth2 credentials attached
type AuthorizedHTTPClient struct {
	*http.Client
	authCredentials *auth.Credentials
}

// use ADC creds to authorize the default client
func AuthorizeClientDefault(ctx context.Context) (*AuthorizedHTTPClient, error) {
	if creds, err := credentials.DetectDefault(&credentials.DetectOptions{
		Scopes: []string{
			"https://www.googleapis.com/auth/firebase",
			"https://www.googleapis.com/auth/firebase.readonly",
			"https://www.googleapis.com/auth/devstorage.read_only",
		},
	}); err != nil {
		panic(err)
	} else if client, err := httptransport.NewClient(&httptransport.Options{
		Credentials: creds,
	}); err != nil {
		panic(err)
	} else {
		log.Printf("credentials authorized.")
		return &AuthorizedHTTPClient{client, creds}, nil
	}
}

// rest call to upload file list to firebase
func (client *AuthorizedHTTPClient) RestUploadFileList(ctx context.Context, versionId string, manifestSet []VersionPopulateFilesReturn, stagingDir string) error {
	work, ctx := errgroup.WithContext(ctx)
	for _, manifest := range manifestSet {
		if len(manifest.UploadRequiredHashes) == 0 {
			continue
		}
		for _, shaHash := range manifest.UploadRequiredHashes {
			work.Go(func() error {
				if f, err := os.Open(ppath.Join(stagingDir, shaHash)); err != nil {
					return err
				} else if err := client.RestUploadFile(ctx, f, shaHash, versionId); err != nil {
					return err
				}
				return nil
			})
		}
		log.Printf("upload complete: %d files", len(manifest.UploadRequiredHashes))
	}
	return work.Wait()
}

// rest set status to published
func (client *AuthorizedHTTPClient) RestVersionSetStatus(versionId string, status string) (r VersionStatusUpdateReturn, e error) {
	resource := "https://firebasehosting.googleapis.com/v1beta1/" + versionId + "?update_mask=status"
	// set up shas
	var bodyJson VersionStatusUpdateRequestBody
	bodyJson.Status = status
	if bodyBuffer, err := json.Marshal(bodyJson); err != nil {
		return VersionStatusUpdateReturn{}, err
	} else if req, err := http.NewRequest(http.MethodPatch, resource, bytes.NewReader(bodyBuffer)); err != nil {
		return VersionStatusUpdateReturn{}, err
	} else {
		req.Header.Add("Content-Type", "application/json")
		if res, err := client.Do(req); err != nil {
			panic(err)
		} else if res.StatusCode < 200 || res.StatusCode > 299 {
			panic(formatRestResponseMessage("RestVersionSetStatus", res))
		} else if bodyBytes, err := io.ReadAll(res.Body); err != nil {
			panic(err)
		} else if err := json.Unmarshal(bodyBytes, &r); err != nil {
			panic(err)
		} else {
			log.Printf("Version status set: %s", status)
			return r, nil
		}
	}
}

// rest call create new release on firebase sites
func (client *AuthorizedHTTPClient) RestReleasesCreate(site, versionId string) (r ReleasesCreateReturn, e error) {
	resource := "https://firebasehosting.googleapis.com/v1beta1/sites/" + site + "/releases?versionName=" + versionId
	// set up shas
	if req, err := http.NewRequest(http.MethodPost, resource, nil); err != nil {
		panic(err)
	} else if res, err := client.Do(req); err != nil {
		panic(err)
	} else if res.StatusCode < 200 || res.StatusCode > 299 {
		panic(formatRestResponseMessage("RestReleasesCreate", res))
	} else if bodyBytes, err := io.ReadAll(res.Body); err != nil {
		panic(err)
	} else if err := json.Unmarshal(bodyBytes, &r); err != nil {
		panic(err)
	} else {
		log.Printf("release created. id = %s", versionId)
		return r, nil
	}
}

// rest call to upload file contents
func (client *AuthorizedHTTPClient) RestUploadFile(ctx context.Context, bodyFile io.Reader, shaHash, versionId string) error {
	resource := "https://upload-firebasehosting.googleapis.com/upload/" + versionId + "/files/" + shaHash
	// set up shas
	if req, err := http.NewRequestWithContext(ctx, http.MethodPost, resource, bodyFile); err != nil {
		panic(err)
	} else {
		req.Header.Add("Content-Type", "application/octet-stream")
		if res, err := client.Do(req); err != nil {
			panic(err)
		} else if res.StatusCode < 200 || res.StatusCode > 299 {
			panic(formatRestResponseMessage("RestUploadFile", res))
		}
	}
	return nil
}

// rest populate files
func (client *AuthorizedHTTPClient) RestCreateVersionPopulateFiles(ctx context.Context, stagingDir string, versionId string) (vpfrs []VersionPopulateFilesReturn, err error) {
	resource := "https://firebasehosting.googleapis.com/v1beta1/" + versionId + ":populateFiles"
	shas := fs.ShaFiles(ctx, "./", stagingDir)
	// goroutine to send requests
	// set up shas
	vpfrs = make([]VersionPopulateFilesReturn, 0, 1)
	var chanOpen = true
	// loop over pages/ size = 1000
	for chanOpen {
		var (
			bodyJson VersionPopulateFilesRequestBody
			r        VersionPopulateFilesReturn
		)
		bodyJson.Files = make(map[string]string)
		for i := 0; i < REST_PAGE_SIZE; i++ {
			var s fs.ShaRec
			if s, chanOpen = <-shas; !chanOpen {
				break
			}
			bodyJson.Files[s.RelPath] = s.Shasum
		}
		// send api call
		if bodyBuffer, err := json.Marshal(bodyJson); err != nil {
			panic(err)
		} else if req, err := http.NewRequestWithContext(ctx, http.MethodPost, resource, bytes.NewReader(bodyBuffer)); err != nil {
			panic(err)
		} else {
			req.Header.Add("Content-Type", "application/json")
			if res, err := client.Do(req); err != nil {
				panic(err)
			} else if res.StatusCode < 200 || res.StatusCode > 299 {
				panic(formatRestResponseMessage("RestCreateVersionPopulateFiles", res))
			} else if bodyBytes, err := io.ReadAll(res.Body); err != nil {
				panic(err)
			} else if err := json.Unmarshal(bodyBytes, &r); err != nil {
				panic(err)
			} else {
				log.Printf("files populated: %d files ", len(bodyJson.Files))
				vpfrs = append(vpfrs, r)
			}
		}
	}
	return vpfrs, nil
}

// rest call to create new version on Firebase Hosting
func (client *AuthorizedHTTPClient) RestCreateVersion(site string) (r VersionCreateReturn, e error) {
	reqBody := ` 
	{
             "config": {
               "headers": [{
                 "glob": "**",
                 "headers": {
                   "Cache-Control": "max-age=1800"
                 }
               }]
             }
           }
	`
	resource := "https://firebasehosting.googleapis.com/v1beta1/sites/" + site + "/versions"
	if req, err := http.NewRequest("POST", resource, strings.NewReader(reqBody)); err != nil {
		panic(err)
	} else if resp, err := client.Do(req); err != nil {
		return VersionCreateReturn{}, err
	} else if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return VersionCreateReturn{},
			formatRestResponseMessage("RestCreateVersion", resp)
	} else if body, err := io.ReadAll(resp.Body); err != nil {
		return VersionCreateReturn{}, err
	} else if err := json.Unmarshal(body, &r); err != nil {
		panic(err)
	}
	log.Printf("version created: id = %s", r.Name)
	return r, nil
}

// transform http.response json to ResponseMessage struct
func readResponseMessage(resp *http.Response) (ResponseMessage, error) {
	var r ResponseMessage
	if bodyBytes, err := io.ReadAll(resp.Body); err != nil {
		return r, err

	} else if err := json.Unmarshal(bodyBytes, &r); err != nil {
		return r, err
	}
	return r, nil
}

// format error message
func formatRestResponseMessage(caller string, resp *http.Response) error {
	message, _ := readResponseMessage(resp)
	return fmt.Errorf("error: %s: non-200 status code = %d, error.status = %s,\n\t error.message = %s ",
		caller, resp.StatusCode, message.Error.Status, message.Error.Message)
}
