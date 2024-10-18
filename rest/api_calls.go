package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	fs "main/fs"
	"net/http"
	"os"
	ppath "path"
	"strings"

	"cloud.google.com/go/auth/credentials"
	"cloud.google.com/go/auth/httptransport"
)

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

func AuthorizeClientDefault(ctx context.Context, flagCred string) (*http.Client, error) {
	if creds, err := credentials.DetectDefault(&credentials.DetectOptions{
		Scopes: []string{
			"https://www.googleapis.com/auth/firebase",
			"https://www.googleapis.com/auth/firebase.readonly",
		},
	}); err != nil {
		panic(err)
	} else if client, err := httptransport.NewClient(&httptransport.Options{
		Credentials: creds,
	}); err != nil {
		panic(err)
	} else {
		log.Printf("credentials authorized")
		return client, nil
	}
}

func RestUploadFileList(client *http.Client, versionId string, manifest VersionPopulateFilesReturn, stagingDir string) error {
	// workerPool to make http requests.
	// jobs = input sha, results = output
	httpWorker := func(jobs <-chan string, results chan<- error) {
		for shaHash := range jobs {
			if f, err := os.Open(ppath.Join(stagingDir, shaHash)); err != nil {
				results <- err
			} else if err := RestUploadFile(client, f, shaHash, versionId); err != nil {
				results <- err
			}
			results <- nil
		}
	}
	var numJobs = len(manifest.UploadRequiredHashes)
	jobs, results := make(chan string, numJobs), make(chan error, numJobs)
	// start workers
	for w := 1; w <= *FlagConn; w++ {
		go httpWorker(jobs, results)
	}
	// send each sha to jobs channel.
	for _, shaHash := range manifest.UploadRequiredHashes {
		jobs <- shaHash
	}
	close(jobs)
	// read from results.  return error
	// TODO: better error handling (e.g. dlq)
	for a := 1; a <= numJobs; a++ {
		err := <-results
		if err != nil {
			return err
		}
	}
	log.Printf("upload complete: %d files", len(manifest.UploadRequiredHashes))
	return nil
}

func RestVersionSetStatus(client *http.Client, versionId string, status string) (r VersionStatusUpdateReturn, e error) {
	resource := "https://firebasehosting.googleapis.com/v1beta1/" + versionId + "?update_mask=status"
	// set up shas
	var bodyJson VersionStatusUpdateRequestBody
	bodyJson.Status = status
	if bodyBuffer, err := json.Marshal(bodyJson); err != nil {
		panic(err)
	} else if req, err := http.NewRequest(http.MethodPatch, resource, bytes.NewReader(bodyBuffer)); err != nil {
		panic(err)
	} else {
		req.Header.Add("Content-Type", "application/json")
		if res, err := client.Do(req); err != nil {
			panic(err)
		} else if res.StatusCode < 200 || res.StatusCode > 299 {
			panic(fmt.Sprintf("http error: status = %s, resource = %s\n", res.Status, res.Request.URL))
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

func RestReleasesCreate(client *http.Client, site, versionId string) (r ReleasesCreateReturn, e error) {
	resource := "https://firebasehosting.googleapis.com/v1beta1/sites/" + site + "/releases?versionName=" + versionId
	// set up shas
	if req, err := http.NewRequest(http.MethodPost, resource, nil); err != nil {
		panic(err)
	} else if res, err := client.Do(req); err != nil {
		panic(err)
	} else if res.StatusCode < 200 || res.StatusCode > 299 {
		panic(fmt.Sprintf("http error: status = %s, resource = %s\n", res.Status, res.Request.URL))
	} else if bodyBytes, err := io.ReadAll(res.Body); err != nil {
		panic(err)
	} else if err := json.Unmarshal(bodyBytes, &r); err != nil {
		panic(err)
	} else {
		log.Printf("release created. id = %s", versionId)
		return r, nil
	}
}

func RestUploadFile(client *http.Client, bodyFile io.Reader, shaHash, versionId string) error {
	resource := "https://upload-firebasehosting.googleapis.com/upload/" + versionId + "/files/" + shaHash
	// set up shas
	if req, err := http.NewRequest(http.MethodPost, resource, bodyFile); err != nil {
		panic(err)
	} else {
		req.Header.Add("Content-Type", "application/octet-stream")
		if res, err := client.Do(req); err != nil {
			panic(err)
		} else if res.StatusCode < 200 || res.StatusCode > 299 {
			panic(fmt.Sprintf("http error: status = %s, resource = %s\n", res.Status, res.Request.URL))
		}
	}
	return nil
}
func RestCreateVersionPopulateFiles(client *http.Client, shas fs.ShaList, versionId string) (r VersionPopulateFilesReturn, err error) {
	resource := "https://firebasehosting.googleapis.com/v1beta1/" + versionId + ":populateFiles"
	// set up shas
	var bodyJson VersionPopulateFilesRequestBody
	bodyJson.Files = make(map[string]string)
	for _, s := range shas {
		bodyJson.Files[s.RelPath] = s.Shasum
	}
	if bodyBuffer, err := json.Marshal(bodyJson); err != nil {
		panic(err)
	} else if req, err := http.NewRequest(http.MethodPost, resource, bytes.NewReader(bodyBuffer)); err != nil {
		panic(err)
	} else {
		req.Header.Add("Content-Type", "application/json")
		if res, err := client.Do(req); err != nil {
			panic(err)
		} else if res.StatusCode < 200 || res.StatusCode > 299 {
			panic(fmt.Sprintf("http error: status = %s, resource = %s\n", res.Status, res.Request.URL))
		} else if bodyBytes, err := io.ReadAll(res.Body); err != nil {
			panic(err)
		} else if err := json.Unmarshal(bodyBytes, &r); err != nil {
			panic(err)
		} else {
			log.Printf("files populated: %d files ", len(bodyJson.Files))
			return r, nil
		}
	}
}

func RestCreateVersion(client *http.Client, site string) (r VersionCreateReturn, e error) {
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
		return VersionCreateReturn{}, fmt.Errorf("error: RestCreateVersion: non-200 status code = %d ", resp.StatusCode)
	} else if body, err := io.ReadAll(resp.Body); err != nil {
		return VersionCreateReturn{}, err
	} else if err := json.Unmarshal(body, &r); err != nil {
		panic(err)
	}
	log.Printf("version created: id = %s", r.Name)
	return r, nil
}

func GetResource(client *http.Client, resource string) (string, error) {
	if resp, err := client.Get(resource); err != nil {
		return "", err
	} else {
		if body, err := io.ReadAll(resp.Body); err != nil {
			return "", err
		} else {
			return string(body), nil
		}
	}
}
